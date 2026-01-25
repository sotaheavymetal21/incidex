package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/notification"
	"incidex/internal/pkg/logger"
	"time"

	"go.uber.org/zap"
)

type IncidentUsecase interface {
	CreateIncident(ctx context.Context, creatorID uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error)
	GetAllIncidents(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error)
	GetIncidentByID(ctx context.Context, id uint) (*domain.Incident, error)
	UpdateIncident(ctx context.Context, userID uint, userRole domain.Role, id uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error)
	DeleteIncident(ctx context.Context, userRole domain.Role, id uint) error
	AssignIncident(ctx context.Context, userID uint, incidentID uint, assigneeID *uint) (*domain.Incident, error)
}

type incidentUsecase struct {
	incidentRepo        domain.IncidentRepository
	tagRepo             domain.TagRepository
	userRepo            domain.UserRepository
	activityRepo        domain.IncidentActivityRepository
	postMortemRepo      domain.PostMortemRepository
	notificationService *notification.NotificationService
	cacheRepo           domain.CacheRepository
}

func NewIncidentUsecase(incidentRepo domain.IncidentRepository, tagRepo domain.TagRepository, userRepo domain.UserRepository, activityRepo domain.IncidentActivityRepository, postMortemRepo domain.PostMortemRepository, notificationService *notification.NotificationService, cacheRepo domain.CacheRepository) IncidentUsecase {
	return &incidentUsecase{
		incidentRepo:        incidentRepo,
		tagRepo:             tagRepo,
		userRepo:            userRepo,
		activityRepo:        activityRepo,
		postMortemRepo:      postMortemRepo,
		notificationService: notificationService,
		cacheRepo:           cacheRepo,
	}
}

func (u *incidentUsecase) CreateIncident(ctx context.Context, creatorID uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error) {
	// 重要度をバリデーション
	if !isValidSeverity(severity) {
		return nil, domain.ErrValidation("invalid severity value")
	}

	// ステータスをバリデーション
	if !isValidStatus(status) {
		return nil, domain.ErrValidation("invalid status value")
	}

	// タグIDが指定されている場合はタグを取得（N+1を回避するためにバッチクエリ）
	var tags []domain.Tag
	if len(tagIDs) > 0 {
		var err error
		tags, err = u.tagRepo.FindByIDs(ctx, tagIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}
		if len(tags) != len(tagIDs) {
			return nil, domain.ErrValidation("one or more tags not found")
		}
	}

	// 重要度に基づいてデフォルトのSLAを設定
	slaHours := domain.GetDefaultSLAHours(severity)

	// インシデントを作成
	incident := &domain.Incident{
		Title:                    title,
		Description:              description,
		Severity:                 severity,
		Status:                   status,
		ImpactScope:              impactScope,
		DetectedAt:               detectedAt,
		AssigneeID:               assigneeID,
		CreatorID:                creatorID,
		Tags:                     tags,
		SLATargetResolutionHours: slaHours,
	}

	// SLA期限を計算して設定
	incident.SLADeadline = incident.CalculateSLADeadline()

	if err := u.incidentRepo.Create(ctx, incident); err != nil {
		return nil, err
	}

	// 作成アクティビティをログに記録
	activity := &domain.IncidentActivity{
		IncidentID:   incident.ID,
		UserID:       creatorID,
		ActivityType: domain.ActivityTypeCreated,
		CreatedAt:    time.Now(),
	}
	if err := u.activityRepo.Create(activity); err != nil {
		// errorをログに記録するが、インシデント作成は失敗させない
		logger.Log.Error("Failed to log creation activity", zap.Error(err))
	}

	// 通知を送信
	if u.notificationService != nil {
		creator, err := u.userRepo.FindByID(ctx, creatorID)
		if err == nil {
			if notifyErr := u.notificationService.NotifyIncidentCreated(incident, creator); notifyErr != nil {
				logger.Log.Error("Failed to send notification", zap.Error(notifyErr))
			}
		}
	}

	// キャッシュを無効化
	u.invalidateStatsCache(ctx)
	u.invalidateSearchCache(ctx)

	// 全リレーションを取得するためにリロード
	return u.incidentRepo.FindByID(ctx, incident.ID)
}

func (u *incidentUsecase) GetAllIncidents(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error) {
	// フィルターとページネーションからキャッシュキーを生成
	cacheKey, err := u.generateSearchCacheKey(filters, pagination)
	if err == nil {
		// キャッシュから取得を試みる
		if cachedData, cacheErr := u.cacheRepo.Get(ctx, cacheKey); cacheErr == nil {
			type CachedSearchResult struct {
				Incidents []*domain.Incident       `json:"incidents"`
				Result    *domain.PaginationResult `json:"result"`
			}
			var cached CachedSearchResult
			if unmarshalErr := json.Unmarshal([]byte(cachedData), &cached); unmarshalErr == nil {
				logger.Log.Debug("Cache hit for search", zap.String("cache_key", cacheKey))
				return cached.Incidents, cached.Result, nil
			}
		}
	}

	logger.Log.Debug("Cache miss for search, querying database")

	// データベースから取得
	incidents, result, err := u.incidentRepo.FindAll(ctx, filters, pagination)
	if err != nil {
		return nil, nil, err
	}

	// 結果を3分間キャッシュ
	if cacheKey != "" {
		type CachedSearchResult struct {
			Incidents []*domain.Incident       `json:"incidents"`
			Result    *domain.PaginationResult `json:"result"`
		}
		cached := CachedSearchResult{
			Incidents: incidents,
			Result:    result,
		}
		if cachedJSON, marshalErr := json.Marshal(cached); marshalErr == nil {
			if setErr := u.cacheRepo.Set(ctx, cacheKey, string(cachedJSON), 3*time.Minute); setErr != nil {
				logger.Log.Warn("Failed to cache search results", zap.Error(setErr))
			}
		}
	}

	return incidents, result, nil
}

func (u *incidentUsecase) GetIncidentByID(ctx context.Context, id uint) (*domain.Incident, error) {
	return u.incidentRepo.FindByID(ctx, id)
}

func (u *incidentUsecase) UpdateIncident(ctx context.Context, userID uint, userRole domain.Role, id uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error) {
	// 既存のインシデントを取得
	incident, err := u.incidentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 権限をチェック: Editorは自分のインシデントのみ編集可能、Adminは全て編集可能
	if userRole == domain.RoleEditor && incident.CreatorID != userID {
		return nil, domain.ErrForbidden("you can only edit your own incidents")
	}
	if userRole == domain.RoleViewer {
		return nil, domain.ErrForbidden("viewers cannot edit incidents")
	}

	// 重要度をバリデーション
	if !isValidSeverity(severity) {
		return nil, domain.ErrValidation("invalid severity value")
	}

	// ステータスをバリデーション
	if !isValidStatus(status) {
		return nil, domain.ErrValidation("invalid status value")
	}

	// ステータス変更に基づいてresolved_atを自動設定
	var resolvedAt *time.Time
	if status == domain.StatusResolved || status == domain.StatusClosed {
		// ステータスがresolved/closedに変更された場合、resolved_atを現在時刻に設定
		if incident.Status != domain.StatusResolved && incident.Status != domain.StatusClosed {
			now := time.Now()
			resolvedAt = &now
		} else {
			// 既にresolved/closedの場合は既存のresolved_atを維持
			resolvedAt = incident.ResolvedAt
		}
	} else {
		// ステータスがresolved/closedでない場合はresolved_atをクリア
		resolvedAt = nil
	}

	// タグIDが指定されている場合はタグを取得（N+1を回避するためにバッチクエリ）
	var tags []domain.Tag
	if len(tagIDs) > 0 {
		tags, err = u.tagRepo.FindByIDs(ctx, tagIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}
		if len(tags) != len(tagIDs) {
			return nil, domain.ErrValidation("one or more tags not found")
		}
	}

	// 変更を追跡してアクティビティをログに記録
	var activities []*domain.IncidentActivity

	// 重要度変更をチェック
	if incident.Severity != severity {
		activities = append(activities, &domain.IncidentActivity{
			IncidentID:   incident.ID,
			UserID:       userID,
			ActivityType: domain.ActivityTypeSeverityChange,
			OldValue:     string(incident.Severity),
			NewValue:     string(severity),
			CreatedAt:    time.Now(),
		})
	}

	// ステータス変更をチェック
	if incident.Status != status {
		activities = append(activities, &domain.IncidentActivity{
			IncidentID:   incident.ID,
			UserID:       userID,
			ActivityType: domain.ActivityTypeStatusChange,
			OldValue:     string(incident.Status),
			NewValue:     string(status),
			CreatedAt:    time.Now(),
		})

		// ステータスがresolvedに変更された場合、resolvedアクティビティをログに記録
		if status == domain.StatusResolved && incident.Status != domain.StatusResolved {
			activities = append(activities, &domain.IncidentActivity{
				IncidentID:   incident.ID,
				UserID:       userID,
				ActivityType: domain.ActivityTypeResolved,
				CreatedAt:    time.Now(),
			})
		}

		// ステータスがresolved/closedからopen/investigatingに変更された場合、reopenedアクティビティをログに記録
		if (incident.Status == domain.StatusResolved || incident.Status == domain.StatusClosed) &&
			(status == domain.StatusOpen || status == domain.StatusInvestigating) {
			activities = append(activities, &domain.IncidentActivity{
				IncidentID:   incident.ID,
				UserID:       userID,
				ActivityType: domain.ActivityTypeReopened,
				CreatedAt:    time.Now(),
			})
		}
	}

	// 通知用に古い値を保存（フィールド更新前に）
	oldSeverity := incident.Severity
	oldStatus := incident.Status

	// 担当者変更をチェック
	oldAssigneeID := incident.AssigneeID
	if (oldAssigneeID == nil && assigneeID != nil) ||
		(oldAssigneeID != nil && assigneeID == nil) ||
		(oldAssigneeID != nil && assigneeID != nil && *oldAssigneeID != *assigneeID) {

		var oldAssigneeName, newAssigneeName string
		if oldAssigneeID != nil {
			if oldAssignee, err := u.userRepo.FindByID(ctx, *oldAssigneeID); err == nil {
				oldAssigneeName = oldAssignee.Name
			}
		}
		if assigneeID != nil {
			if newAssignee, err := u.userRepo.FindByID(ctx, *assigneeID); err == nil {
				newAssigneeName = newAssignee.Name
			}
		}

		activities = append(activities, &domain.IncidentActivity{
			IncidentID:   incident.ID,
			UserID:       userID,
			ActivityType: domain.ActivityTypeAssigneeChange,
			OldValue:     oldAssigneeName,
			NewValue:     newAssigneeName,
			CreatedAt:    time.Now(),
		})
	}

	// インシデントフィールドを更新
	incident.Title = title
	incident.Description = description
	incident.Severity = severity
	incident.Status = status
	incident.ImpactScope = impactScope
	incident.DetectedAt = detectedAt
	incident.ResolvedAt = resolvedAt
	incident.AssigneeID = assigneeID
	incident.Tags = tags

	// 重要度が変更された場合はSLAを更新
	if oldSeverity != severity {
		incident.SLATargetResolutionHours = domain.GetDefaultSLAHours(severity)
		incident.SLADeadline = incident.CalculateSLADeadline()
	}

	// SLA違反ステータスをチェックして更新
	incident.SLAViolated = incident.CheckSLAViolation()

	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		return nil, err
	}

	// 全てのアクティビティを保存
	for _, activity := range activities {
		if err := u.activityRepo.Create(activity); err != nil {
			// errorをログに記録するが、更新は失敗させない
			logger.Log.Error("Failed to log activity", zap.Error(err))
		}
	}

	// 通知を送信
	if u.notificationService != nil {
		updater, _ := u.userRepo.FindByID(ctx, userID)

		// 担当者変更を通知
		if (oldAssigneeID == nil && assigneeID != nil) ||
			(oldAssigneeID != nil && assigneeID != nil && *oldAssigneeID != *assigneeID) {
			if assigneeID != nil {
				assignee, err := u.userRepo.FindByID(ctx, *assigneeID)
				if err == nil && updater != nil {
					if notifyErr := u.notificationService.NotifyAssigned(incident, assignee, updater); notifyErr != nil {
						logger.Log.Error("Failed to send assignee notification", zap.Error(notifyErr))
					}
				}
			}
		}

		// ステータス変更を通知
		if oldStatus != status {
			oldStatusStr := string(oldStatus)
			newStatusStr := string(status)
			if notifyErr := u.notificationService.NotifyStatusChange(incident, oldStatusStr, newStatusStr); notifyErr != nil {
				logger.Log.Error("Failed to send status change notification", zap.Error(notifyErr))
			}

			// 解決を通知
			if status == domain.StatusResolved && oldStatus != domain.StatusResolved && updater != nil {
				if notifyErr := u.notificationService.NotifyResolved(incident, updater); notifyErr != nil {
					logger.Log.Error("Failed to send resolved notification", zap.Error(notifyErr))
				}
			}
		}

		// 重要度変更を通知
		if oldSeverity != severity {
			oldSeverityStr := string(oldSeverity)
			newSeverityStr := string(severity)
			if notifyErr := u.notificationService.NotifySeverityChange(incident, oldSeverityStr, newSeverityStr); notifyErr != nil {
				logger.Log.Error("Failed to send severity change notification", zap.Error(notifyErr))
			}
		}
	}

	// キャッシュを無効化
	u.invalidateStatsCache(ctx)
	u.invalidateSearchCache(ctx)

	// 全リレーションを取得するためにリロード
	return u.incidentRepo.FindByID(ctx, incident.ID)
}

func (u *incidentUsecase) DeleteIncident(ctx context.Context, userRole domain.Role, id uint) error {
	// 管理者のみインシデントを削除可能
	if userRole != domain.RoleAdmin {
		return domain.ErrForbidden("only admins can delete incidents")
	}

	// インシデントが存在するかチェック
	if _, err := u.incidentRepo.FindByID(ctx, id); err != nil {
		return err
	}

	// キャッシュを無効化
	u.invalidateStatsCache(ctx)
	u.invalidateSearchCache(ctx)

	return u.incidentRepo.Delete(ctx, id)
}

func (u *incidentUsecase) AssignIncident(ctx context.Context, userID uint, incidentID uint, assigneeID *uint) (*domain.Incident, error) {
	// インシデントを取得
	incident, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrNotFound("incident")
	}

	// 担当者IDが指定されている場合、担当者が存在するかをバリデーション（後で使用するため結果をキャッシュ）
	var newAssignee *domain.User
	if assigneeID != nil {
		newAssignee, err = u.userRepo.FindByID(ctx, *assigneeID)
		if err != nil {
			logger.Log.Error("Failed to find assignee", zap.Uint("assignee_id", *assigneeID), zap.Error(err))
			return nil, domain.ErrInternal("Failed to validate assignee", err)
		}
		if newAssignee == nil {
			logger.Log.Warn("Assignee user not found", zap.Uint("assignee_id", *assigneeID))
			return nil, domain.ErrNotFound(fmt.Sprintf("User with ID %d", *assigneeID))
		}
	}

	// アクティビティログ用に古い担当者を保存し、古い担当者情報を一度取得
	var oldAssigneeID *uint
	var oldAssignee *domain.User
	if incident.AssigneeID != nil {
		oldAssigneeID = incident.AssigneeID
		oldAssignee, _ = u.userRepo.FindByID(ctx, *oldAssigneeID)
	}

	// 担当者を更新 - AssigneeIDフィールドのみを更新
	if err := u.incidentRepo.UpdateAssignee(ctx, incidentID, assigneeID); err != nil {
		return nil, err
	}

	// アクティビティログ用にユーザー情報を取得
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.Log.Error("Failed to find user", zap.Uint("user_id", userID), zap.Error(err))
		return nil, domain.ErrInternal("Failed to find user", err)
	}
	if user == nil {
		logger.Log.Error("User not found - invalid token", zap.Uint("user_id", userID))
		return nil, domain.ErrUnauthorized("Your session is invalid. Please log in again")
	}

	// 担当者名を準備（N+1を回避するためにキャッシュされたユーザーデータを使用）
	oldAssigneeName := "Unknown"
	if oldAssignee != nil {
		oldAssigneeName = oldAssignee.Name
	}
	newAssigneeName := "Unknown"
	if newAssignee != nil {
		newAssigneeName = newAssignee.Name
	}

	// 担当者変更のアクティビティログを作成
	var activityDescription string
	if assigneeID == nil {
		// 担当者を解除
		if oldAssigneeID != nil {
			activityDescription = fmt.Sprintf("%s が担当者を解除しました（以前の担当者: %s）", user.Name, oldAssigneeName)
		} else {
			activityDescription = fmt.Sprintf("%s が担当者を解除しました", user.Name)
		}
	} else {
		// 誰かに割り当て
		if oldAssigneeID == nil {
			activityDescription = fmt.Sprintf("%s が %s を担当者に割り当てました", user.Name, newAssigneeName)
		} else {
			activityDescription = fmt.Sprintf("%s が担当者を %s から %s に変更しました", user.Name, oldAssigneeName, newAssigneeName)
		}
	}

	// アクティビティログを作成
	activity := &domain.IncidentActivity{
		IncidentID:   incidentID,
		UserID:       userID,
		ActivityType: domain.ActivityTypeAssigneeChange,
		Comment:      activityDescription,
	}
	if err := u.activityRepo.Create(activity); err != nil {
		// errorをログに記録するが、操作は失敗させない
		logger.Log.Error("Failed to create activity log", zap.Error(err))
	}

	// リレーションを含めてインシデントをリロード
	reloadedIncident, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}

	// デバッグログ
	logger.Log.Info("AssignIncident response",
		zap.Uint("incident_id", incidentID),
		zap.Any("assignee_id", reloadedIncident.AssigneeID),
		zap.Any("assignee", reloadedIncident.Assignee),
	)

	return reloadedIncident, nil
}

// ヘルパー関数

func isValidSeverity(severity domain.Severity) bool {
	switch severity {
	case domain.SeverityCritical, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityLow:
		return true
	default:
		return false
	}
}

func isValidStatus(status domain.Status) bool {
	switch status {
	case domain.StatusOpen, domain.StatusInvestigating, domain.StatusResolved, domain.StatusClosed:
		return true
	default:
		return false
	}
}

// invalidateStatsCache は全ての統計キャッシュキーを無効化します
func (u *incidentUsecase) invalidateStatsCache(ctx context.Context) {
	// 全てのダッシュボード統計キャッシュを無効化
	patterns := []string{
		"stats:dashboard:*",
		"stats:sla",
		"stats:tags",
	}

	for _, pattern := range patterns {
		if err := u.cacheRepo.DeleteByPattern(ctx, pattern); err != nil {
			logger.Log.Warn("Failed to invalidate cache pattern", zap.String("pattern", pattern), zap.Error(err))
		}
	}
}

// invalidateSearchCache は全ての検索結果キャッシュを無効化します
func (u *incidentUsecase) invalidateSearchCache(ctx context.Context) {
	pattern := "search:incidents:*"
	if err := u.cacheRepo.DeleteByPattern(ctx, pattern); err != nil {
		logger.Log.Warn("Failed to invalidate search cache", zap.Error(err))
	}
}

// generateSearchCacheKey は検索結果用のキャッシュキーを生成します
func (u *incidentUsecase) generateSearchCacheKey(filters domain.IncidentFilters, pagination domain.Pagination) (string, error) {
	// キャッシュに関連する全フィールドを含むstructを作成
	type CacheKeyData struct {
		Filters    domain.IncidentFilters `json:"filters"`
		Pagination domain.Pagination      `json:"pagination"`
	}

	data := CacheKeyData{
		Filters:    filters,
		Pagination: pagination,
	}

	// JSONにマーシャル
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// SHA256 hashを生成
	hash := sha256.Sum256(jsonData)
	hashStr := hex.EncodeToString(hash[:])

	// キャッシュキーを返却
	return fmt.Sprintf("search:incidents:%s", hashStr), nil
}

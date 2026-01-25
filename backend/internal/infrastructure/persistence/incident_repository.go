package persistence

import (
	"context"
	"incidex/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type incidentRepository struct {
	db *gorm.DB
}

func NewIncidentRepository(db *gorm.DB) domain.IncidentRepository {
	return &incidentRepository{db: db}
}

func (r *incidentRepository) Create(ctx context.Context, incident *domain.Incident) error {
	return r.db.WithContext(ctx).Create(incident).Error
}

func (r *incidentRepository) FindAll(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error) {
	var incidents []*domain.Incident
	var total int64

	// クエリを構築します
	query := r.db.WithContext(ctx).Model(&domain.Incident{})

	// フィルタを適用します
	if filters.Severity != "" {
		query = query.Where("severity = ?", filters.Severity)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.AssignedToID != nil {
		query = query.Where("assignee_id = ?", *filters.AssignedToID)
	}
	if len(filters.TagIDs) > 0 {
		query = query.Joins("JOIN incident_tags ON incident_tags.incident_id = incidents.id").
			Where("incident_tags.tag_id IN ?", filters.TagIDs).
			Distinct()
	}
	if filters.Search != "" {
		// まず全文検索を試みます（search_vectorカラムが存在する場合）
		// tsquery用に検索クエリをフォーマットします
		searchTerms := strings.Fields(filters.Search)
		if len(searchTerms) > 0 {
			// 検索語からtsqueryを構築します
			tsquery := strings.Join(searchTerms, " & ")

			// search_vectorを使用した全文検索を試みます
			var testCount int64
			testQuery := r.db.WithContext(ctx).Model(&domain.Incident{}).
				Where("search_vector @@ to_tsquery('simple', ?)", tsquery)

			testErr := testQuery.Count(&testCount).Error
			if testErr == nil && testCount > 0 {
				// 全文検索が利用可能で結果が見つかった場合、それを使用します
				query = query.Where("search_vector @@ to_tsquery('simple', ?)", tsquery)
			} else {
				// LIKE検索にフォールバックします（日本語などの非英語テキスト、または全文検索で結果がない場合）
				searchPattern := "%" + strings.ToLower(filters.Search) + "%"
				query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(impact_scope) LIKE ?",
					searchPattern, searchPattern, searchPattern)
			}
		}
	}

	// 合計レコード数をカウントします
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// SQLインジェクション防止のためホワイトリストでバリデーションしてソートを適用します
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	// 許可されたソートカラムのホワイトリスト
	allowedSortColumns := map[string]bool{
		"id":          true,
		"title":       true,
		"severity":    true,
		"status":      true,
		"created_at":  true,
		"updated_at":  true,
		"detected_at": true,
		"resolved_at": true,
	}

	if !allowedSortColumns[sortBy] {
		sortBy = "created_at" // 安全なデフォルト値を使用します
	}

	order := strings.ToLower(filters.Order)
	if order != "asc" && order != "desc" {
		order = "desc" // 安全なデフォルト値を使用します
	}

	query = query.Order(sortBy + " " + order)

	// ページネーションを適用します
	if pagination.Limit == 0 {
		pagination.Limit = 20
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}
	offset := (pagination.Page - 1) * pagination.Limit
	query = query.Offset(offset).Limit(pagination.Limit)

	// リレーションをプリロードします
	query = query.Preload("Assignee").Preload("Creator").Preload("Tags")

	// クエリを実行します
	if err := query.Find(&incidents).Error; err != nil {
		return nil, nil, err
	}

	// 総ページ数を計算します
	totalPages := int(total) / pagination.Limit
	if int(total)%pagination.Limit > 0 {
		totalPages++
	}

	paginationResult := &domain.PaginationResult{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return incidents, paginationResult, nil
}

func (r *incidentRepository) FindByID(ctx context.Context, id uint) (*domain.Incident, error) {
	var incident domain.Incident
	if err := r.db.WithContext(ctx).
		Preload("Assignee").
		Preload("Creator").
		Preload("Tags").
		First(&incident, id).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

func (r *incidentRepository) Update(ctx context.Context, incident *domain.Incident) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(incident).Error
}

func (r *incidentRepository) UpdateAssignee(ctx context.Context, incidentID uint, assigneeID *uint) error {
	return r.db.WithContext(ctx).Model(&domain.Incident{}).Where("id = ?", incidentID).Update("assignee_id", assigneeID).Error
}

func (r *incidentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Incident{}, id).Error
}

// 統計メソッド

func (r *incidentRepository) Count(count *int64) error {
	return r.db.Model(&domain.Incident{}).Count(count).Error
}

func (r *incidentRepository) CountBySeverity(severity domain.Severity, count *int64) error {
	return r.db.Model(&domain.Incident{}).Where("severity = ?", severity).Count(count).Error
}

func (r *incidentRepository) CountByStatus(status domain.Status, count *int64) error {
	return r.db.Model(&domain.Incident{}).Where("status = ?", status).Count(count).Error
}

func (r *incidentRepository) FindRecent(limit int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	if err := r.db.
		Preload("Assignee").
		Preload("Creator").
		Preload("Tags").
		Order("detected_at DESC").
		Limit(limit).
		Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

func (r *incidentRepository) GetAllIncidents() ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	if err := r.db.Preload("Tags").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// CountSLAViolated はSLAに違反したインシデント数をカウントします
func (r *incidentRepository) CountSLAViolated(count *int64) error {
	return r.db.Model(&domain.Incident{}).Where("sla_violated = ?", true).Count(count).Error
}

// GetSLAMetrics はSLAパフォーマンスメトリクスを計算して返します
func (r *incidentRepository) GetSLAMetrics() (*domain.SLAMetrics, error) {
	var metrics domain.SLAMetrics

	// 総インシデント数
	if err := r.db.Model(&domain.Incident{}).Count(&metrics.TotalIncidents).Error; err != nil {
		return nil, err
	}

	// 解決済みインシデント数
	if err := r.db.Model(&domain.Incident{}).
		Where("status IN ?", []string{string(domain.StatusResolved), string(domain.StatusClosed)}).
		Count(&metrics.ResolvedIncidents).Error; err != nil {
		return nil, err
	}

	// SLA違反件数
	if err := r.CountSLAViolated(&metrics.SLAViolatedCount); err != nil {
		return nil, err
	}

	// SLA遵守率を計算します
	if metrics.ResolvedIncidents > 0 {
		compliantIncidents := metrics.ResolvedIncidents - metrics.SLAViolatedCount
		metrics.SLAComplianceRate = (float64(compliantIncidents) / float64(metrics.ResolvedIncidents)) * 100
	}

	// MTTR計算用にすべての解決済みインシデントを取得します
	var resolvedIncidents []*domain.Incident
	if err := r.db.Where("status IN ? AND resolved_at IS NOT NULL",
		[]string{string(domain.StatusResolved), string(domain.StatusClosed)}).
		Find(&resolvedIncidents).Error; err != nil {
		return nil, err
	}

	// MTTRを計算します
	if len(resolvedIncidents) > 0 {
		var totalResolutionTime float64
		var resolutionTimes []float64

		for _, incident := range resolvedIncidents {
			if resolutionTime := incident.GetResolutionTime(); resolutionTime != nil {
				hours := resolutionTime.Hours()
				totalResolutionTime += hours
				resolutionTimes = append(resolutionTimes, hours)
			}
		}

		// 平均MTTR
		if len(resolutionTimes) > 0 {
			metrics.AverageMTTR = totalResolutionTime / float64(len(resolutionTimes))
		}

		// 中央値MTTR（シンプルな中央値計算）
		if len(resolutionTimes) > 0 {
			// 中央値計算用に解決時間をソートします
			for i := 0; i < len(resolutionTimes); i++ {
				for j := i + 1; j < len(resolutionTimes); j++ {
					if resolutionTimes[i] > resolutionTimes[j] {
						resolutionTimes[i], resolutionTimes[j] = resolutionTimes[j], resolutionTimes[i]
					}
				}
			}

			mid := len(resolutionTimes) / 2
			if len(resolutionTimes)%2 == 0 {
				metrics.MedianMTTR = (resolutionTimes[mid-1] + resolutionTimes[mid]) / 2
			} else {
				metrics.MedianMTTR = resolutionTimes[mid]
			}
		}
	}

	// 現在期限超過のインシデントをカウントします（オープン状態でSLA期限を過ぎたもの）
	if err := r.db.Model(&domain.Incident{}).
		Where("status IN ? AND sla_deadline IS NOT NULL AND sla_deadline < ?",
			[]string{string(domain.StatusOpen), string(domain.StatusInvestigating)},
			gorm.Expr("NOW()")).
		Count(&metrics.CurrentlyOverdue).Error; err != nil {
		return nil, err
	}

	return &metrics, nil
}

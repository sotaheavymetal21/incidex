package usecase

import (
	"context"
	"encoding/json"
	"incidex/internal/domain"
	"time"
)

type PostMortemUsecase interface {
	CreatePostMortem(ctx context.Context, authorID uint, incidentID uint, rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned string, fiveWhys *domain.FiveWhysAnalysis) (*domain.PostMortem, error)
	GetPostMortemByID(ctx context.Context, id uint) (*domain.PostMortem, error)
	GetPostMortemByIncidentID(ctx context.Context, incidentID uint) (*domain.PostMortem, error)
	UpdatePostMortem(ctx context.Context, userID uint, userRole domain.Role, id uint, rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned string, fiveWhys *domain.FiveWhysAnalysis) (*domain.PostMortem, error)
	PublishPostMortem(ctx context.Context, userID uint, userRole domain.Role, id uint) (*domain.PostMortem, error)
	UnpublishPostMortem(ctx context.Context, userID uint, userRole domain.Role, id uint) (*domain.PostMortem, error)
	DeletePostMortem(ctx context.Context, userRole domain.Role, id uint) error
	GetAllPostMortems(ctx context.Context, filters domain.PostMortemFilters, pagination domain.Pagination) ([]*domain.PostMortem, *domain.PaginationResult, error)
}

type postMortemUsecase struct {
	postMortemRepo domain.PostMortemRepository
	incidentRepo   domain.IncidentRepository
	activityRepo   domain.IncidentActivityRepository
	userRepo       domain.UserRepository
}

func NewPostMortemUsecase(
	postMortemRepo domain.PostMortemRepository,
	incidentRepo domain.IncidentRepository,
	activityRepo domain.IncidentActivityRepository,
	userRepo domain.UserRepository,
) PostMortemUsecase {
	return &postMortemUsecase{
		postMortemRepo: postMortemRepo,
		incidentRepo:   incidentRepo,
		activityRepo:   activityRepo,
		userRepo:       userRepo,
	}
}

const maxFiveWhysFieldLength = 1000

// validateFiveWhys はFiveWhysAnalysisの各フィールドの長さをバリデーションします
func validateFiveWhys(fiveWhys *domain.FiveWhysAnalysis) error {
	if len(fiveWhys.Why1) > maxFiveWhysFieldLength {
		return domain.ErrValidation("Why1 must be at most 1000 characters")
	}
	if len(fiveWhys.Why2) > maxFiveWhysFieldLength {
		return domain.ErrValidation("Why2 must be at most 1000 characters")
	}
	if len(fiveWhys.Why3) > maxFiveWhysFieldLength {
		return domain.ErrValidation("Why3 must be at most 1000 characters")
	}
	if len(fiveWhys.Why4) > maxFiveWhysFieldLength {
		return domain.ErrValidation("Why4 must be at most 1000 characters")
	}
	if len(fiveWhys.Why5) > maxFiveWhysFieldLength {
		return domain.ErrValidation("Why5 must be at most 1000 characters")
	}
	return nil
}

func (u *postMortemUsecase) CreatePostMortem(
	ctx context.Context,
	authorID uint,
	incidentID uint,
	rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned string,
	fiveWhys *domain.FiveWhysAnalysis,
) (*domain.PostMortem, error) {
	// FiveWhysAnalysisのフィールド長をバリデーション
	if fiveWhys != nil {
		if err := validateFiveWhys(fiveWhys); err != nil {
			return nil, err
		}
	}

	// インシデントが存在するかチェック
	_, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, domain.ErrNotFound("incident")
	}

	// このインシデントにポストモーテムが既に存在するかチェック
	existingPM, _ := u.postMortemRepo.FindByIncidentID(ctx, incidentID)
	if existingPM != nil {
		return nil, domain.ErrConflict("Post-mortem already exists for this incident")
	}

	// Five Whys分析をJSONにマーシャル
	var fiveWhysJSON string
	if fiveWhys != nil {
		fiveWhysBytes, err := json.Marshal(fiveWhys)
		if err != nil {
			return nil, domain.ErrInternal("Failed to marshal five whys", err)
		}
		fiveWhysJSON = string(fiveWhysBytes)
	}

	// ポストモーテムを作成
	pm := &domain.PostMortem{
		IncidentID:       incidentID,
		AuthorID:         authorID,
		RootCause:        rootCause,
		ImpactAnalysis:   impactAnalysis,
		WhatWentWell:     whatWentWell,
		WhatWentWrong:    whatWentWrong,
		LessonsLearned:   lessonsLearned,
		FiveWhysAnalysis: fiveWhysJSON,
		Status:           domain.PMStatusDraft,
	}

	if err := u.postMortemRepo.Create(ctx, pm); err != nil {
		return nil, err
	}

	// リレーションを含めてリロード
	return u.postMortemRepo.FindByID(ctx, pm.ID)
}

func (u *postMortemUsecase) GetPostMortemByID(ctx context.Context, id uint) (*domain.PostMortem, error) {
	return u.postMortemRepo.FindByID(ctx, id)
}

func (u *postMortemUsecase) GetPostMortemByIncidentID(ctx context.Context, incidentID uint) (*domain.PostMortem, error) {
	return u.postMortemRepo.FindByIncidentID(ctx, incidentID)
}

func (u *postMortemUsecase) UpdatePostMortem(
	ctx context.Context,
	userID uint,
	userRole domain.Role,
	id uint,
	rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned string,
	fiveWhys *domain.FiveWhysAnalysis,
) (*domain.PostMortem, error) {
	// FiveWhysAnalysisのフィールド長をバリデーション
	if fiveWhys != nil {
		if err := validateFiveWhys(fiveWhys); err != nil {
			return nil, err
		}
	}

	// 既存のポストモーテムを取得
	pm, err := u.postMortemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 権限をチェック
	if userRole == domain.RoleEditor && pm.AuthorID != userID {
		return nil, domain.ErrForbidden("You can only update your own post-mortems")
	}

	// Five Whys分析をJSONにマーシャル
	var fiveWhysJSON string
	if fiveWhys != nil {
		fiveWhysBytes, err := json.Marshal(fiveWhys)
		if err != nil {
			return nil, domain.ErrInternal("Failed to marshal five whys", err)
		}
		fiveWhysJSON = string(fiveWhysBytes)
	}

	// フィールドを更新
	pm.RootCause = rootCause
	pm.ImpactAnalysis = impactAnalysis
	pm.WhatWentWell = whatWentWell
	pm.WhatWentWrong = whatWentWrong
	pm.LessonsLearned = lessonsLearned
	pm.FiveWhysAnalysis = fiveWhysJSON

	if err := u.postMortemRepo.Update(ctx, pm); err != nil {
		return nil, err
	}

	// リレーションを含めてリロード
	return u.postMortemRepo.FindByID(ctx, id)
}

func (u *postMortemUsecase) PublishPostMortem(
	ctx context.Context,
	userID uint,
	userRole domain.Role,
	id uint,
) (*domain.PostMortem, error) {
	// 既存のポストモーテムを取得
	pm, err := u.postMortemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 既に公開済みかチェック
	if pm.Status == domain.PMStatusPublished {
		return nil, domain.ErrValidation("Post-mortem is already published")
	}

	// 権限をチェック
	if userRole == domain.RoleEditor && pm.AuthorID != userID {
		return nil, domain.ErrForbidden("You can only publish your own post-mortems")
	}

	// ステータスを更新
	now := time.Now()
	pm.Status = domain.PMStatusPublished
	pm.PublishedAt = &now

	if err := u.postMortemRepo.Update(ctx, pm); err != nil {
		return nil, err
	}

	// リレーションを含めてリロード
	return u.postMortemRepo.FindByID(ctx, id)
}

func (u *postMortemUsecase) UnpublishPostMortem(
	ctx context.Context,
	userID uint,
	userRole domain.Role,
	id uint,
) (*domain.PostMortem, error) {
	// 既存のポストモーテムを取得
	pm, err := u.postMortemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 既にドラフトかチェック
	if pm.Status == domain.PMStatusDraft {
		return nil, domain.ErrValidation("Post-mortem is already in draft status")
	}

	// 権限をチェック（作者または管理者のみ非公開可能）
	if userRole == domain.RoleEditor && pm.AuthorID != userID {
		return nil, domain.ErrForbidden("You can only unpublish your own post-mortems")
	}

	// ステータスをドラフトに戻す
	pm.Status = domain.PMStatusDraft
	pm.PublishedAt = nil

	if err := u.postMortemRepo.Update(ctx, pm); err != nil {
		return nil, err
	}

	// リレーションを含めてリロード
	return u.postMortemRepo.FindByID(ctx, id)
}

func (u *postMortemUsecase) DeletePostMortem(ctx context.Context, userRole domain.Role, id uint) error {
	// 管理者のみ削除可能
	if userRole != domain.RoleAdmin {
		return domain.ErrForbidden("Only admin can delete post-mortems")
	}

	return u.postMortemRepo.Delete(ctx, id)
}

func (u *postMortemUsecase) GetAllPostMortems(
	ctx context.Context,
	filters domain.PostMortemFilters,
	pagination domain.Pagination,
) ([]*domain.PostMortem, *domain.PaginationResult, error) {
	return u.postMortemRepo.FindAll(ctx, filters, pagination)
}

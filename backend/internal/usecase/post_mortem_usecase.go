package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/ai"
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
	GenerateAIRootCauseSuggestion(ctx context.Context, incidentID uint) (string, error)
}

type postMortemUsecase struct {
	postMortemRepo domain.PostMortemRepository
	incidentRepo   domain.IncidentRepository
	activityRepo   domain.IncidentActivityRepository
	userRepo       domain.UserRepository
	aiService      *ai.OpenAIService
}

func NewPostMortemUsecase(
	postMortemRepo domain.PostMortemRepository,
	incidentRepo domain.IncidentRepository,
	activityRepo domain.IncidentActivityRepository,
	userRepo domain.UserRepository,
	aiService *ai.OpenAIService,
) PostMortemUsecase {
	return &postMortemUsecase{
		postMortemRepo: postMortemRepo,
		incidentRepo:   incidentRepo,
		activityRepo:   activityRepo,
		userRepo:       userRepo,
		aiService:      aiService,
	}
}

func (u *postMortemUsecase) CreatePostMortem(
	ctx context.Context,
	authorID uint,
	incidentID uint,
	rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned string,
	fiveWhys *domain.FiveWhysAnalysis,
) (*domain.PostMortem, error) {
	// Check if incident exists
	incident, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, domain.ErrNotFound("incident")
	}

	// Check if post-mortem already exists for this incident
	existingPM, _ := u.postMortemRepo.FindByIncidentID(ctx, incidentID)
	if existingPM != nil {
		return nil, domain.ErrConflict("Post-mortem already exists for this incident")
	}

	// Marshal Five Whys analysis to JSON
	var fiveWhysJSON string
	if fiveWhys != nil {
		fiveWhysBytes, err := json.Marshal(fiveWhys)
		if err != nil {
			return nil, domain.ErrInternal("Failed to marshal five whys", err)
		}
		fiveWhysJSON = string(fiveWhysBytes)
	}

	// Generate AI root cause suggestion
	var aiSuggestion string
	if u.aiService != nil {
		timeline, err := u.activityRepo.FindByIncidentID(incidentID, 100)
		if err == nil {
			suggestion, err := u.aiService.GeneratePostMortemRootCauseSuggestion(
				incident.Title,
				incident.Description,
				timeline,
			)
			if err != nil {
				// Log error but don't fail the creation
				fmt.Printf("Failed to generate AI root cause suggestion: %v\n", err)
			} else {
				aiSuggestion = suggestion
			}
		}
	}

	// Create post-mortem
	pm := &domain.PostMortem{
		IncidentID:            incidentID,
		AuthorID:              authorID,
		RootCause:             rootCause,
		ImpactAnalysis:        impactAnalysis,
		WhatWentWell:          whatWentWell,
		WhatWentWrong:         whatWentWrong,
		LessonsLearned:        lessonsLearned,
		FiveWhysAnalysis:      fiveWhysJSON,
		AIRootCauseSuggestion: aiSuggestion,
		Status:                domain.PMStatusDraft,
	}

	if err := u.postMortemRepo.Create(ctx, pm); err != nil {
		return nil, err
	}

	// Reload with relations
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
	// Get existing post-mortem
	pm, err := u.postMortemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if userRole == domain.RoleEditor && pm.AuthorID != userID {
		return nil, domain.ErrForbidden("You can only update your own post-mortems")
	}

	// Marshal Five Whys analysis to JSON
	var fiveWhysJSON string
	if fiveWhys != nil {
		fiveWhysBytes, err := json.Marshal(fiveWhys)
		if err != nil {
			return nil, domain.ErrInternal("Failed to marshal five whys", err)
		}
		fiveWhysJSON = string(fiveWhysBytes)
	}

	// Update fields
	pm.RootCause = rootCause
	pm.ImpactAnalysis = impactAnalysis
	pm.WhatWentWell = whatWentWell
	pm.WhatWentWrong = whatWentWrong
	pm.LessonsLearned = lessonsLearned
	pm.FiveWhysAnalysis = fiveWhysJSON

	if err := u.postMortemRepo.Update(ctx, pm); err != nil {
		return nil, err
	}

	// Reload with relations
	return u.postMortemRepo.FindByID(ctx, id)
}

func (u *postMortemUsecase) PublishPostMortem(
	ctx context.Context,
	userID uint,
	userRole domain.Role,
	id uint,
) (*domain.PostMortem, error) {
	// Get existing post-mortem
	pm, err := u.postMortemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if already published
	if pm.Status == domain.PMStatusPublished {
		return nil, domain.ErrValidation("Post-mortem is already published")
	}

	// Check permissions
	if userRole == domain.RoleEditor && pm.AuthorID != userID {
		return nil, domain.ErrForbidden("You can only publish your own post-mortems")
	}

	// Update status
	now := time.Now()
	pm.Status = domain.PMStatusPublished
	pm.PublishedAt = &now

	if err := u.postMortemRepo.Update(ctx, pm); err != nil {
		return nil, err
	}

	// Reload with relations
	return u.postMortemRepo.FindByID(ctx, id)
}

func (u *postMortemUsecase) UnpublishPostMortem(
	ctx context.Context,
	userID uint,
	userRole domain.Role,
	id uint,
) (*domain.PostMortem, error) {
	// Get existing post-mortem
	pm, err := u.postMortemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if already draft
	if pm.Status == domain.PMStatusDraft {
		return nil, domain.ErrValidation("Post-mortem is already in draft status")
	}

	// Check permissions (only author or admin can unpublish)
	if userRole == domain.RoleEditor && pm.AuthorID != userID {
		return nil, domain.ErrForbidden("You can only unpublish your own post-mortems")
	}

	// Update status back to draft
	pm.Status = domain.PMStatusDraft
	pm.PublishedAt = nil

	if err := u.postMortemRepo.Update(ctx, pm); err != nil {
		return nil, err
	}

	// Reload with relations
	return u.postMortemRepo.FindByID(ctx, id)
}

func (u *postMortemUsecase) DeletePostMortem(ctx context.Context, userRole domain.Role, id uint) error {
	// Only admin can delete
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

func (u *postMortemUsecase) GenerateAIRootCauseSuggestion(ctx context.Context, incidentID uint) (string, error) {
	if u.aiService == nil {
		return "", domain.ErrInternal("AI service is not configured", nil)
	}

	// Get incident
	incident, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return "", domain.ErrNotFound("incident")
	}

	// Get timeline
	timeline, err := u.activityRepo.FindByIncidentID(incidentID, 100)
	if err != nil {
		return "", domain.ErrInternal("Failed to get timeline", err)
	}

	// Generate suggestion
	suggestion, err := u.aiService.GeneratePostMortemRootCauseSuggestion(
		incident.Title,
		incident.Description,
		timeline,
	)
	if err != nil {
		return "", domain.ErrInternal("Failed to generate AI suggestion", err)
	}

	return suggestion, nil
}

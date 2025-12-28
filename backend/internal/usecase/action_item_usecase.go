package usecase

import (
	"context"
	"incidex/internal/domain"
	"time"
)

type ActionItemUsecase interface {
	CreateActionItem(ctx context.Context, postMortemID uint, title, description string, assigneeID *uint, priority domain.Priority, dueDate *time.Time, relatedLinks string) (*domain.ActionItem, error)
	GetActionItemByID(ctx context.Context, id uint) (*domain.ActionItem, error)
	GetActionItemsByPostMortemID(ctx context.Context, postMortemID uint) ([]*domain.ActionItem, error)
	UpdateActionItem(ctx context.Context, id uint, title, description string, assigneeID *uint, priority domain.Priority, status domain.ActionStatus, dueDate *time.Time, relatedLinks string) (*domain.ActionItem, error)
	DeleteActionItem(ctx context.Context, userRole domain.Role, id uint) error
	GetAllActionItems(ctx context.Context, filters domain.ActionItemFilters, pagination domain.Pagination) ([]*domain.ActionItem, *domain.PaginationResult, error)
}

type actionItemUsecase struct {
	actionItemRepo domain.ActionItemRepository
	postMortemRepo domain.PostMortemRepository
}

func NewActionItemUsecase(
	actionItemRepo domain.ActionItemRepository,
	postMortemRepo domain.PostMortemRepository,
) ActionItemUsecase {
	return &actionItemUsecase{
		actionItemRepo: actionItemRepo,
		postMortemRepo: postMortemRepo,
	}
}

func (u *actionItemUsecase) CreateActionItem(
	ctx context.Context,
	postMortemID uint,
	title, description string,
	assigneeID *uint,
	priority domain.Priority,
	dueDate *time.Time,
	relatedLinks string,
) (*domain.ActionItem, error) {
	// Check if post-mortem exists
	_, err := u.postMortemRepo.FindByID(ctx, postMortemID)
	if err != nil {
		return nil, domain.ErrNotFound("Post-mortem").WithError(err)
	}

	// Validate priority
	if priority != domain.PriorityHigh && priority != domain.PriorityMedium && priority != domain.PriorityLow {
		return nil, domain.ErrValidation("Invalid priority value")
	}

	// Create action item
	item := &domain.ActionItem{
		PostMortemID: postMortemID,
		Title:        title,
		Description:  description,
		AssigneeID:   assigneeID,
		Priority:     priority,
		Status:       domain.ActionStatusPending,
		DueDate:      dueDate,
		RelatedLinks: relatedLinks,
	}

	if err := u.actionItemRepo.Create(ctx, item); err != nil {
		return nil, domain.ErrDatabase("Failed to create action item", err)
	}

	// Reload with relations
	return u.actionItemRepo.FindByID(ctx, item.ID)
}

func (u *actionItemUsecase) GetActionItemByID(ctx context.Context, id uint) (*domain.ActionItem, error) {
	return u.actionItemRepo.FindByID(ctx, id)
}

func (u *actionItemUsecase) GetActionItemsByPostMortemID(ctx context.Context, postMortemID uint) ([]*domain.ActionItem, error) {
	return u.actionItemRepo.FindByPostMortemID(ctx, postMortemID)
}

func (u *actionItemUsecase) UpdateActionItem(
	ctx context.Context,
	id uint,
	title, description string,
	assigneeID *uint,
	priority domain.Priority,
	status domain.ActionStatus,
	dueDate *time.Time,
	relatedLinks string,
) (*domain.ActionItem, error) {
	// Get existing action item
	item, err := u.actionItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrNotFound("Action item").WithError(err)
	}

	// Validate priority
	if priority != domain.PriorityHigh && priority != domain.PriorityMedium && priority != domain.PriorityLow {
		return nil, domain.ErrValidation("Invalid priority value")
	}

	// Validate status
	if status != domain.ActionStatusPending && status != domain.ActionStatusInProgress && status != domain.ActionStatusCompleted {
		return nil, domain.ErrValidation("Invalid status value")
	}

	// Track old status
	oldStatus := item.Status

	// Update fields
	item.Title = title
	item.Description = description
	item.AssigneeID = assigneeID
	item.Priority = priority
	item.Status = status
	item.DueDate = dueDate
	item.RelatedLinks = relatedLinks

	// Set CompletedAt when status changes to completed
	if status == domain.ActionStatusCompleted && oldStatus != domain.ActionStatusCompleted {
		now := time.Now()
		item.CompletedAt = &now
	}

	// Clear CompletedAt if status changes from completed to something else
	if status != domain.ActionStatusCompleted && oldStatus == domain.ActionStatusCompleted {
		item.CompletedAt = nil
	}

	if err := u.actionItemRepo.Update(ctx, item); err != nil {
		return nil, domain.ErrDatabase("Failed to update action item", err)
	}

	// Reload with relations
	return u.actionItemRepo.FindByID(ctx, id)
}

func (u *actionItemUsecase) DeleteActionItem(ctx context.Context, userRole domain.Role, id uint) error {
	// Only admin can delete
	if userRole != domain.RoleAdmin {
		return domain.ErrForbidden("Only admin can delete action items")
	}

	err := u.actionItemRepo.Delete(ctx, id)
	if err != nil {
		return domain.ErrDatabase("Failed to delete action item", err)
	}

	return nil
}

func (u *actionItemUsecase) GetAllActionItems(
	ctx context.Context,
	filters domain.ActionItemFilters,
	pagination domain.Pagination,
) ([]*domain.ActionItem, *domain.PaginationResult, error) {
	return u.actionItemRepo.FindAll(ctx, filters, pagination)
}

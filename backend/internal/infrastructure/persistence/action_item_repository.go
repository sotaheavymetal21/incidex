package persistence

import (
	"context"
	"incidex/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type actionItemRepository struct {
	db *gorm.DB
}

func NewActionItemRepository(db *gorm.DB) domain.ActionItemRepository {
	return &actionItemRepository{db: db}
}

func (r *actionItemRepository) Create(ctx context.Context, item *domain.ActionItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *actionItemRepository) FindByID(ctx context.Context, id uint) (*domain.ActionItem, error) {
	var item domain.ActionItem
	if err := r.db.WithContext(ctx).
		Preload("PostMortem").
		Preload("Assignee").
		First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *actionItemRepository) FindByPostMortemID(ctx context.Context, postMortemID uint) ([]*domain.ActionItem, error) {
	var items []*domain.ActionItem
	if err := r.db.WithContext(ctx).
		Preload("Assignee").
		Where("post_mortem_id = ?", postMortemID).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *actionItemRepository) Update(ctx context.Context, item *domain.ActionItem) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(item).Error
}

func (r *actionItemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.ActionItem{}, id).Error
}

func (r *actionItemRepository) FindAll(ctx context.Context, filters domain.ActionItemFilters, pagination domain.Pagination) ([]*domain.ActionItem, *domain.PaginationResult, error) {
	var items []*domain.ActionItem
	var total int64

	// Build query
	query := r.db.WithContext(ctx).Model(&domain.ActionItem{})

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Priority != "" {
		query = query.Where("priority = ?", filters.Priority)
	}
	if filters.AssigneeID != 0 {
		query = query.Where("assignee_id = ?", filters.AssigneeID)
	}
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?",
			searchPattern, searchPattern)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	order := filters.Order
	if order == "" {
		order = "desc"
	}
	query = query.Order(sortBy + " " + order)

	// Apply pagination
	if pagination.Limit == 0 {
		pagination.Limit = 20
	}
	if pagination.Page == 0 {
		pagination.Page = 1
	}
	offset := (pagination.Page - 1) * pagination.Limit
	query = query.Offset(offset).Limit(pagination.Limit)

	// Preload relations
	query = query.Preload("PostMortem").Preload("Assignee")

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		return nil, nil, err
	}

	// Calculate total pages
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

	return items, paginationResult, nil
}

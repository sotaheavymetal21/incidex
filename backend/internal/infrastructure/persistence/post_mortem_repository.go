package persistence

import (
	"context"
	"incidex/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type postMortemRepository struct {
	db *gorm.DB
}

func NewPostMortemRepository(db *gorm.DB) domain.PostMortemRepository {
	return &postMortemRepository{db: db}
}

func (r *postMortemRepository) Create(ctx context.Context, pm *domain.PostMortem) error {
	return r.db.WithContext(ctx).Create(pm).Error
}

func (r *postMortemRepository) FindByID(ctx context.Context, id uint) (*domain.PostMortem, error) {
	var pm domain.PostMortem
	if err := r.db.WithContext(ctx).
		Preload("Incident").
		Preload("Author").
		Preload("ActionItems").
		Preload("ActionItems.Assignee").
		First(&pm, id).Error; err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *postMortemRepository) FindByIncidentID(ctx context.Context, incidentID uint) (*domain.PostMortem, error) {
	var pm domain.PostMortem
	if err := r.db.WithContext(ctx).
		Preload("Incident").
		Preload("Author").
		Preload("ActionItems").
		Preload("ActionItems.Assignee").
		Where("incident_id = ?", incidentID).
		First(&pm).Error; err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *postMortemRepository) Update(ctx context.Context, pm *domain.PostMortem) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(pm).Error
}

func (r *postMortemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.PostMortem{}, id).Error
}

func (r *postMortemRepository) FindAll(ctx context.Context, filters domain.PostMortemFilters, pagination domain.Pagination) ([]*domain.PostMortem, *domain.PaginationResult, error) {
	var postMortems []*domain.PostMortem
	var total int64

	// Build query
	query := r.db.WithContext(ctx).Model(&domain.PostMortem{})

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.AuthorID != 0 {
		query = query.Where("author_id = ?", filters.AuthorID)
	}
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(root_cause) LIKE ? OR LOWER(impact_analysis) LIKE ? OR LOWER(lessons_learned) LIKE ?",
			searchPattern, searchPattern, searchPattern)
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
	query = query.Preload("Incident").Preload("Author").Preload("ActionItems").Preload("ActionItems.Assignee")

	// Execute query
	if err := query.Find(&postMortems).Error; err != nil {
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

	return postMortems, paginationResult, nil
}

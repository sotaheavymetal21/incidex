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

	// Build query
	query := r.db.WithContext(ctx).Model(&domain.Incident{})

	// Apply filters
	if filters.Severity != "" {
		query = query.Where("severity = ?", filters.Severity)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if len(filters.TagIDs) > 0 {
		query = query.Joins("JOIN incident_tags ON incident_tags.incident_id = incidents.id").
			Where("incident_tags.tag_id IN ?", filters.TagIDs).
			Distinct()
	}
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
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
	query = query.Preload("Assignee").Preload("Creator").Preload("Tags")

	// Execute query
	if err := query.Find(&incidents).Error; err != nil {
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

func (r *incidentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Incident{}, id).Error
}

// Stats methods

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
	if err := r.db.Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

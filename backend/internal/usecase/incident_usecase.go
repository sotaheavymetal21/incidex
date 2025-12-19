package usecase

import (
	"context"
	"errors"
	"fmt"
	"incidex/internal/domain"
	"time"
)

type IncidentUsecase interface {
	CreateIncident(ctx context.Context, creatorID uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error)
	GetAllIncidents(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error)
	GetIncidentByID(ctx context.Context, id uint) (*domain.Incident, error)
	UpdateIncident(ctx context.Context, userID uint, userRole domain.Role, id uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, resolvedAt *time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error)
	DeleteIncident(ctx context.Context, userRole domain.Role, id uint) error
}

type incidentUsecase struct {
	incidentRepo domain.IncidentRepository
	tagRepo      domain.TagRepository
	userRepo     domain.UserRepository
}

func NewIncidentUsecase(incidentRepo domain.IncidentRepository, tagRepo domain.TagRepository, userRepo domain.UserRepository) IncidentUsecase {
	return &incidentUsecase{
		incidentRepo: incidentRepo,
		tagRepo:      tagRepo,
		userRepo:     userRepo,
	}
}

func (u *incidentUsecase) CreateIncident(ctx context.Context, creatorID uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error) {
	// Validate severity
	if !isValidSeverity(severity) {
		return nil, errors.New("invalid severity")
	}

	// Validate status
	if !isValidStatus(status) {
		return nil, errors.New("invalid status")
	}

	// Fetch tags if tag IDs are provided
	var tags []domain.Tag
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			tag, err := u.tagRepo.FindByID(tagID)
			if err != nil {
				return nil, fmt.Errorf("tag with ID %d not found", tagID)
			}
			tags = append(tags, *tag)
		}
	}

	// Create incident
	incident := &domain.Incident{
		Title:       title,
		Description: description,
		Summary:     "", // AI summary will be implemented in Phase 1C
		Severity:    severity,
		Status:      status,
		ImpactScope: impactScope,
		DetectedAt:  detectedAt,
		AssigneeID:  assigneeID,
		CreatorID:   creatorID,
		Tags:        tags,
	}

	if err := u.incidentRepo.Create(ctx, incident); err != nil {
		return nil, err
	}

	// Reload to get all relations
	return u.incidentRepo.FindByID(ctx, incident.ID)
}

func (u *incidentUsecase) GetAllIncidents(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error) {
	return u.incidentRepo.FindAll(ctx, filters, pagination)
}

func (u *incidentUsecase) GetIncidentByID(ctx context.Context, id uint) (*domain.Incident, error) {
	return u.incidentRepo.FindByID(ctx, id)
}

func (u *incidentUsecase) UpdateIncident(ctx context.Context, userID uint, userRole domain.Role, id uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, resolvedAt *time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error) {
	// Fetch existing incident
	incident, err := u.incidentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check permissions: Editor can only edit own incidents, Admin can edit all
	if userRole == domain.RoleEditor && incident.CreatorID != userID {
		return nil, errors.New("permission denied: you can only edit your own incidents")
	}
	if userRole == domain.RoleViewer {
		return nil, errors.New("permission denied: viewers cannot edit incidents")
	}

	// Validate severity
	if !isValidSeverity(severity) {
		return nil, errors.New("invalid severity")
	}

	// Validate status
	if !isValidStatus(status) {
		return nil, errors.New("invalid status")
	}

	// Validate resolved_at > detected_at
	if resolvedAt != nil && resolvedAt.Before(detectedAt) {
		return nil, errors.New("resolved_at must be after detected_at")
	}

	// Fetch tags if tag IDs are provided
	var tags []domain.Tag
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			tag, err := u.tagRepo.FindByID(tagID)
			if err != nil {
				return nil, fmt.Errorf("tag with ID %d not found", tagID)
			}
			tags = append(tags, *tag)
		}
	}

	// Update incident fields
	incident.Title = title
	incident.Description = description
	incident.Severity = severity
	incident.Status = status
	incident.ImpactScope = impactScope
	incident.DetectedAt = detectedAt
	incident.ResolvedAt = resolvedAt
	incident.AssigneeID = assigneeID
	incident.Tags = tags

	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		return nil, err
	}

	// Reload to get all relations
	return u.incidentRepo.FindByID(ctx, incident.ID)
}

func (u *incidentUsecase) DeleteIncident(ctx context.Context, userRole domain.Role, id uint) error {
	// Only admins can delete incidents
	if userRole != domain.RoleAdmin {
		return errors.New("permission denied: only admins can delete incidents")
	}

	// Check if incident exists
	if _, err := u.incidentRepo.FindByID(ctx, id); err != nil {
		return err
	}

	return u.incidentRepo.Delete(ctx, id)
}

// Helper functions

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

package usecase

import (
	"context"
	"fmt"
	"incidex/internal/domain"
	"time"
)

type IncidentTemplateUsecase struct {
	templateRepo domain.IncidentTemplateRepository
	tagRepo      domain.TagRepository
	incidentRepo domain.IncidentRepository
	userRepo     domain.UserRepository
}

func NewIncidentTemplateUsecase(
	templateRepo domain.IncidentTemplateRepository,
	tagRepo domain.TagRepository,
	incidentRepo domain.IncidentRepository,
	userRepo domain.UserRepository,
) *IncidentTemplateUsecase {
	return &IncidentTemplateUsecase{
		templateRepo: templateRepo,
		tagRepo:      tagRepo,
		incidentRepo: incidentRepo,
		userRepo:     userRepo,
	}
}

// CreateTemplate creates a new incident template
func (u *IncidentTemplateUsecase) CreateTemplate(ctx context.Context, userID uint, name, description, title, content string, severity domain.Severity, impactScope string, isPublic bool, tagIDs []uint) (*domain.IncidentTemplate, error) {
	// Validate severity
	if !isValidSeverity(severity) {
		return nil, domain.ErrValidation("invalid severity")
	}

	// Fetch tags
	var tags []domain.Tag
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			tag, err := u.tagRepo.FindByID(ctx, tagID)
			if err != nil {
				return nil, domain.ErrNotFound(fmt.Sprintf("Tag with ID %d", tagID))
			}
			tags = append(tags, *tag)
		}
	}

	template := &domain.IncidentTemplate{
		Name:        name,
		Description: description,
		Title:       title,
		Content:     content,
		Severity:    severity,
		ImpactScope: impactScope,
		CreatorID:   userID,
		IsPublic:    isPublic,
		Tags:        tags,
	}

	if err := u.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	// Reload to get all relations
	return u.templateRepo.FindByID(ctx, template.ID)
}

// GetAllTemplates retrieves all templates accessible by the user
func (u *IncidentTemplateUsecase) GetAllTemplates(ctx context.Context, userID uint) ([]*domain.IncidentTemplate, error) {
	return u.templateRepo.FindAll(ctx, userID)
}

// GetTemplateByID retrieves a template by ID
func (u *IncidentTemplateUsecase) GetTemplateByID(ctx context.Context, id uint) (*domain.IncidentTemplate, error) {
	return u.templateRepo.FindByID(ctx, id)
}

// UpdateTemplate updates an existing template
func (u *IncidentTemplateUsecase) UpdateTemplate(ctx context.Context, userID uint, userRole domain.Role, id uint, name, description, title, content string, severity domain.Severity, impactScope string, isPublic bool, tagIDs []uint) (*domain.IncidentTemplate, error) {
	// Fetch existing template
	template, err := u.templateRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check permissions: Only creator or admin can edit
	if userRole != domain.RoleAdmin && template.CreatorID != userID {
		return nil, domain.ErrForbidden("only template creator or admin can edit")
	}

	// Validate severity
	if !isValidSeverity(severity) {
		return nil, domain.ErrValidation("invalid severity")
	}

	// Fetch tags
	var tags []domain.Tag
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			tag, err := u.tagRepo.FindByID(ctx, tagID)
			if err != nil {
				return nil, domain.ErrNotFound(fmt.Sprintf("Tag with ID %d", tagID))
			}
			tags = append(tags, *tag)
		}
	}

	// Update template fields
	template.Name = name
	template.Description = description
	template.Title = title
	template.Content = content
	template.Severity = severity
	template.ImpactScope = impactScope
	template.IsPublic = isPublic
	template.Tags = tags

	if err := u.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	// Reload to get all relations
	return u.templateRepo.FindByID(ctx, template.ID)
}

// DeleteTemplate deletes a template
func (u *IncidentTemplateUsecase) DeleteTemplate(ctx context.Context, userID uint, userRole domain.Role, id uint) error {
	// Fetch existing template
	template, err := u.templateRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check permissions: Only creator or admin can delete
	if userRole != domain.RoleAdmin && template.CreatorID != userID {
		return domain.ErrForbidden("only template creator or admin can delete")
	}

	return u.templateRepo.Delete(ctx, id)
}

// CreateIncidentFromTemplate creates an incident from a template
func (u *IncidentTemplateUsecase) CreateIncidentFromTemplate(ctx context.Context, templateID uint, creatorID uint, assigneeID *uint, detectedAt time.Time) (*domain.Incident, error) {
	// Get template
	template, err := u.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Increment usage count
	if err := u.templateRepo.IncrementUsageCount(ctx, templateID); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to increment template usage count: %v\n", err)
	}

	// Set default SLA based on severity
	slaHours := domain.GetDefaultSLAHours(template.Severity)

	// Create incident from template
	incident := &domain.Incident{
		Title:                    template.Title,
		Description:              template.Content,
		Summary:                  "", // AI summary will be generated if configured
		Severity:                 template.Severity,
		Status:                   domain.StatusOpen,
		ImpactScope:              template.ImpactScope,
		DetectedAt:               detectedAt,
		AssigneeID:               assigneeID,
		CreatorID:                creatorID,
		Tags:                     template.Tags,
		SLATargetResolutionHours: slaHours,
	}

	// Calculate SLA deadline
	incident.SLADeadline = incident.CalculateSLADeadline()

	// Create incident
	if err := u.incidentRepo.Create(ctx, incident); err != nil {
		return nil, err
	}

	// Reload to get all relations
	return u.incidentRepo.FindByID(ctx, incident.ID)
}

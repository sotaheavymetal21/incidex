package usecase

import (
	"context"
	"errors"
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/ai"
	"incidex/internal/infrastructure/notification"
	"time"
)

type IncidentUsecase interface {
	CreateIncident(ctx context.Context, creatorID uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error)
	GetAllIncidents(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error)
	GetIncidentByID(ctx context.Context, id uint) (*domain.Incident, error)
	UpdateIncident(ctx context.Context, userID uint, userRole domain.Role, id uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, resolvedAt *time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error)
	DeleteIncident(ctx context.Context, userRole domain.Role, id uint) error
	RegenerateSummary(ctx context.Context, id uint) (string, error)
	AssignIncident(ctx context.Context, userID uint, incidentID uint, assigneeID *uint) (*domain.Incident, error)
}

type incidentUsecase struct {
	incidentRepo        domain.IncidentRepository
	tagRepo             domain.TagRepository
	userRepo            domain.UserRepository
	activityRepo        domain.IncidentActivityRepository
	notificationService *notification.NotificationService
	aiService           *ai.OpenAIService
}

func NewIncidentUsecase(incidentRepo domain.IncidentRepository, tagRepo domain.TagRepository, userRepo domain.UserRepository, activityRepo domain.IncidentActivityRepository, notificationService *notification.NotificationService, aiService *ai.OpenAIService) IncidentUsecase {
	return &incidentUsecase{
		incidentRepo:        incidentRepo,
		tagRepo:             tagRepo,
		userRepo:            userRepo,
		activityRepo:        activityRepo,
		notificationService: notificationService,
		aiService:           aiService,
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

	// Generate AI summary
	var summary string
	if u.aiService != nil {
		aiSummary, err := u.aiService.GenerateIncidentSummary(title, description, string(severity), impactScope)
		if err != nil {
			// Log error but don't fail the incident creation
			fmt.Printf("Failed to generate AI summary: %v\n", err)
		} else {
			summary = aiSummary
		}
	}

	// Set default SLA based on severity
	slaHours := domain.GetDefaultSLAHours(severity)

	// Create incident
	incident := &domain.Incident{
		Title:                    title,
		Description:              description,
		Summary:                  summary,
		Severity:                 severity,
		Status:                   status,
		ImpactScope:              impactScope,
		DetectedAt:               detectedAt,
		AssigneeID:               assigneeID,
		CreatorID:                creatorID,
		Tags:                     tags,
		SLATargetResolutionHours: slaHours,
	}

	// Calculate and set SLA deadline
	incident.SLADeadline = incident.CalculateSLADeadline()

	if err := u.incidentRepo.Create(ctx, incident); err != nil {
		return nil, err
	}

	// Log creation activity
	activity := &domain.IncidentActivity{
		IncidentID:   incident.ID,
		UserID:       creatorID,
		ActivityType: domain.ActivityTypeCreated,
		CreatedAt:    time.Now(),
	}
	if err := u.activityRepo.Create(activity); err != nil {
		// Log error but don't fail the incident creation
		fmt.Printf("Failed to log creation activity: %v\n", err)
	}

	// Send notification
	if u.notificationService != nil {
		creator, err := u.userRepo.FindByID(ctx, creatorID)
		if err == nil {
			if notifyErr := u.notificationService.NotifyIncidentCreated(incident, creator); notifyErr != nil {
				fmt.Printf("Failed to send notification: %v\n", notifyErr)
			}
		}
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

	// Track changes and log activities
	var activities []*domain.IncidentActivity

	// Check severity change
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

	// Check status change
	if incident.Status != status {
		activities = append(activities, &domain.IncidentActivity{
			IncidentID:   incident.ID,
			UserID:       userID,
			ActivityType: domain.ActivityTypeStatusChange,
			OldValue:     string(incident.Status),
			NewValue:     string(status),
			CreatedAt:    time.Now(),
		})

		// Log resolved activity if status changed to resolved
		if status == domain.StatusResolved && incident.Status != domain.StatusResolved {
			activities = append(activities, &domain.IncidentActivity{
				IncidentID:   incident.ID,
				UserID:       userID,
				ActivityType: domain.ActivityTypeResolved,
				CreatedAt:    time.Now(),
			})
		}

		// Log reopened activity if status changed from resolved/closed to open/investigating
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

	// Check assignee change
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

	// Update SLA if severity changed
	if incident.Severity != severity {
		incident.SLATargetResolutionHours = domain.GetDefaultSLAHours(severity)
		incident.SLADeadline = incident.CalculateSLADeadline()
	}

	// Check and update SLA violation status
	incident.SLAViolated = incident.CheckSLAViolation()

	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		return nil, err
	}

	// Save all activities
	for _, activity := range activities {
		if err := u.activityRepo.Create(activity); err != nil {
			// Log error but don't fail the update
			fmt.Printf("Failed to log activity: %v\n", err)
		}
	}

	// Send notifications
	if u.notificationService != nil {
		updater, _ := u.userRepo.FindByID(ctx, userID)

		// Notify assignee change
		if (oldAssigneeID == nil && assigneeID != nil) ||
			(oldAssigneeID != nil && assigneeID != nil && *oldAssigneeID != *assigneeID) {
			if assigneeID != nil {
				assignee, err := u.userRepo.FindByID(ctx, *assigneeID)
				if err == nil && updater != nil {
					if notifyErr := u.notificationService.NotifyAssigned(incident, assignee, updater); notifyErr != nil {
						fmt.Printf("Failed to send assignee notification: %v\n", notifyErr)
					}
				}
			}
		}

		// Notify status change
		if incident.Status != status {
			oldStatusStr := string(incident.Status)
			newStatusStr := string(status)
			if notifyErr := u.notificationService.NotifyStatusChange(incident, oldStatusStr, newStatusStr); notifyErr != nil {
				fmt.Printf("Failed to send status change notification: %v\n", notifyErr)
			}

			// Notify resolved
			if status == domain.StatusResolved && incident.Status != domain.StatusResolved && updater != nil {
				if notifyErr := u.notificationService.NotifyResolved(incident, updater); notifyErr != nil {
					fmt.Printf("Failed to send resolved notification: %v\n", notifyErr)
				}
			}
		}
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

func (u *incidentUsecase) RegenerateSummary(ctx context.Context, id uint) (string, error) {
	// Fetch incident
	incident, err := u.incidentRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	// Generate AI summary
	var summary string
	if u.aiService != nil {
		aiSummary, err := u.aiService.GenerateIncidentSummary(
			incident.Title,
			incident.Description,
			string(incident.Severity),
			incident.ImpactScope,
		)
		if err != nil {
			return "", fmt.Errorf("failed to generate summary: %w", err)
		}
		summary = aiSummary
	} else {
		return "", errors.New("AI service is not configured")
	}

	// Update incident summary
	incident.Summary = summary
	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		return "", err
	}

	return summary, nil
}

func (u *incidentUsecase) AssignIncident(ctx context.Context, userID uint, incidentID uint, assigneeID *uint) (*domain.Incident, error) {
	// Get the incident
	incident, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, errors.New("incident not found")
	}

	// Store old assignee for activity log
	var oldAssigneeID *uint
	if incident.AssigneeID != nil {
		oldAssigneeID = incident.AssigneeID
	}

	// Update assignee
	incident.AssigneeID = assigneeID
	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		return nil, err
	}

	// Get user info for activity log
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create activity log for assignment change
	var activityDescription string
	if assigneeID == nil {
		// Unassigned
		if oldAssigneeID != nil {
			oldAssignee, _ := u.userRepo.FindByID(ctx, *oldAssigneeID)
			oldAssigneeName := "Unknown"
			if oldAssignee != nil {
				oldAssigneeName = oldAssignee.Name
			}
			activityDescription = fmt.Sprintf("%s が担当者を解除しました（以前の担当者: %s）", user.Name, oldAssigneeName)
		} else {
			activityDescription = fmt.Sprintf("%s が担当者を解除しました", user.Name)
		}
	} else {
		// Assigned to someone
		newAssignee, _ := u.userRepo.FindByID(ctx, *assigneeID)
		newAssigneeName := "Unknown"
		if newAssignee != nil {
			newAssigneeName = newAssignee.Name
		}

		if oldAssigneeID == nil {
			activityDescription = fmt.Sprintf("%s が %s を担当者に割り当てました", user.Name, newAssigneeName)
		} else {
			oldAssignee, _ := u.userRepo.FindByID(ctx, *oldAssigneeID)
			oldAssigneeName := "Unknown"
			if oldAssignee != nil {
				oldAssigneeName = oldAssignee.Name
			}
			activityDescription = fmt.Sprintf("%s が担当者を %s から %s に変更しました", user.Name, oldAssigneeName, newAssigneeName)
		}
	}

	// Create activity log
	activity := &domain.IncidentActivity{
		IncidentID:   incidentID,
		UserID:       userID,
		ActivityType: domain.ActivityTypeAssigneeChange,
		Comment:      activityDescription,
	}
	if err := u.activityRepo.Create(activity); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create activity log: %v\n", err)
	}

	// Reload incident with relationships
	return u.incidentRepo.FindByID(ctx, incidentID)
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

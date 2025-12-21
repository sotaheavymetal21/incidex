package usecase

import (
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/notification"
	"time"
)

type IncidentActivityUsecase struct {
	activityRepo        domain.IncidentActivityRepository
	incidentRepo        domain.IncidentRepository
	userRepo            domain.UserRepository
	notificationService *notification.NotificationService
}

func NewIncidentActivityUsecase(
	activityRepo domain.IncidentActivityRepository,
	incidentRepo domain.IncidentRepository,
	userRepo domain.UserRepository,
	notificationService *notification.NotificationService,
) *IncidentActivityUsecase {
	return &IncidentActivityUsecase{
		activityRepo:        activityRepo,
		incidentRepo:        incidentRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

// AddComment adds a comment to an incident.
func (u *IncidentActivityUsecase) AddComment(incidentID uint, userID uint, comment string) error {
	activity := &domain.IncidentActivity{
		IncidentID:   incidentID,
		UserID:       userID,
		ActivityType: domain.ActivityTypeComment,
		Comment:      comment,
		CreatedAt:    time.Now(),
	}

	if err := u.activityRepo.Create(activity); err != nil {
		return err
	}

	// Send notification
	if u.notificationService != nil && u.incidentRepo != nil && u.userRepo != nil {
		incident, err := u.incidentRepo.FindByID(nil, incidentID)
		if err == nil {
			commenter, err := u.userRepo.FindByID(nil, userID)
			if err == nil {
				if notifyErr := u.notificationService.NotifyComment(incident, commenter, comment); notifyErr != nil {
					fmt.Printf("Failed to send comment notification: %v\n", notifyErr)
				}
			}
		}
	}

	return nil
}

// LogActivityChange logs a change to an incident (status, severity, assignee, etc.).
func (u *IncidentActivityUsecase) LogActivityChange(incidentID uint, userID uint, activityType domain.ActivityType, oldValue, newValue string) error {
	activity := &domain.IncidentActivity{
		IncidentID:   incidentID,
		UserID:       userID,
		ActivityType: activityType,
		OldValue:     oldValue,
		NewValue:     newValue,
		CreatedAt:    time.Now(),
	}

	return u.activityRepo.Create(activity)
}

// LogCreation logs the creation of an incident.
func (u *IncidentActivityUsecase) LogCreation(incidentID uint, userID uint) error {
	activity := &domain.IncidentActivity{
		IncidentID:   incidentID,
		UserID:       userID,
		ActivityType: domain.ActivityTypeCreated,
		CreatedAt:    time.Now(),
	}

	return u.activityRepo.Create(activity)
}

// GetActivities retrieves all activities for an incident.
func (u *IncidentActivityUsecase) GetActivities(incidentID uint, limit int) ([]*domain.IncidentActivity, error) {
	return u.activityRepo.FindByIncidentID(incidentID, limit)
}

// GetRecentActivities retrieves recent activities across all incidents.
func (u *IncidentActivityUsecase) GetRecentActivities(limit int) ([]*domain.IncidentActivity, error) {
	return u.activityRepo.FindRecent(limit)
}

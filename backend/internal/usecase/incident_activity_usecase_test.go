package usecase

import (
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestIncidentActivityUsecase(
	activityRepo *mocks.MockIncidentActivityRepository,
	incidentRepo *mocks.MockIncidentRepository,
	userRepo *mocks.MockUserRepository,
) *IncidentActivityUsecase {
	return NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
}

func TestIncidentActivityUsecase_AddComment(t *testing.T) {
	t.Parallel()

	t.Run("successfully adds comment", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.MatchedBy(func(activity *domain.IncidentActivity) bool {
			return activity.IncidentID == 1 &&
				activity.UserID == 1 &&
				activity.ActivityType == domain.ActivityTypeComment &&
				activity.Comment == "Test comment"
		})).Return(nil)

		err := usecase.AddComment(1, 1, "Test comment")

		require.NoError(t, err)
		activityRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).
			Return(domain.ErrDatabase("failed", nil))

		err := usecase.AddComment(1, 1, "Test comment")

		require.Error(t, err)
		activityRepo.AssertExpectations(t)
	})
}

func TestIncidentActivityUsecase_LogActivityChange(t *testing.T) {
	t.Parallel()

	t.Run("successfully logs status change", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.MatchedBy(func(activity *domain.IncidentActivity) bool {
			return activity.IncidentID == 1 &&
				activity.UserID == 1 &&
				activity.ActivityType == domain.ActivityTypeStatusChange &&
				activity.OldValue == "open" &&
				activity.NewValue == "investigating"
		})).Return(nil)

		err := usecase.LogActivityChange(1, 1, domain.ActivityTypeStatusChange, "open", "investigating")

		require.NoError(t, err)
		activityRepo.AssertExpectations(t)
	})

	t.Run("successfully logs severity change", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.MatchedBy(func(activity *domain.IncidentActivity) bool {
			return activity.ActivityType == domain.ActivityTypeSeverityChange &&
				activity.OldValue == "medium" &&
				activity.NewValue == "critical"
		})).Return(nil)

		err := usecase.LogActivityChange(1, 1, domain.ActivityTypeSeverityChange, "medium", "critical")

		require.NoError(t, err)
		activityRepo.AssertExpectations(t)
	})
}

func TestIncidentActivityUsecase_LogCreation(t *testing.T) {
	t.Parallel()

	t.Run("successfully logs incident creation", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.MatchedBy(func(activity *domain.IncidentActivity) bool {
			return activity.IncidentID == 1 &&
				activity.UserID == 1 &&
				activity.ActivityType == domain.ActivityTypeCreated
		})).Return(nil)

		err := usecase.LogCreation(1, 1)

		require.NoError(t, err)
		activityRepo.AssertExpectations(t)
	})
}

func TestIncidentActivityUsecase_GetActivities(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns activities", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		expectedActivities := []*domain.IncidentActivity{
			testutil.NewTestIncidentActivity(1, 1),
			testutil.NewTestIncidentActivity(1, 2),
		}

		activityRepo.On("FindByIncidentID", uint(1), 50).Return(expectedActivities, nil)

		activities, err := usecase.GetActivities(1, 50)

		require.NoError(t, err)
		assert.Len(t, activities, 2)

		activityRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no activities", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("FindByIncidentID", uint(1), 50).Return([]*domain.IncidentActivity{}, nil)

		activities, err := usecase.GetActivities(1, 50)

		require.NoError(t, err)
		assert.Empty(t, activities)

		activityRepo.AssertExpectations(t)
	})
}

func TestIncidentActivityUsecase_GetRecentActivities(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns recent activities", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		expectedActivities := []*domain.IncidentActivity{
			testutil.NewTestIncidentActivity(1, 1),
			testutil.NewTestIncidentActivity(2, 1),
		}

		activityRepo.On("FindRecent", 10).Return(expectedActivities, nil)

		activities, err := usecase.GetRecentActivities(10)

		require.NoError(t, err)
		assert.Len(t, activities, 2)

		activityRepo.AssertExpectations(t)
	})
}

func TestIncidentActivityUsecase_AddTimelineEvent(t *testing.T) {
	t.Parallel()

	t.Run("successfully adds timeline event", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		eventTime := time.Now()
		activityRepo.On("Create", mock.MatchedBy(func(activity *domain.IncidentActivity) bool {
			return activity.IncidentID == 1 &&
				activity.UserID == 1 &&
				activity.ActivityType == domain.ActivityTypeDetected &&
				activity.Comment == "Issue detected in production"
		})).Return(nil)

		activity, err := usecase.AddTimelineEvent(
			1, 1,
			domain.ActivityTypeDetected,
			eventTime,
			"Issue detected in production",
		)

		require.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, domain.ActivityTypeDetected, activity.ActivityType)

		activityRepo.AssertExpectations(t)
	})

	t.Run("validates event type - detected", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)

		_, err := usecase.AddTimelineEvent(1, 1, domain.ActivityTypeDetected, time.Now(), "Description")

		require.NoError(t, err)
	})

	t.Run("validates event type - investigation started", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)

		_, err := usecase.AddTimelineEvent(1, 1, domain.ActivityTypeInvestigationStarted, time.Now(), "Description")

		require.NoError(t, err)
	})

	t.Run("validates event type - root cause identified", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)

		_, err := usecase.AddTimelineEvent(1, 1, domain.ActivityTypeRootCauseIdentified, time.Now(), "Description")

		require.NoError(t, err)
	})

	t.Run("validates event type - mitigation", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)

		_, err := usecase.AddTimelineEvent(1, 1, domain.ActivityTypeMitigation, time.Now(), "Description")

		require.NoError(t, err)
	})

	t.Run("validates event type - resolved", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)

		_, err := usecase.AddTimelineEvent(1, 1, domain.ActivityTypeTimelineResolved, time.Now(), "Description")

		require.NoError(t, err)
	})

	t.Run("validates event type - other", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)

		_, err := usecase.AddTimelineEvent(1, 1, domain.ActivityTypeOther, time.Now(), "Description")

		require.NoError(t, err)
	})

	t.Run("fails with invalid event type", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activity, err := usecase.AddTimelineEvent(
			1, 1,
			domain.ActivityType("invalid_type"),
			time.Now(),
			"Description",
		)

		require.Error(t, err)
		assert.Nil(t, activity)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("fails with comment activity type", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestIncidentActivityUsecase(activityRepo, incidentRepo, userRepo)

		activity, err := usecase.AddTimelineEvent(
			1, 1,
			domain.ActivityTypeComment, // Not a valid timeline event type
			time.Now(),
			"Description",
		)

		require.Error(t, err)
		assert.Nil(t, activity)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})
}

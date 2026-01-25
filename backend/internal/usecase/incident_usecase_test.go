package usecase

import (
	"context"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestIncidentUsecase(
	incidentRepo *mocks.MockIncidentRepository,
	tagRepo *mocks.MockTagRepository,
	userRepo *mocks.MockUserRepository,
	activityRepo *mocks.MockIncidentActivityRepository,
	postMortemRepo *mocks.MockPostMortemRepository,
	cacheRepo *mocks.MockCacheRepository,
) IncidentUsecase {
	return NewIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, nil, cacheRepo)
}

func TestIncidentUsecase_CreateIncident(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful creation with SLA auto-set", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		createdIncident := testutil.NewTestIncident()

		incidentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Incident")).Return(nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		userRepo.On("FindByID", ctx, uint(1)).Return(testutil.NewTestUser(), nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(0)).Return(createdIncident, nil)

		incident, err := usecase.CreateIncident(
			ctx, 1, "Test Incident", "Description",
			domain.SeverityMedium, domain.StatusOpen, "Test scope",
			time.Now(), nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)

		incidentRepo.AssertExpectations(t)
		activityRepo.AssertExpectations(t)
	})

	t.Run("creation with tags", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		tags := []domain.Tag{
			*testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1 }),
			*testutil.NewTestTag(func(t *domain.Tag) { t.ID = 2 }),
		}
		createdIncident := testutil.NewTestIncident()

		tagRepo.On("FindByIDs", ctx, []uint{1, 2}).Return(tags, nil)
		incidentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Incident")).Return(nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		userRepo.On("FindByID", ctx, uint(1)).Return(testutil.NewTestUser(), nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(0)).Return(createdIncident, nil)

		incident, err := usecase.CreateIncident(
			ctx, 1, "Test Incident", "Description",
			domain.SeverityCritical, domain.StatusOpen, "Test scope",
			time.Now(), nil, []uint{1, 2},
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)

		tagRepo.AssertExpectations(t)
	})

	t.Run("fails with invalid severity", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		incident, err := usecase.CreateIncident(
			ctx, 1, "Test Incident", "Description",
			domain.Severity("invalid"), domain.StatusOpen, "Test scope",
			time.Now(), nil, nil,
		)

		require.Error(t, err)
		assert.Nil(t, incident)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("fails with invalid status", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		incident, err := usecase.CreateIncident(
			ctx, 1, "Test Incident", "Description",
			domain.SeverityMedium, domain.Status("invalid"), "Test scope",
			time.Now(), nil, nil,
		)

		require.Error(t, err)
		assert.Nil(t, incident)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("fails when tag not found", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		// Return only 1 tag when 2 were requested
		tags := []domain.Tag{*testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1 })}
		tagRepo.On("FindByIDs", ctx, []uint{1, 999}).Return(tags, nil)

		incident, err := usecase.CreateIncident(
			ctx, 1, "Test Incident", "Description",
			domain.SeverityMedium, domain.StatusOpen, "Test scope",
			time.Now(), nil, []uint{1, 999},
		)

		require.Error(t, err)
		assert.Nil(t, incident)
	})
}

func TestIncidentUsecase_UpdateIncident(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("admin can edit any incident", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.CreatorID = 2 // Different creator
			i.Severity = domain.SeverityHigh        // Same as update to avoid activity log
			i.Status = domain.StatusInvestigating   // Same as update to avoid activity log
		})
		updatedIncident := testutil.NewTestIncident()

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		incidentRepo.On("Update", ctx, mock.AnythingOfType("*domain.Incident")).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 100, domain.RoleAdmin, 1, // Admin user with ID 100 editing incident by creator 2
			"Updated Title", "Updated Description",
			domain.SeverityHigh, domain.StatusInvestigating, "Updated scope",
			time.Now(), nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)

		incidentRepo.AssertExpectations(t)
	})

	t.Run("editor can edit own incident", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.CreatorID = 5 // Same as the editor user
			i.Severity = domain.SeverityHigh       // Same as update value
			i.Status = domain.StatusInvestigating // Same as update value
		})
		updatedIncident := testutil.NewTestIncident()

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		incidentRepo.On("Update", ctx, mock.AnythingOfType("*domain.Incident")).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 5, domain.RoleEditor, 1, // Editor with ID 5 editing own incident
			"Updated Title", "Updated Description",
			domain.SeverityHigh, domain.StatusInvestigating, "Updated scope",
			time.Now(), nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)
	})

	t.Run("editor cannot edit others incident", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.CreatorID = 2 // Different from editor
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 5, domain.RoleEditor, 1, // Editor with ID 5 trying to edit incident by creator 2
			"Updated Title", "Updated Description",
			domain.SeverityHigh, domain.StatusInvestigating, "Updated scope",
			time.Now(), nil, nil,
		)

		require.Error(t, err)
		assert.Nil(t, incident)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})

	t.Run("viewer cannot edit incidents", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.CreatorID = 5 // Same as viewer, but still shouldn't be able to edit
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 5, domain.RoleViewer, 1,
			"Updated Title", "Updated Description",
			domain.SeverityHigh, domain.StatusInvestigating, "Updated scope",
			time.Now(), nil, nil,
		)

		require.Error(t, err)
		assert.Nil(t, incident)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})

	t.Run("sets resolvedAt when status changes to resolved", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.Status = domain.StatusOpen
			i.ResolvedAt = nil
		})
		updatedIncident := testutil.NewTestResolvedIncident()

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		incidentRepo.On("Update", ctx, mock.MatchedBy(func(i *domain.Incident) bool {
			return i.Status == domain.StatusResolved && i.ResolvedAt != nil
		})).Return(nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Updated Title", "Updated Description",
			domain.SeverityMedium, domain.StatusResolved, "Updated scope",
			time.Now(), nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)

		incidentRepo.AssertExpectations(t)
	})
}

func TestIncidentUsecase_DeleteIncident(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("admin can delete incidents", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident()

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil)
		incidentRepo.On("Delete", ctx, uint(1)).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)

		err := usecase.DeleteIncident(ctx, domain.RoleAdmin, 1)

		require.NoError(t, err)
		incidentRepo.AssertExpectations(t)
	})

	t.Run("editor cannot delete incidents", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		err := usecase.DeleteIncident(ctx, domain.RoleEditor, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})

	t.Run("viewer cannot delete incidents", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		err := usecase.DeleteIncident(ctx, domain.RoleViewer, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})
}

func TestIncidentUsecase_GetIncidentByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns incident when found", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		expectedIncident := testutil.NewTestIncident()
		incidentRepo.On("FindByID", ctx, uint(1)).Return(expectedIncident, nil)

		incident, err := usecase.GetIncidentByID(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, expectedIncident, incident)

		incidentRepo.AssertExpectations(t)
	})
}

func TestValidSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity domain.Severity
		want     bool
	}{
		{"critical is valid", domain.SeverityCritical, true},
		{"high is valid", domain.SeverityHigh, true},
		{"medium is valid", domain.SeverityMedium, true},
		{"low is valid", domain.SeverityLow, true},
		{"unknown is invalid", domain.Severity("unknown"), false},
		{"empty is invalid", domain.Severity(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isValidSeverity(tt.severity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status domain.Status
		want   bool
	}{
		{"open is valid", domain.StatusOpen, true},
		{"investigating is valid", domain.StatusInvestigating, true},
		{"resolved is valid", domain.StatusResolved, true},
		{"closed is valid", domain.StatusClosed, true},
		{"unknown is invalid", domain.Status("unknown"), false},
		{"empty is invalid", domain.Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isValidStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

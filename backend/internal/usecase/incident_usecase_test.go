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

	t.Run("creation with assignee", func(t *testing.T) {
		t.Parallel()

		testutil.InitTestLogger()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		assigneeID := uint(2)
		createdIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.AssigneeID = &assigneeID
		})

		incidentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Incident")).Return(nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		userRepo.On("FindByID", ctx, uint(1)).Return(testutil.NewTestUser(), nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(0)).Return(createdIncident, nil)

		incident, err := usecase.CreateIncident(
			ctx, 1, "Test Incident", "Description",
			domain.SeverityCritical, domain.StatusOpen, "Test scope",
			time.Now(), &assigneeID, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)
		assert.NotNil(t, incident.AssigneeID)
	})

	t.Run("fails when tag fetch returns error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		tagRepo.On("FindByIDs", ctx, []uint{1}).Return(nil, domain.ErrDatabase("db error", nil))

		incident, err := usecase.CreateIncident(
			ctx, 1, "Test Incident", "Description",
			domain.SeverityMedium, domain.StatusOpen, "Test scope",
			time.Now(), nil, []uint{1},
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
			i.CreatorID = 2                       // Different creator
			i.Severity = domain.SeverityHigh      // Same as update to avoid activity log
			i.Status = domain.StatusInvestigating // Same as update to avoid activity log
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
			i.CreatorID = 5                       // Same as the editor user
			i.Severity = domain.SeverityHigh      // Same as update value
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

	t.Run("updates SLA when severity changes", func(t *testing.T) {
		t.Parallel()

		testutil.InitTestLogger()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.Severity = domain.SeverityLow
			i.Status = domain.StatusOpen
			i.SLATargetResolutionHours = 168 // Low severity default
		})
		updatedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.Severity = domain.SeverityCritical
			i.SLATargetResolutionHours = 4 // Critical severity
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		incidentRepo.On("Update", ctx, mock.MatchedBy(func(i *domain.Incident) bool {
			// Verify SLA was updated for critical severity
			return i.Severity == domain.SeverityCritical && i.SLATargetResolutionHours == 4
		})).Return(nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.SeverityCritical, domain.StatusOpen, "Scope",
			time.Now(), nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)
	})

	t.Run("logs assignee change activity", func(t *testing.T) {
		t.Parallel()

		testutil.InitTestLogger()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		oldAssigneeID := uint(2)
		newAssigneeID := uint(3)
		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.AssigneeID = &oldAssigneeID
			i.Severity = domain.SeverityMedium
			i.Status = domain.StatusOpen
		})
		oldAssignee := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 2
			u.Name = "Old Assignee"
		})
		newAssignee := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 3
			u.Name = "New Assignee"
		})
		updatedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.AssigneeID = &newAssigneeID
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		userRepo.On("FindByID", ctx, uint(2)).Return(oldAssignee, nil)
		userRepo.On("FindByID", ctx, uint(3)).Return(newAssignee, nil)
		incidentRepo.On("Update", ctx, mock.AnythingOfType("*domain.Incident")).Return(nil)
		activityRepo.On("Create", mock.MatchedBy(func(a *domain.IncidentActivity) bool {
			return a.ActivityType == domain.ActivityTypeAssigneeChange &&
				a.OldValue == "Old Assignee" &&
				a.NewValue == "New Assignee"
		})).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.SeverityMedium, domain.StatusOpen, "Scope",
			time.Now(), &newAssigneeID, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)
	})

	t.Run("logs reopen activity when changing from resolved to open", func(t *testing.T) {
		t.Parallel()

		testutil.InitTestLogger()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		resolvedAt := time.Now().Add(-1 * time.Hour)
		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.Status = domain.StatusResolved
			i.ResolvedAt = &resolvedAt
			i.Severity = domain.SeverityMedium
		})
		reopenedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.Status = domain.StatusOpen
			i.ResolvedAt = nil
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		incidentRepo.On("Update", ctx, mock.MatchedBy(func(i *domain.Incident) bool {
			return i.Status == domain.StatusOpen && i.ResolvedAt == nil
		})).Return(nil)
		// Should create status change activity and reopen activity
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(reopenedIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.SeverityMedium, domain.StatusOpen, "Scope",
			time.Now(), nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)
	})

	t.Run("updates tags successfully", func(t *testing.T) {
		t.Parallel()

		testutil.InitTestLogger()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.Severity = domain.SeverityMedium
			i.Status = domain.StatusOpen
		})
		tags := []domain.Tag{
			*testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1; t.Name = "tag1" }),
			*testutil.NewTestTag(func(t *domain.Tag) { t.ID = 2; t.Name = "tag2" }),
		}
		updatedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.Tags = tags
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		tagRepo.On("FindByIDs", ctx, []uint{1, 2}).Return(tags, nil)
		incidentRepo.On("Update", ctx, mock.MatchedBy(func(i *domain.Incident) bool {
			return len(i.Tags) == 2
		})).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.SeverityMedium, domain.StatusOpen, "Scope",
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

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.Severity("invalid"), domain.StatusOpen, "Scope",
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

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.SeverityMedium, domain.Status("invalid"), "Scope",
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

		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil)
		// Return only 1 tag when 2 were requested
		tags := []domain.Tag{*testutil.NewTestTag(func(t *domain.Tag) { t.ID = 1 })}
		tagRepo.On("FindByIDs", ctx, []uint{1, 999}).Return(tags, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.SeverityMedium, domain.StatusOpen, "Scope",
			time.Now(), nil, []uint{1, 999},
		)

		require.Error(t, err)
		assert.Nil(t, incident)
	})

	t.Run("clears resolvedAt when reopening from closed", func(t *testing.T) {
		t.Parallel()

		testutil.InitTestLogger()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		resolvedAt := time.Now().Add(-1 * time.Hour)
		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.Status = domain.StatusClosed
			i.ResolvedAt = &resolvedAt
			i.Severity = domain.SeverityMedium
		})
		investigatingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.Status = domain.StatusInvestigating
			i.ResolvedAt = nil
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		incidentRepo.On("Update", ctx, mock.MatchedBy(func(i *domain.Incident) bool {
			return i.Status == domain.StatusInvestigating && i.ResolvedAt == nil
		})).Return(nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		cacheRepo.On("DeleteByPattern", ctx, mock.Anything).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(investigatingIncident, nil)

		incident, err := usecase.UpdateIncident(
			ctx, 1, domain.RoleAdmin, 1,
			"Title", "Description",
			domain.SeverityMedium, domain.StatusInvestigating, "Scope",
			time.Now(), nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, incident)
	})

	t.Run("preserves resolvedAt when already resolved", func(t *testing.T) {
		t.Parallel()

		testutil.InitTestLogger()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		originalResolvedAt := time.Now().Add(-1 * time.Hour)
		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.Status = domain.StatusResolved
			i.ResolvedAt = &originalResolvedAt
			i.Severity = domain.SeverityMedium
		})
		updatedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.Status = domain.StatusResolved
			i.ResolvedAt = &originalResolvedAt
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		incidentRepo.On("Update", ctx, mock.MatchedBy(func(i *domain.Incident) bool {
			// ResolvedAt should be preserved (same as original)
			return i.Status == domain.StatusResolved && i.ResolvedAt != nil
		})).Return(nil)
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
	})
}

func TestIncidentUsecase_DeleteIncident(t *testing.T) {
	t.Parallel()

	testutil.InitTestLogger()
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

	t.Run("fails when incident not found", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		incidentRepo.On("FindByID", ctx, uint(999)).Return(nil, domain.ErrNotFound("incident"))

		err := usecase.DeleteIncident(ctx, domain.RoleAdmin, 999)

		require.Error(t, err)
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

func TestIncidentUsecase_GetAllIncidents(t *testing.T) {
	t.Parallel()

	testutil.InitTestLogger()
	ctx := context.Background()

	t.Run("returns incidents without cache", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		expectedIncidents := []*domain.Incident{testutil.NewTestIncident()}
		expectedResult := &domain.PaginationResult{
			Page:       1,
			Limit:      10,
			Total:      1,
			TotalPages: 1,
		}

		filters := domain.IncidentFilters{}
		pagination := domain.Pagination{Page: 1, Limit: 10}

		// キャッシュミス
		cacheRepo.On("Get", ctx, mock.Anything).Return("", nil)
		incidentRepo.On("FindAll", ctx, filters, pagination).Return(expectedIncidents, expectedResult, nil)
		cacheRepo.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		incidents, result, err := usecase.GetAllIncidents(ctx, filters, pagination)

		require.NoError(t, err)
		assert.Equal(t, expectedIncidents, incidents)
		assert.Equal(t, expectedResult, result)

		incidentRepo.AssertExpectations(t)
	})

	t.Run("returns incidents with filters", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		expectedIncidents := []*domain.Incident{testutil.NewTestIncident()}
		expectedResult := &domain.PaginationResult{
			Page:       1,
			Limit:      10,
			Total:      1,
			TotalPages: 1,
		}

		filters := domain.IncidentFilters{
			Severity: string(domain.SeverityCritical),
			Status:   string(domain.StatusOpen),
		}
		pagination := domain.Pagination{Page: 1, Limit: 10}

		cacheRepo.On("Get", ctx, mock.Anything).Return("", nil)
		incidentRepo.On("FindAll", ctx, filters, pagination).Return(expectedIncidents, expectedResult, nil)
		cacheRepo.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		incidents, result, err := usecase.GetAllIncidents(ctx, filters, pagination)

		require.NoError(t, err)
		assert.Len(t, incidents, 1)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("returns cached incidents", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		// Prepare cached data
		cachedJSON := `{"incidents":[{"id":1,"title":"Cached Incident"}],"result":{"page":1,"limit":10,"total":1,"total_pages":1}}`

		filters := domain.IncidentFilters{}
		pagination := domain.Pagination{Page: 1, Limit: 10}

		cacheRepo.On("Get", ctx, mock.Anything).Return(cachedJSON, nil)

		incidents, result, err := usecase.GetAllIncidents(ctx, filters, pagination)

		require.NoError(t, err)
		assert.Len(t, incidents, 1)
		assert.NotNil(t, result)

		// FindAll should not be called when cache hit
		incidentRepo.AssertNotCalled(t, "FindAll")
	})

	t.Run("returns from db when cache set fails", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		expectedIncidents := []*domain.Incident{testutil.NewTestIncident()}
		expectedResult := &domain.PaginationResult{
			Page:       1,
			Limit:      10,
			Total:      1,
			TotalPages: 1,
		}

		filters := domain.IncidentFilters{}
		pagination := domain.Pagination{Page: 1, Limit: 10}

		cacheRepo.On("Get", ctx, mock.Anything).Return("", nil)
		incidentRepo.On("FindAll", ctx, filters, pagination).Return(expectedIncidents, expectedResult, nil)
		cacheRepo.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(domain.ErrDatabase("cache error", nil))

		incidents, result, err := usecase.GetAllIncidents(ctx, filters, pagination)

		require.NoError(t, err)
		assert.Len(t, incidents, 1)
		assert.Equal(t, expectedResult, result)
	})
}

func TestIncidentUsecase_AssignIncident(t *testing.T) {
	t.Parallel()

	testutil.InitTestLogger()
	ctx := context.Background()

	t.Run("assigns user to incident", func(t *testing.T) {
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
			i.AssigneeID = nil
		})
		assigneeID := uint(2)
		callerUser := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 1
			u.Name = "Caller"
		})
		assignee := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 2
			u.Name = "Assignee"
		})
		updatedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.AssigneeID = &assigneeID
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		userRepo.On("FindByID", ctx, uint(2)).Return(assignee, nil)
		incidentRepo.On("UpdateAssignee", ctx, uint(1), &assigneeID).Return(nil)
		userRepo.On("FindByID", ctx, uint(1)).Return(callerUser, nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.AssignIncident(ctx, 1, 1, &assigneeID)

		require.NoError(t, err)
		assert.NotNil(t, incident)

		incidentRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("unassigns user from incident", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		oldAssigneeID := uint(2)
		existingIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.AssigneeID = &oldAssigneeID
		})
		callerUser := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 1
			u.Name = "Caller"
		})
		oldAssignee := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 2
			u.Name = "Old Assignee"
		})
		updatedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 1
			i.AssigneeID = nil
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil).Once()
		userRepo.On("FindByID", ctx, uint(2)).Return(oldAssignee, nil)
		incidentRepo.On("UpdateAssignee", ctx, uint(1), (*uint)(nil)).Return(nil)
		userRepo.On("FindByID", ctx, uint(1)).Return(callerUser, nil)
		activityRepo.On("Create", mock.AnythingOfType("*domain.IncidentActivity")).Return(nil)
		incidentRepo.On("FindByID", ctx, uint(1)).Return(updatedIncident, nil)

		incident, err := usecase.AssignIncident(ctx, 1, 1, nil)

		require.NoError(t, err)
		assert.NotNil(t, incident)
	})

	t.Run("returns error when incident not found", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		tagRepo := mocks.NewMockTagRepository()
		userRepo := mocks.NewMockUserRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		cacheRepo := mocks.NewMockCacheRepository()

		usecase := createTestIncidentUsecase(incidentRepo, tagRepo, userRepo, activityRepo, postMortemRepo, cacheRepo)

		incidentRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		assigneeID := uint(2)
		incident, err := usecase.AssignIncident(ctx, 1, 999, &assigneeID)

		require.Error(t, err)
		assert.Nil(t, incident)
	})

	t.Run("returns error when assignee not found", func(t *testing.T) {
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
		})

		incidentRepo.On("FindByID", ctx, uint(1)).Return(existingIncident, nil)
		userRepo.On("FindByID", ctx, uint(999)).Return(nil, nil)

		assigneeID := uint(999)
		incident, err := usecase.AssignIncident(ctx, 1, 1, &assigneeID)

		require.Error(t, err)
		assert.Nil(t, incident)
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

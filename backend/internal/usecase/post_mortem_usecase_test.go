package usecase

import (
	"context"
	"errors"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestPostMortemUsecase(
	postMortemRepo *mocks.MockPostMortemRepository,
	incidentRepo *mocks.MockIncidentRepository,
	activityRepo *mocks.MockIncidentActivityRepository,
	userRepo *mocks.MockUserRepository,
) PostMortemUsecase {
	return NewPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)
}

func TestPostMortemUsecase_CreatePostMortem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully creates post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		incident := testutil.NewTestIncident()
		createdPM := testutil.NewTestPostMortem(1, 1)

		incidentRepo.On("FindByID", ctx, uint(1)).Return(incident, nil)
		postMortemRepo.On("FindByIncidentID", ctx, uint(1)).Return(nil, nil)
		postMortemRepo.On("Create", ctx, mock.AnythingOfType("*domain.PostMortem")).Return(nil)
		postMortemRepo.On("FindByID", ctx, uint(0)).Return(createdPM, nil)

		pm, err := usecase.CreatePostMortem(
			ctx, 1, 1,
			"Root cause", "Impact analysis", "What went well",
			"What went wrong", "Lessons learned", nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, pm)

		postMortemRepo.AssertExpectations(t)
		incidentRepo.AssertExpectations(t)
	})

	t.Run("fails when incident not found", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		incidentRepo.On("FindByID", ctx, uint(999)).Return(nil, errors.New("not found"))

		pm, err := usecase.CreatePostMortem(
			ctx, 1, 999,
			"Root cause", "Impact", "Well", "Wrong", "Lessons", nil,
		)

		require.Error(t, err)
		assert.Nil(t, pm)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("fails when post-mortem already exists for incident", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		incident := testutil.NewTestIncident()
		existingPM := testutil.NewTestPostMortem(1, 1)

		incidentRepo.On("FindByID", ctx, uint(1)).Return(incident, nil)
		postMortemRepo.On("FindByIncidentID", ctx, uint(1)).Return(existingPM, nil)

		pm, err := usecase.CreatePostMortem(
			ctx, 1, 1,
			"Root cause", "Impact", "Well", "Wrong", "Lessons", nil,
		)

		require.Error(t, err)
		assert.Nil(t, pm)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeConflict, domainErr.Code)
	})

	t.Run("validates five whys field length", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		// Create a string longer than 1000 characters
		longString := make([]byte, 1001)
		for i := range longString {
			longString[i] = 'a'
		}

		fiveWhys := &domain.FiveWhysAnalysis{
			Why1: string(longString),
		}

		pm, err := usecase.CreatePostMortem(
			ctx, 1, 1,
			"Root cause", "Impact", "Well", "Wrong", "Lessons", fiveWhys,
		)

		require.Error(t, err)
		assert.Nil(t, pm)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})
}

func TestPostMortemUsecase_UpdatePostMortem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("admin can update any post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 2) // Author ID is 2
		updatedPM := testutil.NewTestPostMortem(1, 2, func(pm *domain.PostMortem) {
			pm.RootCause = "Updated root cause"
		})

		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil).Once()
		postMortemRepo.On("Update", ctx, mock.AnythingOfType("*domain.PostMortem")).Return(nil)
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(updatedPM, nil)

		pm, err := usecase.UpdatePostMortem(
			ctx, 100, domain.RoleAdmin, 1, // Admin user ID 100
			"Updated root cause", "Impact", "Well", "Wrong", "Lessons", nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, pm)

		postMortemRepo.AssertExpectations(t)
	})

	t.Run("editor can update own post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 5) // Author ID is 5
		updatedPM := testutil.NewTestPostMortem(1, 5)

		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil).Once()
		postMortemRepo.On("Update", ctx, mock.AnythingOfType("*domain.PostMortem")).Return(nil)
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(updatedPM, nil)

		pm, err := usecase.UpdatePostMortem(
			ctx, 5, domain.RoleEditor, 1, // Editor with ID 5 updating own PM
			"Updated root cause", "Impact", "Well", "Wrong", "Lessons", nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, pm)

		postMortemRepo.AssertExpectations(t)
	})

	t.Run("editor cannot update others post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 2) // Author ID is 2
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil)

		pm, err := usecase.UpdatePostMortem(
			ctx, 5, domain.RoleEditor, 1, // Editor with ID 5 trying to update PM by author 2
			"Updated root cause", "Impact", "Well", "Wrong", "Lessons", nil,
		)

		require.Error(t, err)
		assert.Nil(t, pm)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})
}

func TestPostMortemUsecase_PublishPostMortem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully publishes post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusDraft
		})
		publishedPM := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusPublished
		})

		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil).Once()
		postMortemRepo.On("Update", ctx, mock.MatchedBy(func(pm *domain.PostMortem) bool {
			return pm.Status == domain.PMStatusPublished && pm.PublishedAt != nil
		})).Return(nil)
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(publishedPM, nil)

		pm, err := usecase.PublishPostMortem(ctx, 1, domain.RoleAdmin, 1)

		require.NoError(t, err)
		assert.NotNil(t, pm)
		assert.Equal(t, domain.PMStatusPublished, pm.Status)

		postMortemRepo.AssertExpectations(t)
	})

	t.Run("fails when already published", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusPublished
		})
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil)

		pm, err := usecase.PublishPostMortem(ctx, 1, domain.RoleAdmin, 1)

		require.Error(t, err)
		assert.Nil(t, pm)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("editor cannot publish others post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 2, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusDraft
		})
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil)

		pm, err := usecase.PublishPostMortem(ctx, 5, domain.RoleEditor, 1)

		require.Error(t, err)
		assert.Nil(t, pm)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})
}

func TestPostMortemUsecase_UnpublishPostMortem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully unpublishes post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusPublished
		})
		unpublishedPM := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusDraft
		})

		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil).Once()
		postMortemRepo.On("Update", ctx, mock.MatchedBy(func(pm *domain.PostMortem) bool {
			return pm.Status == domain.PMStatusDraft && pm.PublishedAt == nil
		})).Return(nil)
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(unpublishedPM, nil)

		pm, err := usecase.UnpublishPostMortem(ctx, 1, domain.RoleAdmin, 1)

		require.NoError(t, err)
		assert.NotNil(t, pm)
		assert.Equal(t, domain.PMStatusDraft, pm.Status)

		postMortemRepo.AssertExpectations(t)
	})

	t.Run("fails when already draft", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		existingPM := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusDraft
		})
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(existingPM, nil)

		pm, err := usecase.UnpublishPostMortem(ctx, 1, domain.RoleAdmin, 1)

		require.Error(t, err)
		assert.Nil(t, pm)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})
}

func TestPostMortemUsecase_DeletePostMortem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("admin can delete post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		postMortemRepo.On("Delete", ctx, uint(1)).Return(nil)

		err := usecase.DeletePostMortem(ctx, domain.RoleAdmin, 1)

		require.NoError(t, err)
		postMortemRepo.AssertExpectations(t)
	})

	t.Run("editor cannot delete post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		err := usecase.DeletePostMortem(ctx, domain.RoleEditor, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})

	t.Run("viewer cannot delete post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		err := usecase.DeletePostMortem(ctx, domain.RoleViewer, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})
}

func TestPostMortemUsecase_GetPostMortemByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns post-mortem", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		expectedPM := testutil.NewTestPostMortem(1, 1)
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(expectedPM, nil)

		pm, err := usecase.GetPostMortemByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, pm)
		assert.Equal(t, uint(1), pm.ID)

		postMortemRepo.AssertExpectations(t)
	})
}

func TestPostMortemUsecase_GetPostMortemByIncidentID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns post-mortem when found", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		expectedPM := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.ID = 5
			pm.IncidentID = 10
		})

		postMortemRepo.On("FindByIncidentID", ctx, uint(10)).Return(expectedPM, nil)

		pm, err := usecase.GetPostMortemByIncidentID(ctx, 10)

		require.NoError(t, err)
		assert.NotNil(t, pm)
		assert.Equal(t, uint(10), pm.IncidentID)

		postMortemRepo.AssertExpectations(t)
	})

	t.Run("returns nil when post-mortem not found", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		postMortemRepo.On("FindByIncidentID", ctx, uint(999)).Return(nil, nil)

		pm, err := usecase.GetPostMortemByIncidentID(ctx, 999)

		require.NoError(t, err)
		assert.Nil(t, pm)

		postMortemRepo.AssertExpectations(t)
	})
}

func TestPostMortemUsecase_GetAllPostMortems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns all post-mortems with pagination", func(t *testing.T) {
		t.Parallel()

		postMortemRepo := mocks.NewMockPostMortemRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		activityRepo := mocks.NewMockIncidentActivityRepository()
		userRepo := mocks.NewMockUserRepository()
		usecase := createTestPostMortemUsecase(postMortemRepo, incidentRepo, activityRepo, userRepo)

		expectedPMs := []*domain.PostMortem{
			testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) { pm.ID = 1 }),
			testutil.NewTestPostMortem(2, 2, func(pm *domain.PostMortem) { pm.ID = 2 }),
		}
		paginationResult := &domain.PaginationResult{
			Page:       1,
			Limit:      10,
			Total:      2,
			TotalPages: 1,
		}

		filters := domain.PostMortemFilters{}
		pagination := domain.Pagination{Page: 1, Limit: 10}

		postMortemRepo.On("FindAll", ctx, filters, pagination).Return(expectedPMs, paginationResult, nil)

		pms, result, err := usecase.GetAllPostMortems(ctx, filters, pagination)

		require.NoError(t, err)
		assert.Len(t, pms, 2)
		assert.Equal(t, int64(2), result.Total)

		postMortemRepo.AssertExpectations(t)
	})
}

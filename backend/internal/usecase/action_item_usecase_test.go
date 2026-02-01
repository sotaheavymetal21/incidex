package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestActionItemUsecase(
	actionItemRepo *mocks.MockActionItemRepository,
	postMortemRepo *mocks.MockPostMortemRepository,
) ActionItemUsecase {
	return NewActionItemUsecase(actionItemRepo, postMortemRepo)
}

func TestActionItemUsecase_CreateActionItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully creates action item", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		pm := testutil.NewTestPostMortem(1, 1)
		createdItem := testutil.NewTestActionItem(1)

		postMortemRepo.On("FindByID", ctx, uint(1)).Return(pm, nil)
		actionItemRepo.On("Create", ctx, mock.AnythingOfType("*domain.ActionItem")).Return(nil)
		actionItemRepo.On("FindByID", ctx, uint(0)).Return(createdItem, nil)

		dueDate := time.Now().Add(7 * 24 * time.Hour)
		item, err := usecase.CreateActionItem(
			ctx, 1, "Fix bug", "Description",
			nil, domain.PriorityHigh, &dueDate, "[]",
		)

		require.NoError(t, err)
		assert.NotNil(t, item)

		postMortemRepo.AssertExpectations(t)
		actionItemRepo.AssertExpectations(t)
	})

	t.Run("fails when post-mortem not found", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		postMortemRepo.On("FindByID", ctx, uint(999)).Return(nil, errors.New("not found"))

		item, err := usecase.CreateActionItem(
			ctx, 999, "Fix bug", "Description",
			nil, domain.PriorityHigh, nil, "[]",
		)

		require.Error(t, err)
		assert.Nil(t, item)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("fails with invalid priority", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		pm := testutil.NewTestPostMortem(1, 1)
		postMortemRepo.On("FindByID", ctx, uint(1)).Return(pm, nil)

		item, err := usecase.CreateActionItem(
			ctx, 1, "Fix bug", "Description",
			nil, domain.Priority("invalid"), nil, "[]",
		)

		require.Error(t, err)
		assert.Nil(t, item)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("creates with assignee", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		pm := testutil.NewTestPostMortem(1, 1)
		assigneeID := uint(5)
		createdItem := testutil.NewTestActionItem(1, func(i *domain.ActionItem) {
			i.AssigneeID = &assigneeID
		})

		postMortemRepo.On("FindByID", ctx, uint(1)).Return(pm, nil)
		actionItemRepo.On("Create", ctx, mock.MatchedBy(func(item *domain.ActionItem) bool {
			return item.AssigneeID != nil && *item.AssigneeID == assigneeID
		})).Return(nil)
		actionItemRepo.On("FindByID", ctx, uint(0)).Return(createdItem, nil)

		item, err := usecase.CreateActionItem(
			ctx, 1, "Fix bug", "Description",
			&assigneeID, domain.PriorityMedium, nil, "[]",
		)

		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, &assigneeID, item.AssigneeID)

		actionItemRepo.AssertExpectations(t)
	})

	t.Run("fails when create returns error", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		pm := testutil.NewTestPostMortem(1, 1)

		postMortemRepo.On("FindByID", ctx, uint(1)).Return(pm, nil)
		actionItemRepo.On("Create", ctx, mock.AnythingOfType("*domain.ActionItem")).Return(errors.New("db error"))

		item, err := usecase.CreateActionItem(
			ctx, 1, "Fix bug", "Description",
			nil, domain.PriorityHigh, nil, "[]",
		)

		require.Error(t, err)
		assert.Nil(t, item)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})
}

func TestActionItemUsecase_UpdateActionItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully updates action item", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		existingItem := testutil.NewTestActionItem(1)
		updatedItem := testutil.NewTestActionItem(1, func(i *domain.ActionItem) {
			i.Title = "Updated Title"
			i.Status = domain.ActionStatusInProgress
		})

		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil).Once()
		actionItemRepo.On("Update", ctx, mock.AnythingOfType("*domain.ActionItem")).Return(nil)
		actionItemRepo.On("FindByID", ctx, uint(1)).Return(updatedItem, nil)

		item, err := usecase.UpdateActionItem(
			ctx, 1, "Updated Title", "Updated Description",
			nil, domain.PriorityMedium, domain.ActionStatusInProgress, nil, "[]",
		)

		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, "Updated Title", item.Title)
		assert.Equal(t, domain.ActionStatusInProgress, item.Status)

		actionItemRepo.AssertExpectations(t)
	})

	t.Run("sets CompletedAt when status changes to completed", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		existingItem := testutil.NewTestActionItem(1, func(i *domain.ActionItem) {
			i.Status = domain.ActionStatusInProgress
		})

		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil).Once()
		actionItemRepo.On("Update", ctx, mock.MatchedBy(func(item *domain.ActionItem) bool {
			return item.Status == domain.ActionStatusCompleted && item.CompletedAt != nil
		})).Return(nil)
		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil)

		_, err := usecase.UpdateActionItem(
			ctx, 1, "Title", "Description",
			nil, domain.PriorityMedium, domain.ActionStatusCompleted, nil, "[]",
		)

		require.NoError(t, err)
		actionItemRepo.AssertExpectations(t)
	})

	t.Run("clears CompletedAt when status changes from completed", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		completedAt := time.Now()
		existingItem := testutil.NewTestActionItem(1, func(i *domain.ActionItem) {
			i.Status = domain.ActionStatusCompleted
			i.CompletedAt = &completedAt
		})

		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil).Once()
		actionItemRepo.On("Update", ctx, mock.MatchedBy(func(item *domain.ActionItem) bool {
			return item.Status == domain.ActionStatusInProgress && item.CompletedAt == nil
		})).Return(nil)
		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil)

		_, err := usecase.UpdateActionItem(
			ctx, 1, "Title", "Description",
			nil, domain.PriorityMedium, domain.ActionStatusInProgress, nil, "[]",
		)

		require.NoError(t, err)
		actionItemRepo.AssertExpectations(t)
	})

	t.Run("fails with invalid status", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		existingItem := testutil.NewTestActionItem(1)
		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil)

		item, err := usecase.UpdateActionItem(
			ctx, 1, "Title", "Description",
			nil, domain.PriorityMedium, domain.ActionStatus("invalid"), nil, "[]",
		)

		require.Error(t, err)
		assert.Nil(t, item)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("fails when action item not found", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		actionItemRepo.On("FindByID", ctx, uint(999)).Return(nil, errors.New("not found"))

		item, err := usecase.UpdateActionItem(
			ctx, 999, "Title", "Description",
			nil, domain.PriorityMedium, domain.ActionStatusPending, nil, "[]",
		)

		require.Error(t, err)
		assert.Nil(t, item)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeNotFound, domainErr.Code)
	})

	t.Run("fails when update returns error", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		existingItem := testutil.NewTestActionItem(1)

		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil)
		actionItemRepo.On("Update", ctx, mock.AnythingOfType("*domain.ActionItem")).Return(errors.New("db error"))

		item, err := usecase.UpdateActionItem(
			ctx, 1, "Title", "Description",
			nil, domain.PriorityMedium, domain.ActionStatusPending, nil, "[]",
		)

		require.Error(t, err)
		assert.Nil(t, item)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})

	t.Run("fails with invalid priority", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		existingItem := testutil.NewTestActionItem(1)
		actionItemRepo.On("FindByID", ctx, uint(1)).Return(existingItem, nil)

		item, err := usecase.UpdateActionItem(
			ctx, 1, "Title", "Description",
			nil, domain.Priority("invalid"), domain.ActionStatusPending, nil, "[]",
		)

		require.Error(t, err)
		assert.Nil(t, item)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})
}

func TestActionItemUsecase_DeleteActionItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("admin can delete action item", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		actionItemRepo.On("Delete", ctx, uint(1)).Return(nil)

		err := usecase.DeleteActionItem(ctx, domain.RoleAdmin, 1)

		require.NoError(t, err)
		actionItemRepo.AssertExpectations(t)
	})

	t.Run("editor cannot delete action item", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		err := usecase.DeleteActionItem(ctx, domain.RoleEditor, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})

	t.Run("viewer cannot delete action item", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		err := usecase.DeleteActionItem(ctx, domain.RoleViewer, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})

	t.Run("fails when delete returns error", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		actionItemRepo.On("Delete", ctx, uint(1)).Return(errors.New("db error"))

		err := usecase.DeleteActionItem(ctx, domain.RoleAdmin, 1)

		require.Error(t, err)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})
}

func TestActionItemUsecase_GetActionItemByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns action item", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		expectedItem := testutil.NewTestActionItem(1)
		actionItemRepo.On("FindByID", ctx, uint(1)).Return(expectedItem, nil)

		item, err := usecase.GetActionItemByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, uint(1), item.ID)

		actionItemRepo.AssertExpectations(t)
	})
}

func TestActionItemUsecase_GetActionItemsByPostMortemID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns action items by post-mortem ID", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		expectedItems := []*domain.ActionItem{
			testutil.NewTestActionItem(1, func(i *domain.ActionItem) { i.ID = 1 }),
			testutil.NewTestActionItem(1, func(i *domain.ActionItem) { i.ID = 2 }),
		}
		actionItemRepo.On("FindByPostMortemID", ctx, uint(1)).Return(expectedItems, nil)

		items, err := usecase.GetActionItemsByPostMortemID(ctx, 1)

		require.NoError(t, err)
		assert.Len(t, items, 2)

		actionItemRepo.AssertExpectations(t)
	})
}

func TestActionItemUsecase_GetAllActionItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully returns all action items with pagination", func(t *testing.T) {
		t.Parallel()

		actionItemRepo := mocks.NewMockActionItemRepository()
		postMortemRepo := mocks.NewMockPostMortemRepository()
		usecase := createTestActionItemUsecase(actionItemRepo, postMortemRepo)

		expectedItems := []*domain.ActionItem{
			testutil.NewTestActionItem(1, func(i *domain.ActionItem) { i.ID = 1 }),
			testutil.NewTestActionItem(1, func(i *domain.ActionItem) { i.ID = 2 }),
		}
		paginationResult := &domain.PaginationResult{
			Page:       1,
			Limit:      10,
			Total:      2,
			TotalPages: 1,
		}

		filters := domain.ActionItemFilters{Status: "pending"}
		pagination := domain.Pagination{Page: 1, Limit: 10}

		actionItemRepo.On("FindAll", ctx, filters, pagination).Return(expectedItems, paginationResult, nil)

		items, result, err := usecase.GetAllActionItems(ctx, filters, pagination)

		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, int64(2), result.Total)

		actionItemRepo.AssertExpectations(t)
	})
}

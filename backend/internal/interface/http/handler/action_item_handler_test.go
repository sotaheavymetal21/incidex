package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockActionItemUsecase は usecase.ActionItemUsecase のモック実装です
type MockActionItemUsecase struct {
	mock.Mock
}

func NewMockActionItemUsecase() *MockActionItemUsecase {
	return &MockActionItemUsecase{}
}

func (m *MockActionItemUsecase) CreateActionItem(ctx context.Context, postMortemID uint, title, description string, assigneeID *uint, priority domain.Priority, dueDate *time.Time, relatedLinks string) (*domain.ActionItem, error) {
	args := m.Called(ctx, postMortemID, title, description, assigneeID, priority, dueDate, relatedLinks)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ActionItem), args.Error(1)
}

func (m *MockActionItemUsecase) GetActionItemByID(ctx context.Context, id uint) (*domain.ActionItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ActionItem), args.Error(1)
}

func (m *MockActionItemUsecase) GetActionItemsByPostMortemID(ctx context.Context, postMortemID uint) ([]*domain.ActionItem, error) {
	args := m.Called(ctx, postMortemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ActionItem), args.Error(1)
}

func (m *MockActionItemUsecase) UpdateActionItem(ctx context.Context, id uint, title, description string, assigneeID *uint, priority domain.Priority, status domain.ActionStatus, dueDate *time.Time, relatedLinks string) (*domain.ActionItem, error) {
	args := m.Called(ctx, id, title, description, assigneeID, priority, status, dueDate, relatedLinks)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ActionItem), args.Error(1)
}

func (m *MockActionItemUsecase) DeleteActionItem(ctx context.Context, userRole domain.Role, id uint) error {
	args := m.Called(ctx, userRole, id)
	return args.Error(0)
}

func (m *MockActionItemUsecase) GetAllActionItems(ctx context.Context, filters domain.ActionItemFilters, pagination domain.Pagination) ([]*domain.ActionItem, *domain.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*domain.ActionItem), args.Get(1).(*domain.PaginationResult), args.Error(2)
}

func TestActionItemHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates action item", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		createdItem := testutil.NewTestActionItem(1)

		aiUsecase.On("CreateActionItem", mock.Anything, uint(1), "Fix bug", "Description", (*uint)(nil), domain.PriorityHigh, (*time.Time)(nil), "").Return(createdItem, nil)

		reqBody := CreateActionItemRequest{
			PostMortemID: 1,
			Title:        "Fix bug",
			Description:  "Description",
			Priority:     "high",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/action-items", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		aiUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing required fields", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		reqBody := map[string]interface{}{"title": "Fix bug"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/action-items", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		aiUsecase.AssertNotCalled(t, "CreateActionItem")
	})

	t.Run("fails with invalid due_date format", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		dueDate := "invalid-date"
		reqBody := CreateActionItemRequest{
			PostMortemID: 1,
			Title:        "Fix bug",
			Priority:     "high",
			DueDate:      &dueDate,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/action-items", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		aiUsecase.AssertNotCalled(t, "CreateActionItem")
	})

	t.Run("fails with invalid priority", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		reqBody := map[string]interface{}{
			"post_mortem_id": 1,
			"title":          "Fix bug",
			"priority":       "invalid",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/action-items", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		aiUsecase.AssertNotCalled(t, "CreateActionItem")
	})
}

func TestActionItemHandler_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns action item", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		item := testutil.NewTestActionItem(1)
		aiUsecase.On("GetActionItemByID", mock.Anything, uint(1)).Return(item, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/action-items/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusOK, w.Code)
		aiUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/action-items/invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		aiUsecase.AssertNotCalled(t, "GetActionItemByID")
	})
}

func TestActionItemHandler_Update(t *testing.T) {
	t.Parallel()

	t.Run("successfully updates action item", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		updatedItem := testutil.NewTestActionItem(1, func(i *domain.ActionItem) {
			i.Title = "Updated Title"
			i.Status = domain.ActionStatusInProgress
		})

		aiUsecase.On("UpdateActionItem", mock.Anything, uint(1), "Updated Title", "Description", (*uint)(nil), domain.PriorityMedium, domain.ActionStatusInProgress, (*time.Time)(nil), "").Return(updatedItem, nil)

		reqBody := UpdateActionItemRequest{
			Title:       "Updated Title",
			Description: "Description",
			Priority:    "medium",
			Status:      "in_progress",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/action-items/1", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		aiUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid status", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		reqBody := map[string]interface{}{
			"title":    "Updated Title",
			"priority": "medium",
			"status":   "invalid_status",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/action-items/1", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		aiUsecase.AssertNotCalled(t, "UpdateActionItem")
	})
}

func TestActionItemHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("admin can delete action item", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		aiUsecase.On("DeleteActionItem", mock.Anything, domain.RoleAdmin, uint(1)).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/action-items/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["message"], "deleted")

		aiUsecase.AssertExpectations(t)
	})

	t.Run("fails without role", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/action-items/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		// No role set

		handler.Delete(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		aiUsecase.AssertNotCalled(t, "DeleteActionItem")
	})

	t.Run("editor cannot delete action item", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		aiUsecase.On("DeleteActionItem", mock.Anything, domain.RoleEditor, uint(1)).
			Return(domain.ErrForbidden("Only admin can delete action items"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/action-items/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("role", domain.RoleEditor)

		handler.Delete(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		aiUsecase.AssertExpectations(t)
	})
}

func TestActionItemHandler_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns all action items with default pagination", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		items := []*domain.ActionItem{
			testutil.NewTestActionItem(1),
		}
		pagination := &domain.PaginationResult{Page: 1, Limit: 20, Total: 1, TotalPages: 1}

		aiUsecase.On("GetAllActionItems", mock.Anything, mock.Anything, mock.Anything).Return(items, pagination, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/action-items", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "action_items")
		assert.Contains(t, response, "pagination")

		aiUsecase.AssertExpectations(t)
	})

	t.Run("filters by status", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		items := []*domain.ActionItem{}
		pagination := &domain.PaginationResult{Page: 1, Limit: 20, Total: 0, TotalPages: 0}

		aiUsecase.On("GetAllActionItems", mock.Anything, mock.MatchedBy(func(f domain.ActionItemFilters) bool {
			return f.Status == "pending"
		}), mock.Anything).Return(items, pagination, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/action-items?status=pending", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		aiUsecase.AssertExpectations(t)
	})
}

func TestActionItemHandler_GetByPostMortemID(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns action items by post-mortem ID", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		items := []*domain.ActionItem{
			testutil.NewTestActionItem(1),
			testutil.NewTestActionItem(1, func(i *domain.ActionItem) { i.ID = 2 }),
		}
		aiUsecase.On("GetActionItemsByPostMortemID", mock.Anything, uint(1)).Return(items, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/action-items/post-mortem/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByPostMortemID(c)

		assert.Equal(t, http.StatusOK, w.Code)
		aiUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid post-mortem ID", func(t *testing.T) {
		t.Parallel()

		aiUsecase := NewMockActionItemUsecase()
		handler := NewActionItemHandler(aiUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/action-items/post-mortem/invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetByPostMortemID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		aiUsecase.AssertNotCalled(t, "GetActionItemsByPostMortemID")
	})
}

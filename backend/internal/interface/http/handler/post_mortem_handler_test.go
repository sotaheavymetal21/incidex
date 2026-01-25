package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPostMortemUsecase は usecase.PostMortemUsecase のモック実装です
type MockPostMortemUsecase struct {
	mock.Mock
}

func NewMockPostMortemUsecase() *MockPostMortemUsecase {
	return &MockPostMortemUsecase{}
}

func (m *MockPostMortemUsecase) CreatePostMortem(ctx context.Context, authorID uint, incidentID uint, rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned string, fiveWhys *domain.FiveWhysAnalysis) (*domain.PostMortem, error) {
	args := m.Called(ctx, authorID, incidentID, rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned, fiveWhys)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemUsecase) GetPostMortemByID(ctx context.Context, id uint) (*domain.PostMortem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemUsecase) GetPostMortemByIncidentID(ctx context.Context, incidentID uint) (*domain.PostMortem, error) {
	args := m.Called(ctx, incidentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemUsecase) UpdatePostMortem(ctx context.Context, userID uint, userRole domain.Role, id uint, rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned string, fiveWhys *domain.FiveWhysAnalysis) (*domain.PostMortem, error) {
	args := m.Called(ctx, userID, userRole, id, rootCause, impactAnalysis, whatWentWell, whatWentWrong, lessonsLearned, fiveWhys)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemUsecase) PublishPostMortem(ctx context.Context, userID uint, userRole domain.Role, id uint) (*domain.PostMortem, error) {
	args := m.Called(ctx, userID, userRole, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemUsecase) UnpublishPostMortem(ctx context.Context, userID uint, userRole domain.Role, id uint) (*domain.PostMortem, error) {
	args := m.Called(ctx, userID, userRole, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemUsecase) DeletePostMortem(ctx context.Context, userRole domain.Role, id uint) error {
	args := m.Called(ctx, userRole, id)
	return args.Error(0)
}

func (m *MockPostMortemUsecase) GetAllPostMortems(ctx context.Context, filters domain.PostMortemFilters, pagination domain.Pagination) ([]*domain.PostMortem, *domain.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*domain.PostMortem), args.Get(1).(*domain.PaginationResult), args.Error(2)
}

func TestPostMortemHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates post-mortem", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		createdPM := testutil.NewTestPostMortem(1, 1)

		pmUsecase.On("CreatePostMortem", mock.Anything, uint(1), uint(1), "Root cause", "", "", "", "", (*domain.FiveWhysAnalysis)(nil)).Return(createdPM, nil)

		reqBody := CreatePostMortemRequest{
			IncidentID: 1,
			RootCause:  "Root cause",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/post-mortems", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		pmUsecase.AssertExpectations(t)
	})

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		reqBody := CreatePostMortemRequest{IncidentID: 1, RootCause: "Root cause"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/post-mortems", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		// No userID set

		handler.Create(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		pmUsecase.AssertNotCalled(t, "CreatePostMortem")
	})

	t.Run("fails with missing incident_id", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		reqBody := map[string]string{"root_cause": "Root cause"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/post-mortems", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		pmUsecase.AssertNotCalled(t, "CreatePostMortem")
	})
}

func TestPostMortemHandler_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns post-mortem by ID", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		pm := testutil.NewTestPostMortem(1, 1)
		pmUsecase.On("GetPostMortemByID", mock.Anything, uint(1)).Return(pm, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/post-mortems/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusOK, w.Code)
		pmUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/post-mortems/invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		pmUsecase.AssertNotCalled(t, "GetPostMortemByID")
	})
}

func TestPostMortemHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("admin can delete post-mortem", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		pmUsecase.On("DeletePostMortem", mock.Anything, domain.RoleAdmin, uint(1)).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/post-mortems/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["message"], "deleted")

		pmUsecase.AssertExpectations(t)
	})

	t.Run("fails without role", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/post-mortems/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		// No role set

		handler.Delete(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		pmUsecase.AssertNotCalled(t, "DeletePostMortem")
	})

	t.Run("editor cannot delete post-mortem", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		pmUsecase.On("DeletePostMortem", mock.Anything, domain.RoleEditor, uint(1)).
			Return(domain.ErrForbidden("Only admin can delete post-mortems"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/post-mortems/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("role", domain.RoleEditor)

		handler.Delete(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		pmUsecase.AssertExpectations(t)
	})
}

func TestPostMortemHandler_Publish(t *testing.T) {
	t.Parallel()

	t.Run("successfully publishes post-mortem", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		pm := testutil.NewTestPostMortem(1, 1, func(pm *domain.PostMortem) {
			pm.Status = domain.PMStatusPublished
		})
		pmUsecase.On("PublishPostMortem", mock.Anything, uint(1), domain.RoleAdmin, uint(1)).Return(pm, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/post-mortems/1/publish", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Publish(c)

		assert.Equal(t, http.StatusOK, w.Code)
		pmUsecase.AssertExpectations(t)
	})

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/post-mortems/1/publish", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		// No userID set

		handler.Publish(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		pmUsecase.AssertNotCalled(t, "PublishPostMortem")
	})
}

func TestPostMortemHandler_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns all post-mortems with default pagination", func(t *testing.T) {
		t.Parallel()

		pmUsecase := NewMockPostMortemUsecase()
		handler := NewPostMortemHandler(pmUsecase)

		pms := []*domain.PostMortem{
			testutil.NewTestPostMortem(1, 1),
		}
		pagination := &domain.PaginationResult{Page: 1, Limit: 20, Total: 1, TotalPages: 1}

		pmUsecase.On("GetAllPostMortems", mock.Anything, mock.Anything, mock.Anything).Return(pms, pagination, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/post-mortems", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "post_mortems")
		assert.Contains(t, response, "pagination")

		pmUsecase.AssertExpectations(t)
	})
}

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

// MockIncidentUsecase は usecase.IncidentUsecase のモック実装です
type MockIncidentUsecase struct {
	mock.Mock
}

func NewMockIncidentUsecase() *MockIncidentUsecase {
	return &MockIncidentUsecase{}
}

func (m *MockIncidentUsecase) CreateIncident(ctx context.Context, userID uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error) {
	args := m.Called(ctx, userID, title, description, severity, status, impactScope, detectedAt, assigneeID, tagIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Incident), args.Error(1)
}

func (m *MockIncidentUsecase) GetIncidentByID(ctx context.Context, id uint) (*domain.Incident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Incident), args.Error(1)
}

func (m *MockIncidentUsecase) GetAllIncidents(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*domain.Incident), args.Get(1).(*domain.PaginationResult), args.Error(2)
}

func (m *MockIncidentUsecase) UpdateIncident(ctx context.Context, userID uint, role domain.Role, incidentID uint, title, description string, severity domain.Severity, status domain.Status, impactScope string, detectedAt time.Time, assigneeID *uint, tagIDs []uint) (*domain.Incident, error) {
	args := m.Called(ctx, userID, role, incidentID, title, description, severity, status, impactScope, detectedAt, assigneeID, tagIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Incident), args.Error(1)
}

func (m *MockIncidentUsecase) DeleteIncident(ctx context.Context, userRole domain.Role, id uint) error {
	args := m.Called(ctx, userRole, id)
	return args.Error(0)
}

func (m *MockIncidentUsecase) AssignIncident(ctx context.Context, userID uint, incidentID uint, assigneeID *uint) (*domain.Incident, error) {
	args := m.Called(ctx, userID, incidentID, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Incident), args.Error(1)
}

func TestIncidentHandler_Create(t *testing.T) {
	t.Parallel()

	detectedAt := time.Now().Add(-1 * time.Hour)

	t.Run("successfully creates incident", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.Title = "Test Incident"
			i.Severity = domain.SeverityHigh
			i.Status = domain.StatusOpen
		})

		incidentUsecase.On("CreateIncident",
			mock.Anything,
			uint(1),
			"Test Incident",
			"Test Description",
			domain.SeverityHigh,
			domain.StatusOpen,
			"Production",
			mock.MatchedBy(func(t time.Time) bool { return true }),
			(*uint)(nil),
			[]uint{},
		).Return(incident, nil)

		reqBody := CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "Test Description",
			Severity:    "high",
			Status:      "open",
			ImpactScope: "Production",
			DetectedAt:  detectedAt.Format(time.RFC3339),
			TagIDs:      []uint{},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "title")
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing user ID", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "Test Description",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		// userID を設定しない

		handler.Create(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		incidentUsecase.AssertNotCalled(t, "CreateIncident")
	})

	t.Run("fails with invalid severity", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "Test Description",
			Severity:    "invalid",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "CreateIncident")
	})

	t.Run("fails with invalid detected_at format", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "Test Description",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  "invalid-date",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "RFC3339")
		incidentUsecase.AssertNotCalled(t, "CreateIncident")
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("CreateIncident", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrInternal("Database error", nil))

		reqBody := CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "Test Description",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		incidentUsecase.AssertExpectations(t)
	})
}

func TestIncidentHandler_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("successfully retrieves incident by ID", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 123
			i.Title = "Test Incident"
		})

		incidentUsecase.On("GetIncidentByID", mock.Anything, uint(123)).
			Return(incident, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents/123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "title")
		assert.Equal(t, "Test Incident", response["title"])
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID format", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents/invalid", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "Invalid ID")
		incidentUsecase.AssertNotCalled(t, "GetIncidentByID")
	})

	t.Run("fails when incident not found", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("GetIncidentByID", mock.Anything, uint(999)).
			Return(nil, domain.ErrNotFound("Incident"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		incidentUsecase.AssertExpectations(t)
	})
}

func TestIncidentHandler_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("successfully retrieves all incidents with pagination", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidents := []*domain.Incident{
			testutil.NewTestIncident(),
			testutil.NewTestIncident(func(i *domain.Incident) { i.ID = 2 }),
		}

		paginationResult := &domain.PaginationResult{
			Total:      2,
			TotalPages: 1,
			Page:       1,
			Limit:      20,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.MatchedBy(func(f domain.IncidentFilters) bool { return true }),
			mock.MatchedBy(func(p domain.Pagination) bool { return p.Page == 1 && p.Limit == 20 }),
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response IncidentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Incidents, 2)
		assert.NotNil(t, response.Pagination)
		assert.Equal(t, int64(2), response.Pagination.Total)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("successfully retrieves incidents with filters", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidents := []*domain.Incident{
			testutil.NewTestCriticalIncident(),
		}

		paginationResult := &domain.PaginationResult{
			Total:      1,
			TotalPages: 1,
			Page:       1,
			Limit:      20,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.MatchedBy(func(f domain.IncidentFilters) bool {
				return f.Severity == "critical" && f.Status == "open"
			}),
			mock.Anything,
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents?severity=critical&status=open", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response IncidentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Incidents, 1)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("successfully retrieves incidents with custom pagination", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidents := []*domain.Incident{}
		paginationResult := &domain.PaginationResult{
			Total:      50,
			TotalPages: 5,
			Page:       2,
			Limit:      10,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.Anything,
			mock.MatchedBy(func(p domain.Pagination) bool {
				return p.Page == 2 && p.Limit == 10
			}),
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents?page=2&limit=10", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response IncidentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Pagination.Page)
		assert.Equal(t, 10, response.Pagination.Limit)
		incidentUsecase.AssertExpectations(t)
	})
}

func TestIncidentHandler_Update(t *testing.T) {
	t.Parallel()

	detectedAt := time.Now().Add(-2 * time.Hour)

	t.Run("successfully updates incident", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		updatedIncident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 123
			i.Title = "Updated Incident"
		})

		incidentUsecase.On("UpdateIncident",
			mock.Anything,
			uint(1),
			domain.RoleAdmin,
			uint(123),
			"Updated Incident",
			"Updated Description",
			domain.SeverityCritical,
			domain.StatusInvestigating,
			"Production",
			mock.MatchedBy(func(t time.Time) bool { return true }),
			(*uint)(nil),
			[]uint{},
		).Return(updatedIncident, nil)

		reqBody := UpdateIncidentRequest{
			Title:       "Updated Incident",
			Description: "Updated Description",
			Severity:    "critical",
			Status:      "investigating",
			ImpactScope: "Production",
			DetectedAt:  detectedAt.Format(time.RFC3339),
			TagIDs:      []uint{},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "title")
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing user ID", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		// userID を設定しない

		handler.Update(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		incidentUsecase.AssertNotCalled(t, "UpdateIncident")
	})

	t.Run("fails with missing role", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))
		// role を設定しない

		handler.Update(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		incidentUsecase.AssertNotCalled(t, "UpdateIncident")
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/invalid", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "Invalid ID")
		incidentUsecase.AssertNotCalled(t, "UpdateIncident")
	})

	t.Run("fails when user not authorized", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("UpdateIncident", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrForbidden("You do not have permission to update this incident"))

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(2))
		c.Set("role", domain.RoleViewer)

		handler.Update(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		incidentUsecase.AssertExpectations(t)
	})
}

func TestIncidentHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("successfully deletes incident", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("DeleteIncident", mock.Anything, domain.RoleAdmin, uint(123)).
			Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing role", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		// role を設定しない

		handler.Delete(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		incidentUsecase.AssertNotCalled(t, "DeleteIncident")
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/invalid", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "DeleteIncident")
	})

	t.Run("fails when user not authorized", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("DeleteIncident", mock.Anything, domain.RoleViewer, uint(123)).
			Return(domain.ErrForbidden("You do not have permission to delete this incident"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("role", domain.RoleViewer)

		handler.Delete(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails when incident not found", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("DeleteIncident", mock.Anything, domain.RoleAdmin, uint(999)).
			Return(domain.ErrNotFound("Incident"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
		c.Set("role", domain.RoleAdmin)

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		incidentUsecase.AssertExpectations(t)
	})
}

func TestIncidentHandler_AdditionalCreateScenarios(t *testing.T) {
	t.Parallel()

	t.Run("fails with invalid status", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "Test Description",
			Severity:    "high",
			Status:      "invalid-status",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "CreateIncident")
	})

	t.Run("fails with missing title", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := CreateIncidentRequest{
			Description: "Test Description",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "CreateIncident")
	})

	t.Run("fails with missing description", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := CreateIncidentRequest{
			Title:      "Test Incident",
			Severity:   "high",
			Status:     "open",
			DetectedAt: time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "CreateIncident")
	})

	t.Run("fails with title exceeding max length", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		// Create a title longer than 500 characters
		longTitle := make([]byte, 501)
		for i := range longTitle {
			longTitle[i] = 'a'
		}

		reqBody := CreateIncidentRequest{
			Title:       string(longTitle),
			Description: "Test Description",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "CreateIncident")
	})

	t.Run("successfully creates incident with assignee and tags", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.Title = "Test Incident"
			assigneeID := uint(5)
			i.AssigneeID = &assigneeID
		})

		assigneeID := uint(5)
		tagIDs := []uint{1, 2, 3}

		incidentUsecase.On("CreateIncident",
			mock.Anything,
			uint(1),
			"Test Incident",
			"Test Description",
			domain.SeverityHigh,
			domain.StatusOpen,
			"Production",
			mock.MatchedBy(func(t time.Time) bool { return true }),
			&assigneeID,
			tagIDs,
		).Return(incident, nil)

		reqBody := CreateIncidentRequest{
			Title:       "Test Incident",
			Description: "Test Description",
			Severity:    "high",
			Status:      "open",
			ImpactScope: "Production",
			DetectedAt:  time.Now().Format(time.RFC3339),
			AssigneeID:  &assigneeID,
			TagIDs:      tagIDs,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", uint(1))

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		incidentUsecase.AssertExpectations(t)
	})
}

func TestIncidentHandler_AdditionalGetAllScenarios(t *testing.T) {
	t.Parallel()

	t.Run("successfully retrieves empty results", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidents := []*domain.Incident{}

		paginationResult := &domain.PaginationResult{
			Total:      0,
			TotalPages: 0,
			Page:       1,
			Limit:      20,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response IncidentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Incidents, 0)
		assert.Equal(t, int64(0), response.Pagination.Total)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("successfully retrieves incidents with assigned_to_id filter", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		assigneeID := uint(5)
		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) { i.AssigneeID = &assigneeID }),
		}

		paginationResult := &domain.PaginationResult{
			Total:      1,
			TotalPages: 1,
			Page:       1,
			Limit:      20,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.MatchedBy(func(f domain.IncidentFilters) bool {
				return f.AssignedToID != nil && *f.AssignedToID == 5
			}),
			mock.Anything,
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents?assigned_to_id=5", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response IncidentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Incidents, 1)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("successfully retrieves incidents with tag_ids filter", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidents := []*domain.Incident{
			testutil.NewTestIncident(),
		}

		paginationResult := &domain.PaginationResult{
			Total:      1,
			TotalPages: 1,
			Page:       1,
			Limit:      20,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.MatchedBy(func(f domain.IncidentFilters) bool {
				return len(f.TagIDs) == 3 && f.TagIDs[0] == 1 && f.TagIDs[1] == 2 && f.TagIDs[2] == 3
			}),
			mock.Anything,
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents?tag_ids=1,2,3", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response IncidentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Incidents, 1)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("successfully retrieves incidents with search filter", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) { i.Title = "Database connection error" }),
		}

		paginationResult := &domain.PaginationResult{
			Total:      1,
			TotalPages: 1,
			Page:       1,
			Limit:      20,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.MatchedBy(func(f domain.IncidentFilters) bool {
				return f.Search == "database"
			}),
			mock.Anything,
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents?search=database", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response IncidentListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Incidents, 1)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("successfully retrieves incidents with custom sort order", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidents := []*domain.Incident{
			testutil.NewTestIncident(),
		}

		paginationResult := &domain.PaginationResult{
			Total:      1,
			TotalPages: 1,
			Page:       1,
			Limit:      20,
		}

		incidentUsecase.On("GetAllIncidents",
			mock.Anything,
			mock.MatchedBy(func(f domain.IncidentFilters) bool {
				return f.SortBy == "severity" && f.Order == "asc"
			}),
			mock.Anything,
		).Return(incidents, paginationResult, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents?sort=severity&order=asc", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("GetAllIncidents", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, nil, domain.ErrInternal("Database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		incidentUsecase.AssertExpectations(t)
	})
}

func TestIncidentHandler_AdditionalUpdateScenarios(t *testing.T) {
	t.Parallel()

	t.Run("fails when incident not found", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incidentUsecase.On("UpdateIncident", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrNotFound("Incident"))

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/999", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid severity in update", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "invalid-severity",
			Status:      "open",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "UpdateIncident")
	})

	t.Run("fails with invalid status in update", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "high",
			Status:      "invalid-status",
			DetectedAt:  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		incidentUsecase.AssertNotCalled(t, "UpdateIncident")
	})

	t.Run("fails with invalid detected_at format in update", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		reqBody := UpdateIncidentRequest{
			Title:       "Updated",
			Description: "Updated",
			Severity:    "high",
			Status:      "open",
			DetectedAt:  "not-a-date",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))
		c.Set("role", domain.RoleAdmin)

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "RFC3339")
		incidentUsecase.AssertNotCalled(t, "UpdateIncident")
	})
}

func TestIncidentHandler_AssignIncident(t *testing.T) {
	t.Parallel()

	t.Run("successfully assigns incident to user", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		assigneeID := uint(10)
		incident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 123
			i.AssigneeID = &assigneeID
		})

		incidentUsecase.On("AssignIncident", mock.Anything, uint(1), uint(123), &assigneeID).
			Return(incident, nil)

		reqBody := AssignIncidentRequest{
			AssigneeID: &assigneeID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123/assign", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))

		handler.AssignIncident(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "id")
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("successfully unassigns incident", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		incident := testutil.NewTestIncident(func(i *domain.Incident) {
			i.ID = 123
			i.AssigneeID = nil
		})

		incidentUsecase.On("AssignIncident", mock.Anything, uint(1), uint(123), (*uint)(nil)).
			Return(incident, nil)

		reqBody := AssignIncidentRequest{
			AssigneeID: nil,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123/assign", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))

		handler.AssignIncident(c)

		assert.Equal(t, http.StatusOK, w.Code)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid incident ID", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		assigneeID := uint(10)
		reqBody := AssignIncidentRequest{
			AssigneeID: &assigneeID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/invalid/assign", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}
		c.Set("userID", uint(1))

		handler.AssignIncident(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "invalid incident ID")
		incidentUsecase.AssertNotCalled(t, "AssignIncident")
	})

	t.Run("fails with missing user ID", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		assigneeID := uint(10)
		reqBody := AssignIncidentRequest{
			AssigneeID: &assigneeID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123/assign", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		// userID を設定しない

		handler.AssignIncident(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"], "User ID not found")
		incidentUsecase.AssertNotCalled(t, "AssignIncident")
	})

	t.Run("fails when incident not found", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		assigneeID := uint(10)
		incidentUsecase.On("AssignIncident", mock.Anything, uint(1), uint(999), &assigneeID).
			Return(nil, domain.ErrNotFound("Incident"))

		reqBody := AssignIncidentRequest{
			AssigneeID: &assigneeID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/999/assign", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
		c.Set("userID", uint(1))

		handler.AssignIncident(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails when assignee user not found", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		assigneeID := uint(999)
		incidentUsecase.On("AssignIncident", mock.Anything, uint(1), uint(123), &assigneeID).
			Return(nil, domain.ErrNotFound("User"))

		reqBody := AssignIncidentRequest{
			AssigneeID: &assigneeID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123/assign", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(1))

		handler.AssignIncident(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		incidentUsecase.AssertExpectations(t)
	})

	t.Run("fails when user not authorized to assign", func(t *testing.T) {
		t.Parallel()

		incidentUsecase := NewMockIncidentUsecase()
		handler := NewIncidentHandler(incidentUsecase)

		assigneeID := uint(10)
		incidentUsecase.On("AssignIncident", mock.Anything, uint(2), uint(123), &assigneeID).
			Return(nil, domain.ErrForbidden("You do not have permission to assign this incident"))

		reqBody := AssignIncidentRequest{
			AssigneeID: &assigneeID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/incidents/123/assign", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}
		c.Set("userID", uint(2))

		handler.AssignIncident(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		incidentUsecase.AssertExpectations(t)
	})
}

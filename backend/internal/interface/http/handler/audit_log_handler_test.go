package handler

import (
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

// MockAuditLogUsecase は usecase.AuditLogUsecase のモック実装です
type MockAuditLogUsecase struct {
	mock.Mock
}

func NewMockAuditLogUsecase() *MockAuditLogUsecase {
	return &MockAuditLogUsecase{}
}

func (m *MockAuditLogUsecase) GetAll(ctx context.Context, filters domain.AuditLogFilters) ([]*domain.AuditLog, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.AuditLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditLogUsecase) GetByID(ctx context.Context, id uint) (*domain.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

func TestAuditLogHandler_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns audit logs with default pagination", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		logs := []*domain.AuditLog{
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ID = 1 }),
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ID = 2 }),
		}

		auditLogUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(f domain.AuditLogFilters) bool {
			return f.Page == 1 && f.Limit == 50
		})).Return(logs, int64(2), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "logs")
		assert.Contains(t, response, "pagination")

		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(50), pagination["limit"])
		assert.Equal(t, float64(2), pagination["total"])

		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("successfully returns audit logs with custom pagination", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		logs := []*domain.AuditLog{
			testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ID = 1 }),
		}

		auditLogUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(f domain.AuditLogFilters) bool {
			return f.Page == 2 && f.Limit == 10
		})).Return(logs, int64(15), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?page=2&limit=10", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(2), pagination["page"])
		assert.Equal(t, float64(10), pagination["limit"])
		assert.Equal(t, float64(15), pagination["total"])
		assert.Equal(t, float64(2), pagination["total_pages"])

		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("filters by user ID", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		logs := []*domain.AuditLog{
			testutil.NewTestAuditLog(),
		}

		auditLogUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(f domain.AuditLogFilters) bool {
			return f.UserID != nil && *f.UserID == uint(5)
		})).Return(logs, int64(1), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?user_id=5", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("filters by action", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		logs := []*domain.AuditLog{
			testutil.NewTestAuditLog(),
		}

		auditLogUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(f domain.AuditLogFilters) bool {
			return f.Action != nil && *f.Action == domain.AuditActionCreate
		})).Return(logs, int64(1), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?action=create", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("filters by resource type", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		logs := []*domain.AuditLog{
			testutil.NewTestAuditLog(),
		}

		auditLogUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(f domain.AuditLogFilters) bool {
			return f.ResourceType != nil && *f.ResourceType == "incident"
		})).Return(logs, int64(1), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?resource_type=incident", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("filters by date range", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		logs := []*domain.AuditLog{
			testutil.NewTestAuditLog(),
		}

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		auditLogUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(f domain.AuditLogFilters) bool {
			return f.StartDate != nil && f.EndDate != nil
		})).Return(logs, int64(1), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/audit-logs?start_date=" + startDate.Format(time.RFC3339) + "&end_date=" + endDate.Format(time.RFC3339)
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		auditLogUsecase.On("GetAll", mock.Anything, mock.Anything).
			Return(nil, int64(0), domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)

		handler.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		auditLogUsecase.AssertExpectations(t)
	})
}

func TestAuditLogHandler_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns audit log by ID", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		log := testutil.NewTestAuditLog(func(l *domain.AuditLog) { l.ID = 1 })

		auditLogUsecase.On("GetByID", mock.Anything, uint(1)).Return(log, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.AuditLog
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, uint(1), response.ID)

		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs/invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		auditLogUsecase.AssertNotCalled(t, "GetByID")
	})

	t.Run("returns not found when log doesn't exist", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		auditLogUsecase.On("GetByID", mock.Anything, uint(999)).Return(nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs/999", nil)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		auditLogUsecase.AssertExpectations(t)
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		auditLogUsecase := NewMockAuditLogUsecase()
		handler := NewAuditLogHandler(auditLogUsecase)

		auditLogUsecase.On("GetByID", mock.Anything, uint(1)).
			Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		auditLogUsecase.AssertExpectations(t)
	})
}

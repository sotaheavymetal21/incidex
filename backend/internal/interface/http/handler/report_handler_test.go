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

// MockReportUsecase は usecase.ReportUsecase のモック実装です
type MockReportUsecase struct {
	mock.Mock
}

func NewMockReportUsecase() *MockReportUsecase {
	return &MockReportUsecase{}
}

func (m *MockReportUsecase) GetMonthlyReport(ctx context.Context, year, month int) (*domain.MonthlyReport, error) {
	args := m.Called(ctx, year, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MonthlyReport), args.Error(1)
}

func (m *MockReportUsecase) GetCustomReport(ctx context.Context, startDate, endDate time.Time) (*domain.MonthlyReport, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MonthlyReport), args.Error(1)
}

func TestReportHandler_GetMonthlyReport(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns monthly report with specified year and month", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		report := testutil.NewTestMonthlyReport()

		reportUsecase.On("GetMonthlyReport", mock.Anything, 2024, 6).Return(report, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2024&month=6", nil)

		handler.GetMonthlyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.MonthlyReport
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 10, response.Summary.TotalIncidents)

		reportUsecase.AssertExpectations(t)
	})

	t.Run("uses default year and month when not specified", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		report := testutil.NewTestMonthlyReport()
		now := time.Now()

		reportUsecase.On("GetMonthlyReport", mock.Anything, now.Year(), int(now.Month())).Return(report, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly", nil)

		handler.GetMonthlyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		reportUsecase.AssertExpectations(t)
	})

	t.Run("fails with invalid month", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2024&month=13", nil)

		handler.GetMonthlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Month must be between 1 and 12")

		reportUsecase.AssertNotCalled(t, "GetMonthlyReport")
	})

	t.Run("fails with invalid month zero", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2024&month=0", nil)

		handler.GetMonthlyReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		reportUsecase.AssertNotCalled(t, "GetMonthlyReport")
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		reportUsecase.On("GetMonthlyReport", mock.Anything, 2024, 1).
			Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2024&month=1", nil)

		handler.GetMonthlyReport(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		reportUsecase.AssertExpectations(t)
	})
}

func TestReportHandler_GetCustomReport(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns custom report", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		report := testutil.NewTestMonthlyReport()

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

		reportUsecase.On("GetCustomReport", mock.Anything, startDate, endDate).Return(report, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/reports/custom?start_date=" + startDate.Format(time.RFC3339) + "&end_date=" + endDate.Format(time.RFC3339)
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetCustomReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.MonthlyReport
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 10, response.Summary.TotalIncidents)

		reportUsecase.AssertExpectations(t)
	})

	t.Run("fails with missing start_date", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		endDate := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/reports/custom?end_date=" + endDate.Format(time.RFC3339)
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "start_date and end_date are required")

		reportUsecase.AssertNotCalled(t, "GetCustomReport")
	})

	t.Run("fails with missing end_date", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/reports/custom?start_date=" + startDate.Format(time.RFC3339)
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "start_date and end_date are required")

		reportUsecase.AssertNotCalled(t, "GetCustomReport")
	})

	t.Run("fails with invalid start_date format", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		endDate := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/reports/custom?start_date=invalid&end_date=" + endDate.Format(time.RFC3339)
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid start_date format")

		reportUsecase.AssertNotCalled(t, "GetCustomReport")
	})

	t.Run("fails with invalid end_date format", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/reports/custom?start_date=" + startDate.Format(time.RFC3339) + "&end_date=invalid"
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid end_date format")

		reportUsecase.AssertNotCalled(t, "GetCustomReport")
	})

	t.Run("fails when end_date is before start_date", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		startDate := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/reports/custom?start_date=" + startDate.Format(time.RFC3339) + "&end_date=" + endDate.Format(time.RFC3339)
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetCustomReport(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "end_date must be after start_date")

		reportUsecase.AssertNotCalled(t, "GetCustomReport")
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		reportUsecase := NewMockReportUsecase()
		handler := NewReportHandler(reportUsecase)

		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

		reportUsecase.On("GetCustomReport", mock.Anything, startDate, endDate).
			Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		url := "/api/v1/reports/custom?start_date=" + startDate.Format(time.RFC3339) + "&end_date=" + endDate.Format(time.RFC3339)
		c.Request = httptest.NewRequest(http.MethodGet, url, nil)

		handler.GetCustomReport(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		reportUsecase.AssertExpectations(t)
	})
}

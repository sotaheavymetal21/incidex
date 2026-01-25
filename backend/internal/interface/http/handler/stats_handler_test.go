package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"incidex/internal/domain"
	"incidex/internal/testutil/mocks"
	"incidex/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStatsHandler_GetDashboardStats(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns dashboard stats with default period", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
		handler := NewStatsHandler(statsUsecase)

		// Setup mock expectations
		cacheRepo.On("Get", mock.Anything, "stats:dashboard:daily").Return("", domain.ErrNotFound("cache miss"))

		incidentRepo.On("Count", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			if ptr, ok := args.Get(0).(*int64); ok {
				*ptr = 10
			}
		})

		// Setup severity counts
		for _, severity := range []domain.Severity{domain.SeverityCritical, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityLow} {
			incidentRepo.On("CountBySeverity", severity, mock.Anything).Return(nil)
		}

		// Setup status counts
		for _, status := range []domain.Status{domain.StatusOpen, domain.StatusInvestigating, domain.StatusResolved, domain.StatusClosed} {
			incidentRepo.On("CountByStatus", status, mock.Anything).Return(nil)
		}

		incidentRepo.On("FindRecent", 10).Return([]*domain.Incident{}, nil)
		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)
		cacheRepo.On("Set", mock.Anything, "stats:dashboard:daily", mock.Anything, mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats/dashboard", nil)

		handler.GetDashboardStats(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response usecase.DashboardStats
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotNil(t, response.BySeverity)
		assert.NotNil(t, response.ByStatus)
	})

	t.Run("accepts custom period parameter", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
		handler := NewStatsHandler(statsUsecase)

		// Setup mock for weekly period
		cacheRepo.On("Get", mock.Anything, "stats:dashboard:weekly").Return("", domain.ErrNotFound("cache miss"))

		incidentRepo.On("Count", mock.Anything).Return(nil)

		for _, severity := range []domain.Severity{domain.SeverityCritical, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityLow} {
			incidentRepo.On("CountBySeverity", severity, mock.Anything).Return(nil)
		}

		for _, status := range []domain.Status{domain.StatusOpen, domain.StatusInvestigating, domain.StatusResolved, domain.StatusClosed} {
			incidentRepo.On("CountByStatus", status, mock.Anything).Return(nil)
		}

		incidentRepo.On("FindRecent", 10).Return([]*domain.Incident{}, nil)
		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)
		cacheRepo.On("Set", mock.Anything, "stats:dashboard:weekly", mock.Anything, mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats/dashboard?period=weekly", nil)

		handler.GetDashboardStats(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestStatsHandler_GetSLAMetrics(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns SLA metrics", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
		handler := NewStatsHandler(statsUsecase)

		metrics := &domain.SLAMetrics{
			TotalIncidents:    100,
			ResolvedIncidents: 95,
			SLAComplianceRate: 95.0,
			AverageMTTR:       24.0,
		}

		cacheRepo.On("Get", mock.Anything, "stats:sla").Return("", domain.ErrNotFound("cache miss"))
		incidentRepo.On("GetSLAMetrics").Return(metrics, nil)
		cacheRepo.On("Set", mock.Anything, "stats:sla", mock.Anything, mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats/sla", nil)

		handler.GetSLAMetrics(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.SLAMetrics
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 95.0, response.SLAComplianceRate)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
		handler := NewStatsHandler(statsUsecase)

		cacheRepo.On("Get", mock.Anything, "stats:sla").Return("", domain.ErrNotFound("cache miss"))
		incidentRepo.On("GetSLAMetrics").Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats/sla", nil)

		handler.GetSLAMetrics(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestStatsHandler_GetTagStats(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns tag stats", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
		handler := NewStatsHandler(statsUsecase)

		incidents := []*domain.Incident{
			{
				ID: 1,
				Tags: []domain.Tag{
					{ID: 1, Name: "network", Color: "#ff0000"},
				},
			},
			{
				ID: 2,
				Tags: []domain.Tag{
					{ID: 1, Name: "network", Color: "#ff0000"},
					{ID: 2, Name: "database", Color: "#00ff00"},
				},
			},
		}

		cacheRepo.On("Get", mock.Anything, "stats:tags").Return("", domain.ErrNotFound("cache miss"))
		incidentRepo.On("GetAllIncidents").Return(incidents, nil)
		cacheRepo.On("Set", mock.Anything, "stats:tags", mock.Anything, mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats/tags", nil)

		handler.GetTagStats(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "tag_stats")
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		statsUsecase := usecase.NewStatsUsecase(incidentRepo, cacheRepo)
		handler := NewStatsHandler(statsUsecase)

		cacheRepo.On("Get", mock.Anything, "stats:tags").Return("", domain.ErrNotFound("cache miss"))
		incidentRepo.On("GetAllIncidents").Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stats/tags", nil)

		handler.GetTagStats(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

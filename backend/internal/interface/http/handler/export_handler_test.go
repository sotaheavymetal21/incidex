package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"incidex/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportHandler_ExportIncidentsCSV(t *testing.T) {
	t.Parallel()

	t.Run("successfully exports incidents as CSV", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockIncidentUsecase()
		handler := &ExportHandler{incidentUsecase: mockUsecase}

		now := time.Now()
		creator := &domain.User{ID: 1, Name: "Test User", Email: "test@example.com"}
		incidents := []*domain.Incident{
			{
				ID:          1,
				Title:       "Test Incident",
				Description: "Test description",
				Severity:    domain.SeverityHigh,
				Status:      domain.StatusOpen,
				ImpactScope: "Production",
				DetectedAt:  now,
				Creator:     creator,
				CreatedAt:   now,
			},
		}

		mockUsecase.On("GetAllIncidents", mock.Anything, mock.Anything, mock.Anything).
			Return(incidents, &domain.PaginationResult{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/export/incidents", nil)

		handler.ExportIncidentsCSV(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
		assert.Contains(t, w.Body.String(), "Test Incident")
	})

	t.Run("successfully exports with filters", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockIncidentUsecase()
		handler := &ExportHandler{incidentUsecase: mockUsecase}

		mockUsecase.On("GetAllIncidents", mock.Anything, mock.MatchedBy(func(f domain.IncidentFilters) bool {
			return f.Severity == "high" && f.Status == "open"
		}), mock.Anything).Return([]*domain.Incident{}, &domain.PaginationResult{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/export/incidents?severity=high&status=open", nil)

		handler.ExportIncidentsCSV(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("fails when usecase returns error", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockIncidentUsecase()
		handler := &ExportHandler{incidentUsecase: mockUsecase}

		mockUsecase.On("GetAllIncidents", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/export/incidents", nil)

		handler.ExportIncidentsCSV(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestExportHandler_ExportIncidentPDF(t *testing.T) {
	t.Parallel()

	t.Run("successfully exports incident as PDF", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockIncidentUsecase()
		handler := NewExportHandler(mockUsecase)

		now := time.Now()
		creator := &domain.User{ID: 1, Name: "Test User", Email: "test@example.com"}
		incident := &domain.Incident{
			ID:          1,
			Title:       "Test Incident",
			Description: "Test description",
			Severity:    domain.SeverityHigh,
			Status:      domain.StatusOpen,
			ImpactScope: "Production",
			DetectedAt:  now,
			Creator:     creator,
			CreatedAt:   now,
		}

		mockUsecase.On("GetIncidentByID", mock.Anything, uint(1)).Return(incident, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/export/incidents/1/pdf", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.ExportIncidentPDF(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	})

	t.Run("fails with invalid incident ID", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockIncidentUsecase()
		handler := NewExportHandler(mockUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/export/incidents/invalid/pdf", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.ExportIncidentPDF(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails when incident not found", func(t *testing.T) {
		t.Parallel()

		mockUsecase := NewMockIncidentUsecase()
		handler := NewExportHandler(mockUsecase)

		mockUsecase.On("GetIncidentByID", mock.Anything, uint(999)).
			Return(nil, domain.ErrNotFound("incident not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/export/incidents/999/pdf", nil)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.ExportIncidentPDF(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

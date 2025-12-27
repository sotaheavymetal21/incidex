package handler

import (
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/pdf"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportUsecase  usecase.ReportUsecase
	incidentUsecase usecase.IncidentUsecase
	pdfService     *pdf.IncidentPDFService
}

func NewReportHandler(u usecase.ReportUsecase, incidentUsecase usecase.IncidentUsecase) *ReportHandler {
	return &ReportHandler{
		reportUsecase: u,
		incidentUsecase: incidentUsecase,
		pdfService:    pdf.NewIncidentPDFService(),
	}
}

// GetMonthlyReport generates a monthly report
// @Summary Get monthly report
// @Description Get comprehensive monthly incident report for a specific month
// @Tags reports
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g., 2024)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} domain.MonthlyReport
// @Router /reports/monthly [get]
func (h *ReportHandler) GetMonthlyReport(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	// Default to current month if not specified
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil {
			if m >= 1 && m <= 12 {
				month = m
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Month must be between 1 and 12"})
				return
			}
		}
	}

	report, err := h.reportUsecase.GetMonthlyReport(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetCustomReport generates a custom period report
// @Summary Get custom period report
// @Description Get incident report for a custom date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (RFC3339 format)"
// @Param end_date query string true "End date (RFC3339 format)"
// @Success 200 {object} domain.MonthlyReport
// @Router /reports/custom [get]
func (h *ReportHandler) GetCustomReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use RFC3339 format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use RFC3339 format"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
		return
	}

	report, err := h.reportUsecase.GetCustomReport(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetMonthlyReportPDF generates a monthly report in PDF format
// @Summary Get monthly report PDF
// @Description Get comprehensive monthly incident report for a specific month in PDF format
// @Tags reports
// @Accept json
// @Produce application/pdf
// @Param year query int false "Year (e.g., 2024)"
// @Param month query int false "Month (1-12)"
// @Success 200 {file} file "PDF file"
// @Router /reports/monthly/pdf [get]
func (h *ReportHandler) GetMonthlyReportPDF(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	// Default to current month if not specified
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil {
			if m >= 1 && m <= 12 {
				month = m
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Month must be between 1 and 12"})
				return
			}
		}
	}

	// Calculate start and end dates for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second) // Last second of the month

	// Get all incidents in the month
	filters := domain.IncidentFilters{}
	pagination := domain.Pagination{
		Page:  1,
		Limit: 10000,
	}

	incidents, _, err := h.incidentUsecase.GetAllIncidents(c.Request.Context(), filters, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter by date range
	filteredIncidents := make([]*domain.Incident, 0)
	for _, incident := range incidents {
		if (incident.DetectedAt.After(startDate) || incident.DetectedAt.Equal(startDate)) &&
			(incident.DetectedAt.Before(endDate) || incident.DetectedAt.Equal(endDate)) {
			filteredIncidents = append(filteredIncidents, incident)
		}
	}

	// Calculate statistics
	stats := calculateMonthlyStats(filteredIncidents)

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateSummaryReport(filteredIncidents, startDate, endDate, stats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate PDF: %v", err)})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("monthly_report_%d_%02d.pdf", year, month)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func calculateMonthlyStats(incidents []*domain.Incident) *pdf.SummaryStats {
	stats := &pdf.SummaryStats{
		TotalIncidents: len(incidents),
		BySeverity:     make(map[string]int),
		ByStatus:       make(map[string]int),
	}

	totalResolutionTime := 0.0
	resolvedCount := 0

	for _, incident := range incidents {
		// Count by severity
		stats.BySeverity[string(incident.Severity)]++

		// Count by status
		stats.ByStatus[string(incident.Status)]++

		// Count resolved
		if incident.Status == "resolved" || incident.Status == "closed" {
			stats.ResolvedCount++

			// Calculate MTTR
			if incident.ResolvedAt != nil {
				duration := incident.ResolvedAt.Sub(incident.DetectedAt).Hours()
				totalResolutionTime += duration
				resolvedCount++
			}
		}

		// Count SLA violations
		if incident.SLAViolated {
			stats.SLAViolatedCount++
		}
	}

	// Calculate average MTTR
	if resolvedCount > 0 {
		stats.AverageMTTR = totalResolutionTime / float64(resolvedCount)
	}

	return stats
}

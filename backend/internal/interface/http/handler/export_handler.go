package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/pdf"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	incidentUsecase usecase.IncidentUsecase
	pdfService      *pdf.IncidentPDFService
}

func NewExportHandler(incidentUsecase usecase.IncidentUsecase) *ExportHandler {
	return &ExportHandler{
		incidentUsecase: incidentUsecase,
		pdfService:      pdf.NewIncidentPDFService(),
	}
}

// ExportIncidentsCSV godoc
// @Summary Export incidents to CSV
// @Description Export all incidents (with optional filters) to CSV format
// @Tags export
// @Accept json
// @Produce text/csv
// @Param severity query string false "Filter by severity"
// @Param status query string false "Filter by status"
// @Param tag_ids query string false "Filter by tag IDs (comma-separated)"
// @Param search query string false "Search in title/description"
// @Success 200 {file} file "CSV file"
// @Failure 500 {object} map[string]string
// @Router /api/export/incidents [get]
// @Security BearerAuth
func (h *ExportHandler) ExportIncidentsCSV(c *gin.Context) {
	// Parse filters from query parameters
	filters := domain.IncidentFilters{
		Severity: c.Query("severity"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
	}

	// Parse tag IDs if provided
	if tagIDsStr := c.Query("tag_ids"); tagIDsStr != "" {
		tagIDsParts := strings.Split(tagIDsStr, ",")
		for _, part := range tagIDsParts {
			if tagID, err := strconv.ParseUint(part, 10, 32); err == nil {
				filters.TagIDs = append(filters.TagIDs, uint(tagID))
			}
		}
	}

	// Get all incidents without pagination (set a large limit)
	pagination := domain.Pagination{
		Page:  1,
		Limit: 10000, // Large number to get all incidents
	}

	incidents, _, err := h.incidentUsecase.GetAllIncidents(c.Request.Context(), filters, pagination)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write UTF-8 BOM for Excel compatibility
	buf.WriteString("\xEF\xBB\xBF")

	// Write CSV header
	header := []string{
		"ID",
		"タイトル",
		"説明",
		"重要度",
		"ステータス",
		"影響範囲",
		"検出日時",
		"解決日時",
		"担当者",
		"作成者",
		"タグ",
		"作成日時",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// Write incident data
	for _, incident := range incidents {
		var assigneeName string
		if incident.Assignee != nil {
			assigneeName = incident.Assignee.Name
		}

		var resolvedAt string
		if incident.ResolvedAt != nil {
			resolvedAt = incident.ResolvedAt.Format("2006-01-02 15:04:05")
		}

		// Collect tag names
		var tagNames []string
		for _, tag := range incident.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		tagsStr := strings.Join(tagNames, ", ")

		row := []string{
			fmt.Sprintf("%d", incident.ID),
			incident.Title,
			incident.Description,
			string(incident.Severity),
			string(incident.Status),
			incident.ImpactScope,
			incident.DetectedAt.Format("2006-01-02 15:04:05"),
			resolvedAt,
			assigneeName,
			incident.Creator.Name,
			tagsStr,
			incident.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV row"})
			return
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize CSV"})
		return
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=incidents.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Length", strconv.Itoa(buf.Len()))

	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

// ExportIncidentPDF godoc
// @Summary Export single incident to PDF
// @Description Generate a PDF report for a single incident
// @Tags export
// @Produce application/pdf
// @Param id path int true "Incident ID"
// @Success 200 {file} file "PDF file"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/export/incidents/{id}/pdf [get]
// @Security BearerAuth
func (h *ExportHandler) ExportIncidentPDF(c *gin.Context) {
	// Parse incident ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	// Get incident
	incident, err := h.incidentUsecase.GetIncidentByID(c.Request.Context(), uint(id))
	if err != nil {
		HandleError(c, err)
		return
	}

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateIncidentReport(incident)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate PDF: %v", err)})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("incident_%d_%s.pdf", incident.ID, time.Now().Format("20060102"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ExportSummaryPDF godoc
// @Summary Export summary report to PDF
// @Description Generate a PDF summary report for a date range
// @Tags export
// @Produce application/pdf
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param severity query string false "Filter by severity"
// @Param status query string false "Filter by status"
// @Success 200 {file} file "PDF file"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/export/summary/pdf [get]
// @Security BearerAuth
func (h *ExportHandler) ExportSummaryPDF(c *gin.Context) {
	// Parse date range
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	// Set end date to end of day
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Parse filters
	filters := domain.IncidentFilters{
		Severity: c.Query("severity"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
	}

	// Get all incidents in date range without pagination
	pagination := domain.Pagination{
		Page:  1,
		Limit: 10000,
	}

	incidents, _, err := h.incidentUsecase.GetAllIncidents(c.Request.Context(), filters, pagination)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Filter by date range
	filteredIncidents := []*domain.Incident{}
	for _, incident := range incidents {
		if incident.DetectedAt.After(startDate) && incident.DetectedAt.Before(endDate) {
			filteredIncidents = append(filteredIncidents, incident)
		}
	}

	// Calculate statistics
	stats := calculateSummaryStats(filteredIncidents)

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateSummaryReport(filteredIncidents, startDate, endDate, stats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate PDF: %v", err)})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("summary_%s_to_%s.pdf",
		startDate.Format("20060102"),
		endDate.Format("20060102"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func calculateSummaryStats(incidents []*domain.Incident) *pdf.SummaryStats {
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
		if incident.Status == domain.StatusResolved || incident.Status == domain.StatusClosed {
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

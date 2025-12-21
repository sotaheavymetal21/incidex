package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	incidentUsecase usecase.IncidentUsecase
}

func NewExportHandler(incidentUsecase usecase.IncidentUsecase) *ExportHandler {
	return &ExportHandler{
		incidentUsecase: incidentUsecase,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

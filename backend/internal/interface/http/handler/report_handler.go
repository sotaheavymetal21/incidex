package handler

import (
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportUsecase usecase.ReportUsecase
}

func NewReportHandler(u usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{reportUsecase: u}
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

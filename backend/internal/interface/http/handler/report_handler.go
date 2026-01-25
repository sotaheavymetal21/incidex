package handler

import (
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ReportHandler はレポート関連の HTTP handler を提供します
type ReportHandler struct {
	reportUsecase usecase.ReportUsecase
}

// NewReportHandler は新しい ReportHandler を作成します
func NewReportHandler(u usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{
		reportUsecase: u,
	}
}

// GetMonthlyReport godoc
// @Summary 月次レポートを生成します
// @Description 特定の月の包括的なインシデント月次レポートを取得します
// @Tags reports
// @Accept json
// @Produce json
// @Param year query int true "年（例: 2024）"
// @Param month query int true "月（1-12）"
// @Success 200 {object} domain.MonthlyReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports/monthly [get]
// @Security BearerAuth
func (h *ReportHandler) GetMonthlyReport(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	// 指定がない場合は現在の月をデフォルトとする
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetCustomReport godoc
// @Summary カスタム期間レポートを生成します
// @Description カスタム日付範囲のインシデントレポートを取得します
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "開始日（RFC3339 形式）"
// @Param end_date query string true "終了日（RFC3339 形式）"
// @Success 200 {object} domain.MonthlyReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports/custom [get]
// @Security BearerAuth
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

package handler

import (
	"incidex/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsUsecase *usecase.StatsUsecase
}

func NewStatsHandler(statsUsecase *usecase.StatsUsecase) *StatsHandler {
	return &StatsHandler{
		statsUsecase: statsUsecase,
	}
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Retrieve statistics for the dashboard including counts, distributions, and trends
// @Tags stats
// @Accept json
// @Produce json
// @Param period query string false "Period for trend data (daily, weekly, monthly)" default(daily)
// @Success 200 {object} usecase.DashboardStats
// @Failure 500 {object} map[string]string
// @Router /api/stats/dashboard [get]
// @Security BearerAuth
func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

	stats, err := h.statsUsecase.GetDashboardStats(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSLAMetrics godoc
// @Summary Get SLA performance metrics
// @Description Retrieve SLA metrics including compliance rate, MTTR, and violations
// @Tags stats
// @Accept json
// @Produce json
// @Success 200 {object} domain.SLAMetrics
// @Failure 500 {object} map[string]string
// @Router /api/stats/sla [get]
// @Security BearerAuth
func (h *StatsHandler) GetSLAMetrics(c *gin.Context) {
	metrics, err := h.statsUsecase.GetSLAMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetTagStats godoc
// @Summary Get tag statistics
// @Description Retrieve incident count statistics by tag
// @Tags stats
// @Accept json
// @Produce json
// @Success 200 {object} []usecase.TagStats
// @Failure 500 {object} map[string]string
// @Router /api/stats/tags [get]
// @Security BearerAuth
func (h *StatsHandler) GetTagStats(c *gin.Context) {
	tagStats, err := h.statsUsecase.GetTagStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tag_stats": tagStats})
}

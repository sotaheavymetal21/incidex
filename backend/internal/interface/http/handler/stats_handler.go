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

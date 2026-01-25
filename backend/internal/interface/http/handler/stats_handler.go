package handler

import (
	"incidex/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StatsHandler は統計情報関連の HTTP handler を提供します
type StatsHandler struct {
	statsUsecase *usecase.StatsUsecase
}

// NewStatsHandler は新しい StatsHandler を作成します
func NewStatsHandler(statsUsecase *usecase.StatsUsecase) *StatsHandler {
	return &StatsHandler{
		statsUsecase: statsUsecase,
	}
}

// GetDashboardStats godoc
// @Summary ダッシュボード統計を取得します
// @Description カウント、分布、トレンドを含むダッシュボード用の統計情報を取得します
// @Tags stats
// @Accept json
// @Produce json
// @Param period query string false "トレンドデータの期間 (daily, weekly, monthly)" default(daily)
// @Success 200 {object} usecase.DashboardStats
// @Failure 500 {object} map[string]string
// @Router /api/stats/dashboard [get]
// @Security BearerAuth
func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

	stats, err := h.statsUsecase.GetDashboardStats(period)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSLAMetrics godoc
// @Summary SLA パフォーマンスメトリクスを取得します
// @Description コンプライアンス率、MTTR、違反を含む SLA メトリクスを取得します
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetTagStats godoc
// @Summary タグ統計を取得します
// @Description タグ別のインシデント数統計を取得します
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tag_stats": tagStats})
}

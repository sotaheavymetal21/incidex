package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogHandler は監査ログ関連の HTTP handler を提供します
type AuditLogHandler struct {
	auditLogUsecase usecase.AuditLogUsecase
}

// NewAuditLogHandler は新しい AuditLogHandler を作成します
func NewAuditLogHandler(u usecase.AuditLogUsecase) *AuditLogHandler {
	return &AuditLogHandler{auditLogUsecase: u}
}

// GetAll godoc
// @Summary すべての監査ログを取得します
// @Description フィルタとページネーション付きですべての監査ログを取得します
// @Tags audit-logs
// @Accept json
// @Produce json
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "1ページあたりの件数" default(50)
// @Param user_id query int false "ユーザー ID フィルタ"
// @Param action query string false "アクションフィルタ"
// @Param resource_type query string false "リソースタイプフィルタ"
// @Param start_date query string false "開始日時（RFC3339 形式）"
// @Param end_date query string false "終了日時（RFC3339 形式）"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/audit-logs [get]
// @Security BearerAuth
func (h *AuditLogHandler) GetAll(c *gin.Context) {
	// クエリパラメータをパース
	filters := domain.AuditLogFilters{
		Page:  1,
		Limit: 50,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filters.Page = p
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	}

	if userID := c.Query("user_id"); userID != "" {
		if uid, err := strconv.ParseUint(userID, 10, 32); err == nil {
			id := uint(uid)
			filters.UserID = &id
		}
	}

	if action := c.Query("action"); action != "" {
		act := domain.AuditAction(action)
		filters.Action = &act
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters.ResourceType = &resourceType
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if sd, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters.StartDate = &sd
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if ed, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters.EndDate = &ed
		}
	}

	logs, total, err := h.auditLogUsecase.GetAll(c.Request.Context(), filters)
	if err != nil {
		HandleError(c, err)
		return
	}

	// ページネーション情報を計算
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"pagination": gin.H{
			"page":        filters.Page,
			"limit":       filters.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetByID godoc
// @Summary ID で監査ログを取得します
// @Description ID を指定して監査ログを取得します
// @Tags audit-logs
// @Accept json
// @Produce json
// @Param id path int true "監査ログ ID"
// @Success 200 {object} domain.AuditLog
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/audit-logs/{id} [get]
// @Security BearerAuth
func (h *AuditLogHandler) GetByID(c *gin.Context) {
	id, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	log, err := h.auditLogUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if log == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

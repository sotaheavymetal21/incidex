package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ActionItemHandler はアクションアイテム関連の HTTP handler を提供します
type ActionItemHandler struct {
	actionItemUsecase usecase.ActionItemUsecase
}

// NewActionItemHandler は新しい ActionItemHandler を作成します
func NewActionItemHandler(actionItemUsecase usecase.ActionItemUsecase) *ActionItemHandler {
	return &ActionItemHandler{
		actionItemUsecase: actionItemUsecase,
	}
}

// CreateActionItemRequest はアクションアイテム作成の request body を表します
type CreateActionItemRequest struct {
	PostMortemID uint    `json:"post_mortem_id" binding:"required"`
	Title        string  `json:"title" binding:"required,min=1,max=500"`
	Description  string  `json:"description" binding:"max=5000"`
	AssigneeID   *uint   `json:"assignee_id"`
	Priority     string  `json:"priority" binding:"required,oneof=high medium low"`
	DueDate      *string `json:"due_date"` // RFC3339 形式
	RelatedLinks string  `json:"related_links" binding:"max=2000"`
}

// UpdateActionItemRequest はアクションアイテム更新の request body を表します
type UpdateActionItemRequest struct {
	Title        string  `json:"title" binding:"required,min=1,max=500"`
	Description  string  `json:"description" binding:"max=5000"`
	AssigneeID   *uint   `json:"assignee_id"`
	Priority     string  `json:"priority" binding:"required,oneof=high medium low"`
	Status       string  `json:"status" binding:"required,oneof=pending in_progress completed"`
	DueDate      *string `json:"due_date"` // RFC3339 形式
	RelatedLinks string  `json:"related_links" binding:"max=2000"`
}

// Create godoc
// @Summary 新しいアクションアイテムを作成します
// @Description ポストモーテムに対して新しいアクションアイテムを作成します
// @Tags action-items
// @Accept json
// @Produce json
// @Param action_item body CreateActionItemRequest true "アクションアイテムデータ"
// @Success 201 {object} domain.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/action-items [post]
// @Security BearerAuth
func (h *ActionItemHandler) Create(c *gin.Context) {
	var req CreateActionItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 期限日をパース（指定されている場合）
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format (expected RFC3339)"})
			return
		}
		dueDate = &parsedDate
	}

	item, err := h.actionItemUsecase.CreateActionItem(
		c.Request.Context(),
		req.PostMortemID,
		req.Title,
		req.Description,
		req.AssigneeID,
		domain.Priority(req.Priority),
		dueDate,
		req.RelatedLinks,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

// GetByID godoc
// @Summary ID でアクションアイテムを取得します
// @Description ID を指定してアクションアイテムを取得します
// @Tags action-items
// @Accept json
// @Produce json
// @Param id path int true "アクションアイテム ID"
// @Success 200 {object} domain.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/action-items/{id} [get]
// @Security BearerAuth
func (h *ActionItemHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action item ID"})
		return
	}

	item, err := h.actionItemUsecase.GetActionItemByID(c.Request.Context(), uint(id))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetByPostMortemID godoc
// @Summary ポストモーテム ID でアクションアイテムを取得します
// @Description ポストモーテムに紐づくすべてのアクションアイテムを取得します
// @Tags action-items
// @Accept json
// @Produce json
// @Param postMortemId path int true "ポストモーテム ID"
// @Success 200 {array} domain.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/action-items/post-mortem/{postMortemId} [get]
// @Security BearerAuth
func (h *ActionItemHandler) GetByPostMortemID(c *gin.Context) {
	idStr := c.Param("id")
	postMortemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post-mortem ID"})
		return
	}

	items, err := h.actionItemUsecase.GetActionItemsByPostMortemID(c.Request.Context(), uint(postMortemID))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetAll godoc
// @Summary すべてのアクションアイテムを取得します
// @Description フィルタとページネーション付きですべてのアクションアイテムを取得します
// @Tags action-items
// @Accept json
// @Produce json
// @Param status query string false "ステータスフィルタ"
// @Param priority query string false "優先度フィルタ"
// @Param assignee_id query int false "担当者 ID フィルタ"
// @Param search query string false "検索クエリ"
// @Param sort_by query string false "ソートフィールド"
// @Param order query string false "ソート順 (asc/desc)"
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "1ページあたりの件数" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/action-items [get]
// @Security BearerAuth
func (h *ActionItemHandler) GetAll(c *gin.Context) {
	filters := domain.ActionItemFilters{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Search:   c.Query("search"),
		SortBy:   c.Query("sort_by"),
		Order:    c.Query("order"),
	}

	if assigneeIDStr := c.Query("assignee_id"); assigneeIDStr != "" {
		assigneeID, err := strconv.ParseUint(assigneeIDStr, 10, 32)
		if err == nil {
			filters.AssigneeID = uint(assigneeID)
		}
	}

	pagination := domain.Pagination{
		Page:  1,
		Limit: 20,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			pagination.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			pagination.Limit = limit
		}
	}

	items, paginationResult, err := h.actionItemUsecase.GetAllActionItems(
		c.Request.Context(),
		filters,
		pagination,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"action_items": items,
		"pagination":   paginationResult,
	})
}

// Update godoc
// @Summary アクションアイテムを更新します
// @Description アクションアイテムを更新します
// @Tags action-items
// @Accept json
// @Produce json
// @Param id path int true "アクションアイテム ID"
// @Param action_item body UpdateActionItemRequest true "アクションアイテムデータ"
// @Success 200 {object} domain.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/action-items/{id} [put]
// @Security BearerAuth
func (h *ActionItemHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action item ID"})
		return
	}

	var req UpdateActionItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 期限日をパース（指定されている場合）
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format (expected RFC3339)"})
			return
		}
		dueDate = &parsedDate
	}

	item, err := h.actionItemUsecase.UpdateActionItem(
		c.Request.Context(),
		uint(id),
		req.Title,
		req.Description,
		req.AssigneeID,
		domain.Priority(req.Priority),
		domain.ActionStatus(req.Status),
		dueDate,
		req.RelatedLinks,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// Delete godoc
// @Summary アクションアイテムを削除します
// @Description アクションアイテムを削除します（管理者専用）
// @Tags action-items
// @Accept json
// @Produce json
// @Param id path int true "アクションアイテム ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/action-items/{id} [delete]
// @Security BearerAuth
func (h *ActionItemHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action item ID"})
		return
	}

	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	userRoleTyped, ok := userRole.(domain.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user role type"})
		return
	}

	err = h.actionItemUsecase.DeleteActionItem(
		c.Request.Context(),
		userRoleTyped,
		uint(id),
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Action item deleted successfully"})
}

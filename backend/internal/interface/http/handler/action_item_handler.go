package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ActionItemHandler struct {
	actionItemUsecase usecase.ActionItemUsecase
}

func NewActionItemHandler(actionItemUsecase usecase.ActionItemUsecase) *ActionItemHandler {
	return &ActionItemHandler{
		actionItemUsecase: actionItemUsecase,
	}
}

type CreateActionItemRequest struct {
	PostMortemID uint    `json:"post_mortem_id" binding:"required"`
	Title        string  `json:"title" binding:"required,max=500"`
	Description  string  `json:"description"`
	AssigneeID   *uint   `json:"assignee_id"`
	Priority     string  `json:"priority" binding:"required,oneof=high medium low"`
	DueDate      *string `json:"due_date"` // RFC3339 format
	RelatedLinks string  `json:"related_links"`
}

type UpdateActionItemRequest struct {
	Title        string  `json:"title" binding:"required,max=500"`
	Description  string  `json:"description"`
	AssigneeID   *uint   `json:"assignee_id"`
	Priority     string  `json:"priority" binding:"required,oneof=high medium low"`
	Status       string  `json:"status" binding:"required,oneof=pending in_progress completed"`
	DueDate      *string `json:"due_date"` // RFC3339 format
	RelatedLinks string  `json:"related_links"`
}

// Create godoc
// @Summary Create a new action item
// @Description Create a new action item for a post-mortem
// @Tags action-items
// @Accept json
// @Produce json
// @Param action_item body CreateActionItemRequest true "Action item data"
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

	// Parse due date if provided
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// GetByID godoc
// @Summary Get action item by ID
// @Description Get an action item by ID
// @Tags action-items
// @Accept json
// @Produce json
// @Param id path int true "Action item ID"
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
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetByPostMortemID godoc
// @Summary Get action items by post-mortem ID
// @Description Get all action items for a post-mortem
// @Tags action-items
// @Accept json
// @Produce json
// @Param postMortemId path int true "Post-mortem ID"
// @Success 200 {array} domain.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/action-items/post-mortem/{postMortemId} [get]
// @Security BearerAuth
func (h *ActionItemHandler) GetByPostMortemID(c *gin.Context) {
	idStr := c.Param("postMortemId")
	postMortemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post-mortem ID"})
		return
	}

	items, err := h.actionItemUsecase.GetActionItemsByPostMortemID(c.Request.Context(), uint(postMortemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetAll godoc
// @Summary Get all action items
// @Description Get all action items with filters and pagination
// @Tags action-items
// @Accept json
// @Produce json
// @Param status query string false "Status filter"
// @Param priority query string false "Priority filter"
// @Param assignee_id query int false "Assignee ID filter"
// @Param search query string false "Search query"
// @Param sort_by query string false "Sort by field"
// @Param order query string false "Sort order (asc/desc)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"action_items": items,
		"pagination":   paginationResult,
	})
}

// Update godoc
// @Summary Update an action item
// @Description Update an action item
// @Tags action-items
// @Accept json
// @Produce json
// @Param id path int true "Action item ID"
// @Param action_item body UpdateActionItemRequest true "Action item data"
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

	// Parse due date if provided
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// Delete godoc
// @Summary Delete an action item
// @Description Delete an action item (admin only)
// @Tags action-items
// @Accept json
// @Produce json
// @Param id path int true "Action item ID"
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

	userRoleStr, ok := userRole.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user role type"})
		return
	}

	err = h.actionItemUsecase.DeleteActionItem(
		c.Request.Context(),
		domain.Role(userRoleStr),
		uint(id),
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Action item deleted successfully"})
}

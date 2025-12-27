package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostMortemHandler struct {
	postMortemUsecase usecase.PostMortemUsecase
}

func NewPostMortemHandler(postMortemUsecase usecase.PostMortemUsecase) *PostMortemHandler {
	return &PostMortemHandler{
		postMortemUsecase: postMortemUsecase,
	}
}

type CreatePostMortemRequest struct {
	IncidentID       uint                       `json:"incident_id" binding:"required"`
	RootCause        string                     `json:"root_cause"`
	ImpactAnalysis   string                     `json:"impact_analysis"`
	WhatWentWell     string                     `json:"what_went_well"`
	WhatWentWrong    string                     `json:"what_went_wrong"`
	LessonsLearned   string                     `json:"lessons_learned"`
	FiveWhysAnalysis *domain.FiveWhysAnalysis   `json:"five_whys_analysis"`
}

type UpdatePostMortemRequest struct {
	RootCause        string                     `json:"root_cause"`
	ImpactAnalysis   string                     `json:"impact_analysis"`
	WhatWentWell     string                     `json:"what_went_well"`
	WhatWentWrong    string                     `json:"what_went_wrong"`
	LessonsLearned   string                     `json:"lessons_learned"`
	FiveWhysAnalysis *domain.FiveWhysAnalysis   `json:"five_whys_analysis"`
}

// Create godoc
// @Summary Create a new post-mortem
// @Description Create a new post-mortem for an incident
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param post_mortem body CreatePostMortemRequest true "Post-mortem data"
// @Success 201 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems [post]
// @Security BearerAuth
func (h *PostMortemHandler) Create(c *gin.Context) {
	var req CreatePostMortemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	pm, err := h.postMortemUsecase.CreatePostMortem(
		c.Request.Context(),
		userIDUint,
		req.IncidentID,
		req.RootCause,
		req.ImpactAnalysis,
		req.WhatWentWell,
		req.WhatWentWrong,
		req.LessonsLearned,
		req.FiveWhysAnalysis,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pm)
}

// GetByID godoc
// @Summary Get post-mortem by ID
// @Description Get a post-mortem by ID
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "Post-mortem ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/post-mortems/{id} [get]
// @Security BearerAuth
func (h *PostMortemHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post-mortem ID"})
		return
	}

	pm, err := h.postMortemUsecase.GetPostMortemByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pm)
}

// GetByIncidentID godoc
// @Summary Get post-mortem by incident ID
// @Description Get a post-mortem by incident ID
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param incidentId path int true "Incident ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/post-mortems/incident/{incidentId} [get]
// @Security BearerAuth
func (h *PostMortemHandler) GetByIncidentID(c *gin.Context) {
	idStr := c.Param("id")
	incidentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	pm, err := h.postMortemUsecase.GetPostMortemByIncidentID(c.Request.Context(), uint(incidentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pm)
}

// GetAll godoc
// @Summary Get all post-mortems
// @Description Get all post-mortems with filters and pagination
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param status query string false "Status filter"
// @Param author_id query int false "Author ID filter"
// @Param search query string false "Search query"
// @Param sort_by query string false "Sort by field"
// @Param order query string false "Sort order (asc/desc)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems [get]
// @Security BearerAuth
func (h *PostMortemHandler) GetAll(c *gin.Context) {
	filters := domain.PostMortemFilters{
		Status:   c.Query("status"),
		Search:   c.Query("search"),
		SortBy:   c.Query("sort_by"),
		Order:    c.Query("order"),
	}

	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		authorID, err := strconv.ParseUint(authorIDStr, 10, 32)
		if err == nil {
			filters.AuthorID = uint(authorID)
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

	postMortems, paginationResult, err := h.postMortemUsecase.GetAllPostMortems(
		c.Request.Context(),
		filters,
		pagination,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"post_mortems": postMortems,
		"pagination":   paginationResult,
	})
}

// Update godoc
// @Summary Update a post-mortem
// @Description Update a post-mortem (draft only)
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "Post-mortem ID"
// @Param post_mortem body UpdatePostMortemRequest true "Post-mortem data"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id} [put]
// @Security BearerAuth
func (h *PostMortemHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post-mortem ID"})
		return
	}

	var req UpdatePostMortemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
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

	pm, err := h.postMortemUsecase.UpdatePostMortem(
		c.Request.Context(),
		userIDUint,
		domain.Role(userRoleStr),
		uint(id),
		req.RootCause,
		req.ImpactAnalysis,
		req.WhatWentWell,
		req.WhatWentWrong,
		req.LessonsLearned,
		req.FiveWhysAnalysis,
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pm)
}

// Publish godoc
// @Summary Publish a post-mortem
// @Description Publish a post-mortem (makes it immutable)
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "Post-mortem ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id}/publish [post]
// @Security BearerAuth
func (h *PostMortemHandler) Publish(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post-mortem ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
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

	pm, err := h.postMortemUsecase.PublishPostMortem(
		c.Request.Context(),
		userIDUint,
		domain.Role(userRoleStr),
		uint(id),
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pm)
}

// Unpublish godoc
// @Summary Unpublish a post-mortem
// @Description Unpublish a post-mortem (revert to draft status for editing)
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "Post-mortem ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id}/unpublish [post]
// @Security BearerAuth
func (h *PostMortemHandler) Unpublish(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post-mortem ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
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

	pm, err := h.postMortemUsecase.UnpublishPostMortem(
		c.Request.Context(),
		userIDUint,
		domain.Role(userRoleStr),
		uint(id),
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pm)
}

// Delete godoc
// @Summary Delete a post-mortem
// @Description Delete a post-mortem (admin only)
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "Post-mortem ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id} [delete]
// @Security BearerAuth
func (h *PostMortemHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post-mortem ID"})
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

	err = h.postMortemUsecase.DeletePostMortem(
		c.Request.Context(),
		domain.Role(userRoleStr),
		uint(id),
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post-mortem deleted successfully"})
}

// GenerateAISuggestion godoc
// @Summary Generate AI root cause suggestion
// @Description Generate AI root cause suggestion for an incident
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param incidentId path int true "Incident ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/incident/{incidentId}/ai-suggestion [post]
// @Security BearerAuth
func (h *PostMortemHandler) GenerateAISuggestion(c *gin.Context) {
	idStr := c.Param("id")
	incidentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	suggestion, err := h.postMortemUsecase.GenerateAIRootCauseSuggestion(
		c.Request.Context(),
		uint(incidentID),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestion": suggestion})
}

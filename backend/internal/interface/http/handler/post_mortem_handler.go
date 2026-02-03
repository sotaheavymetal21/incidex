package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PostMortemHandler はポストモーテム関連の HTTP handler を提供します
type PostMortemHandler struct {
	postMortemUsecase usecase.PostMortemUsecase
}

// NewPostMortemHandler は新しい PostMortemHandler を作成します
func NewPostMortemHandler(postMortemUsecase usecase.PostMortemUsecase) *PostMortemHandler {
	return &PostMortemHandler{
		postMortemUsecase: postMortemUsecase,
	}
}

// CreatePostMortemRequest はポストモーテム作成の request body を表します
type CreatePostMortemRequest struct {
	IncidentID       uint                     `json:"incident_id" binding:"required"`
	RootCause        string                   `json:"root_cause" binding:"max=10000"`
	ImpactAnalysis   string                   `json:"impact_analysis" binding:"max=10000"`
	WhatWentWell     string                   `json:"what_went_well" binding:"max=10000"`
	WhatWentWrong    string                   `json:"what_went_wrong" binding:"max=10000"`
	LessonsLearned   string                   `json:"lessons_learned" binding:"max=10000"`
	FiveWhysAnalysis *domain.FiveWhysAnalysis `json:"five_whys_analysis"`
}

// UpdatePostMortemRequest はポストモーテム更新の request body を表します
type UpdatePostMortemRequest struct {
	RootCause        string                   `json:"root_cause" binding:"max=10000"`
	ImpactAnalysis   string                   `json:"impact_analysis" binding:"max=10000"`
	WhatWentWell     string                   `json:"what_went_well" binding:"max=10000"`
	WhatWentWrong    string                   `json:"what_went_wrong" binding:"max=10000"`
	LessonsLearned   string                   `json:"lessons_learned" binding:"max=10000"`
	FiveWhysAnalysis *domain.FiveWhysAnalysis `json:"five_whys_analysis"`
}

// Create godoc
// @Summary 新しいポストモーテムを作成します
// @Description インシデントに対して新しいポストモーテムを作成します
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param post_mortem body CreatePostMortemRequest true "ポストモーテムデータ"
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

	userIDUint, err := GetUserIDFromContext(c)
	if err != nil {
		HandleError(c, err)
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, pm)
}

// GetByID godoc
// @Summary ID でポストモーテムを取得します
// @Description ID を指定してポストモーテムを取得します
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "ポストモーテム ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/post-mortems/{id} [get]
// @Security BearerAuth
func (h *PostMortemHandler) GetByID(c *gin.Context) {
	id, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	pm, err := h.postMortemUsecase.GetPostMortemByID(c.Request.Context(), id)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pm)
}

// GetByIncidentID godoc
// @Summary インシデント ID でポストモーテムを取得します
// @Description インシデント ID を指定してポストモーテムを取得します
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param incidentId path int true "インシデント ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/post-mortems/incident/{incidentId} [get]
// @Security BearerAuth
func (h *PostMortemHandler) GetByIncidentID(c *gin.Context) {
	incidentID, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	pm, err := h.postMortemUsecase.GetPostMortemByIncidentID(c.Request.Context(), incidentID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pm)
}

// GetAll godoc
// @Summary すべてのポストモーテムを取得します
// @Description フィルタとページネーション付きですべてのポストモーテムを取得します
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param status query string false "ステータスフィルタ"
// @Param author_id query int false "作成者 ID フィルタ"
// @Param search query string false "検索クエリ"
// @Param sort_by query string false "ソートフィールド"
// @Param order query string false "ソート順 (asc/desc)"
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "1ページあたりの件数" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems [get]
// @Security BearerAuth
func (h *PostMortemHandler) GetAll(c *gin.Context) {
	filters := domain.PostMortemFilters{
		Status: c.Query("status"),
		Search: c.Query("search"),
		SortBy: c.Query("sort_by"),
		Order:  c.Query("order"),
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"post_mortems": postMortems,
		"pagination":   paginationResult,
	})
}

// Update godoc
// @Summary ポストモーテムを更新します
// @Description ポストモーテムを更新します（下書きのみ）
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "ポストモーテム ID"
// @Param post_mortem body UpdatePostMortemRequest true "ポストモーテムデータ"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id} [put]
// @Security BearerAuth
func (h *PostMortemHandler) Update(c *gin.Context) {
	id, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	var req UpdatePostMortemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDUint, err := GetUserIDFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	userRoleTyped, err := GetUserRoleFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	pm, err := h.postMortemUsecase.UpdatePostMortem(
		c.Request.Context(),
		userIDUint,
		userRoleTyped,
		id,
		req.RootCause,
		req.ImpactAnalysis,
		req.WhatWentWell,
		req.WhatWentWrong,
		req.LessonsLearned,
		req.FiveWhysAnalysis,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pm)
}

// Publish godoc
// @Summary ポストモーテムを公開します
// @Description ポストモーテムを公開します（変更不可になります）
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "ポストモーテム ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id}/publish [post]
// @Security BearerAuth
func (h *PostMortemHandler) Publish(c *gin.Context) {
	id, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	userIDUint, err := GetUserIDFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	userRoleTyped, err := GetUserRoleFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	pm, err := h.postMortemUsecase.PublishPostMortem(
		c.Request.Context(),
		userIDUint,
		userRoleTyped,
		id,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pm)
}

// Unpublish godoc
// @Summary ポストモーテムを非公開にします
// @Description ポストモーテムを非公開にします（編集可能な下書き状態に戻します）
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "ポストモーテム ID"
// @Success 200 {object} domain.PostMortem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id}/unpublish [post]
// @Security BearerAuth
func (h *PostMortemHandler) Unpublish(c *gin.Context) {
	id, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	userIDUint, err := GetUserIDFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	userRoleTyped, err := GetUserRoleFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	pm, err := h.postMortemUsecase.UnpublishPostMortem(
		c.Request.Context(),
		userIDUint,
		userRoleTyped,
		id,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pm)
}

// Delete godoc
// @Summary ポストモーテムを削除します
// @Description ポストモーテムを削除します（管理者専用）
// @Tags post-mortems
// @Accept json
// @Produce json
// @Param id path int true "ポストモーテム ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/post-mortems/{id} [delete]
// @Security BearerAuth
func (h *PostMortemHandler) Delete(c *gin.Context) {
	id, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	userRoleTyped, err := GetUserRoleFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	err = h.postMortemUsecase.DeletePostMortem(
		c.Request.Context(),
		userRoleTyped,
		id,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post-mortem deleted successfully"})
}

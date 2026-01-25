package handler

import (
	"incidex/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TagHandler はタグ関連の HTTP handler を提供します
type TagHandler struct {
	tagUsecase usecase.TagUsecase
}

// NewTagHandler は新しい TagHandler を作成します
func NewTagHandler(u usecase.TagUsecase) *TagHandler {
	return &TagHandler{tagUsecase: u}
}

// CreateTagRequest はタグ作成の request body を表します
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// UpdateTagRequest はタグ更新の request body を表します
type UpdateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// Create は新しいタグを作成します
func (h *TagHandler) Create(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	tag, err := h.tagUsecase.CreateTag(c.Request.Context(), req.Name, req.Color)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// GetAll はすべてのタグを取得します
func (h *TagHandler) GetAll(c *gin.Context) {
	tags, err := h.tagUsecase.GetAllTags(c.Request.Context())
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tags)
}

// Update は既存のタグを更新します
func (h *TagHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		HandleValidationError(c, err)
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	tag, err := h.tagUsecase.UpdateTag(c.Request.Context(), uint(id), req.Name, req.Color)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tag)
}

// Delete は指定されたタグを削除します
func (h *TagHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		HandleValidationError(c, err)
		return
	}

	if err := h.tagUsecase.DeleteTag(c.Request.Context(), uint(id)); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

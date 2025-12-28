package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type IncidentTemplateHandler struct {
	templateUsecase *usecase.IncidentTemplateUsecase
}

func NewIncidentTemplateHandler(templateUsecase *usecase.IncidentTemplateUsecase) *IncidentTemplateHandler {
	return &IncidentTemplateHandler{
		templateUsecase: templateUsecase,
	}
}

// CreateTemplateRequest represents the request body for creating a template
type CreateTemplateRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Title       string          `json:"title" binding:"required"`
	Content     string          `json:"content" binding:"required"`
	Severity    domain.Severity `json:"severity" binding:"required"`
	ImpactScope string          `json:"impact_scope"`
	IsPublic    bool            `json:"is_public"`
	TagIDs      []uint          `json:"tag_ids"`
}

// UpdateTemplateRequest represents the request body for updating a template
type UpdateTemplateRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Title       string          `json:"title" binding:"required"`
	Content     string          `json:"content" binding:"required"`
	Severity    domain.Severity `json:"severity" binding:"required"`
	ImpactScope string          `json:"impact_scope"`
	IsPublic    bool            `json:"is_public"`
	TagIDs      []uint          `json:"tag_ids"`
}

// CreateIncidentFromTemplateRequest represents the request body for creating an incident from a template
type CreateIncidentFromTemplateRequest struct {
	TemplateID uint   `json:"template_id" binding:"required"`
	AssigneeID *uint  `json:"assignee_id"`
	DetectedAt string `json:"detected_at"`
}

// Create godoc
// @Summary Create a new incident template
// @Tags templates
// @Accept json
// @Produce json
// @Param template body CreateTemplateRequest true "Template details"
// @Success 201 {object} domain.IncidentTemplate
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /templates [post]
func (h *IncidentTemplateHandler) Create(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateUsecase.CreateTemplate(
		c.Request.Context(),
		userID,
		req.Name,
		req.Description,
		req.Title,
		req.Content,
		req.Severity,
		req.ImpactScope,
		req.IsPublic,
		req.TagIDs,
	)

	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetAll godoc
// @Summary Get all templates accessible by the user
// @Tags templates
// @Accept json
// @Produce json
// @Success 200 {array} domain.IncidentTemplate
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /templates [get]
func (h *IncidentTemplateHandler) GetAll(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	templates, err := h.templateUsecase.GetAllTemplates(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, templates)
}

// GetByID godoc
// @Summary Get a template by ID
// @Tags templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} domain.IncidentTemplate
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /templates/:id [get]
func (h *IncidentTemplateHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.templateUsecase.GetTemplateByID(c.Request.Context(), uint(id))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

// Update godoc
// @Summary Update a template
// @Tags templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param template body UpdateTemplateRequest true "Template details"
// @Success 200 {object} domain.IncidentTemplate
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /templates/:id [put]
func (h *IncidentTemplateHandler) Update(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role, ok := roleValue.(domain.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user role"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateUsecase.UpdateTemplate(
		c.Request.Context(),
		userID,
		role,
		uint(id),
		req.Name,
		req.Description,
		req.Title,
		req.Content,
		req.Severity,
		req.ImpactScope,
		req.IsPublic,
		req.TagIDs,
	)

	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

// Delete godoc
// @Summary Delete a template
// @Tags templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /templates/:id [delete]
func (h *IncidentTemplateHandler) Delete(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role, ok := roleValue.(domain.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user role"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.templateUsecase.DeleteTemplate(c.Request.Context(), userID, role, uint(id)); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

// CreateIncidentFromTemplate godoc
// @Summary Create an incident from a template
// @Tags templates
// @Accept json
// @Produce json
// @Param request body CreateIncidentFromTemplateRequest true "Request details"
// @Success 201 {object} domain.Incident
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /templates/create-incident [post]
func (h *IncidentTemplateHandler) CreateIncidentFromTemplate(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	var req CreateIncidentFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse detected_at
	var detectedAt time.Time
	if req.DetectedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.DetectedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detected_at format"})
			return
		}
		detectedAt = parsedTime
	} else {
		detectedAt = time.Now()
	}

	incident, err := h.templateUsecase.CreateIncidentFromTemplate(
		c.Request.Context(),
		req.TemplateID,
		userID,
		req.AssigneeID,
		detectedAt,
	)

	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, incident)
}

package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type IncidentHandler struct {
	incidentUsecase usecase.IncidentUsecase
}

func NewIncidentHandler(u usecase.IncidentUsecase) *IncidentHandler {
	return &IncidentHandler{incidentUsecase: u}
}

type CreateIncidentRequest struct {
	Title       string   `json:"title" binding:"required,max=500"`
	Description string   `json:"description" binding:"required"`
	Severity    string   `json:"severity" binding:"required,oneof=critical high medium low"`
	Status      string   `json:"status" binding:"required,oneof=open investigating resolved closed"`
	ImpactScope string   `json:"impact_scope"`
	DetectedAt  string   `json:"detected_at" binding:"required"`
	AssigneeID  *uint    `json:"assignee_id"`
	TagIDs      []uint   `json:"tag_ids"`
}

type UpdateIncidentRequest struct {
	Title       string   `json:"title" binding:"required,max=500"`
	Description string   `json:"description" binding:"required"`
	Severity    string   `json:"severity" binding:"required,oneof=critical high medium low"`
	Status      string   `json:"status" binding:"required,oneof=open investigating resolved closed"`
	ImpactScope string   `json:"impact_scope"`
	DetectedAt  string   `json:"detected_at" binding:"required"`
	ResolvedAt  *string  `json:"resolved_at"`
	AssigneeID  *uint    `json:"assignee_id"`
	TagIDs      []uint   `json:"tag_ids"`
}

type IncidentListResponse struct {
	Incidents  []*domain.Incident       `json:"incidents"`
	Pagination *domain.PaginationResult `json:"pagination"`
}

func (h *IncidentHandler) Create(c *gin.Context) {
	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT context (set by middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDFloat, ok := userID.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	// Parse detected_at
	detectedAt, err := time.Parse(time.RFC3339, req.DetectedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detected_at format (expected RFC3339)"})
		return
	}

	incident, err := h.incidentUsecase.CreateIncident(
		c.Request.Context(),
		uint(userIDFloat),
		req.Title,
		req.Description,
		domain.Severity(req.Severity),
		domain.Status(req.Status),
		req.ImpactScope,
		detectedAt,
		req.AssigneeID,
		req.TagIDs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, incident)
}

func (h *IncidentHandler) GetAll(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	severity := c.Query("severity")
	status := c.Query("status")
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// Parse tag_ids (comma-separated)
	var tagIDs []uint
	tagIDsStr := c.Query("tag_ids")
	if tagIDsStr != "" {
		for _, idStr := range strings.Split(tagIDsStr, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32)
			if err == nil {
				tagIDs = append(tagIDs, uint(id))
			}
		}
	}

	filters := domain.IncidentFilters{
		Severity: severity,
		Status:   status,
		TagIDs:   tagIDs,
		Search:   search,
		SortBy:   sortBy,
		Order:    order,
	}

	pagination := domain.Pagination{
		Page:  page,
		Limit: limit,
	}

	incidents, paginationResult, err := h.incidentUsecase.GetAllIncidents(c.Request.Context(), filters, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := IncidentListResponse{
		Incidents:  incidents,
		Pagination: paginationResult,
	}

	c.JSON(http.StatusOK, response)
}

func (h *IncidentHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	incident, err := h.incidentUsecase.GetIncidentByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
		return
	}

	c.JSON(http.StatusOK, incident)
}

func (h *IncidentHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID and role from JWT context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDFloat, ok := userID.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Parse detected_at
	detectedAt, err := time.Parse(time.RFC3339, req.DetectedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detected_at format (expected RFC3339)"})
		return
	}

	// Parse resolved_at if provided
	var resolvedAt *time.Time
	if req.ResolvedAt != nil && *req.ResolvedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ResolvedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resolved_at format (expected RFC3339)"})
			return
		}
		resolvedAt = &parsed
	}

	incident, err := h.incidentUsecase.UpdateIncident(
		c.Request.Context(),
		uint(userIDFloat),
		domain.Role(role.(string)),
		uint(id),
		req.Title,
		req.Description,
		domain.Severity(req.Severity),
		domain.Status(req.Status),
		req.ImpactScope,
		detectedAt,
		resolvedAt,
		req.AssigneeID,
		req.TagIDs,
	)
	if err != nil {
		if err.Error() == "permission denied: you can only edit your own incidents" || err.Error() == "permission denied: viewers cannot edit incidents" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, incident)
}

func (h *IncidentHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Get user role from JWT context
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	if err := h.incidentUsecase.DeleteIncident(c.Request.Context(), domain.Role(role.(string)), uint(id)); err != nil {
		if err.Error() == "permission denied: only admins can delete incidents" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident deleted successfully"})
}

func (h *IncidentHandler) RegenerateSummary(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Get user role from JWT context (editors and admins can regenerate)
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	userRole := domain.Role(role.(string))
	if userRole == domain.RoleViewer {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: viewers cannot regenerate summaries"})
		return
	}

	summary, err := h.incidentUsecase.RegenerateSummary(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary":     summary,
		"generated_at": time.Now().Format(time.RFC3339),
	})
}

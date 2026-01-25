package handler

import (
	"incidex/internal/domain"
	"incidex/internal/interface/http/validator"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// IncidentHandler はインシデント関連の HTTP handler を提供します
type IncidentHandler struct {
	incidentUsecase usecase.IncidentUsecase
}

// NewIncidentHandler は新しい IncidentHandler を作成します
func NewIncidentHandler(u usecase.IncidentUsecase) *IncidentHandler {
	return &IncidentHandler{incidentUsecase: u}
}

// CreateIncidentRequest はインシデント作成の request body を表します
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

// UpdateIncidentRequest はインシデント更新の request body を表します
type UpdateIncidentRequest struct {
	Title       string  `json:"title" binding:"required,max=500"`
	Description string  `json:"description" binding:"required"`
	Severity    string  `json:"severity" binding:"required,oneof=critical high medium low"`
	Status      string  `json:"status" binding:"required,oneof=open investigating resolved closed"`
	ImpactScope string  `json:"impact_scope"`
	DetectedAt  string  `json:"detected_at" binding:"required"`
	AssigneeID  *uint   `json:"assignee_id"`
	TagIDs      []uint  `json:"tag_ids"`
}

// IncidentListResponse はインシデント一覧の response を表します
type IncidentListResponse struct {
	Incidents  []*domain.Incident       `json:"incidents"`
	Pagination *domain.PaginationResult `json:"pagination"`
}

// Create は新しいインシデントを作成します
func (h *IncidentHandler) Create(c *gin.Context) {
	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// カスタムバリデーション
	if err := validator.ValidateTitle(req.Title); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateDescription(req.Description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateImpactScope(req.ImpactScope); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateSeverity(req.Severity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateStatus(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// JWT context からユーザー ID を取得（middleware で設定済み）
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	// detected_at をパース
	detectedAt, err := time.Parse(time.RFC3339, req.DetectedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detected_at format (expected RFC3339)"})
		return
	}

	incident, err := h.incidentUsecase.CreateIncident(
		c.Request.Context(),
		userID,
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, incident)
}

// GetAll はすべてのインシデントを取得します
func (h *IncidentHandler) GetAll(c *gin.Context) {
	// クエリパラメータをパース
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	severity := c.Query("severity")
	status := c.Query("status")
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// tag_ids をパース（カンマ区切り）
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

	// assigned_to_id をパース
	var assignedToID *uint
	if assignedToIDStr := c.Query("assigned_to_id"); assignedToIDStr != "" {
		id, err := strconv.ParseUint(assignedToIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			assignedToID = &uid
		}
	}

	filters := domain.IncidentFilters{
		Severity:     severity,
		Status:       status,
		TagIDs:       tagIDs,
		Search:       search,
		SortBy:       sortBy,
		Order:        order,
		AssignedToID: assignedToID,
	}

	pagination := domain.Pagination{
		Page:  page,
		Limit: limit,
	}

	incidents, paginationResult, err := h.incidentUsecase.GetAllIncidents(c.Request.Context(), filters, pagination)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := IncidentListResponse{
		Incidents:  incidents,
		Pagination: paginationResult,
	}

	c.JSON(http.StatusOK, response)
}

// GetByID は指定された ID のインシデントを取得します
func (h *IncidentHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	incident, err := h.incidentUsecase.GetIncidentByID(c.Request.Context(), uint(id))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, incident)
}

// Update は既存のインシデントを更新します
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

	// JWT context からユーザー ID とロールを取得
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	userRole, ok := role.(domain.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user role type"})
		return
	}

	// detected_at をパース
	detectedAt, err := time.Parse(time.RFC3339, req.DetectedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detected_at format (expected RFC3339)"})
		return
	}

	incident, err := h.incidentUsecase.UpdateIncident(
		c.Request.Context(),
		userIDUint,
		userRole,
		uint(id),
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, incident)
}

// Delete は指定されたインシデントを削除します
func (h *IncidentHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// JWT context からユーザーロールを取得
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	userRole, ok := role.(domain.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user role type"})
		return
	}

	if err := h.incidentUsecase.DeleteIncident(c.Request.Context(), userRole, uint(id)); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident deleted successfully"})
}

// AssignIncidentRequest はインシデント担当者割り当ての request body を表します
type AssignIncidentRequest struct{
	AssigneeID *uint `json:"assignee_id"`
}

// AssignIncident はインシデントに担当者を割り当てます
func (h *IncidentHandler) AssignIncident(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}

	var req AssignIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// context からユーザー ID を取得
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	incident, err := h.incidentUsecase.AssignIncident(c.Request.Context(), userID, uint(id), req.AssigneeID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, incident)
}

package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type IncidentActivityHandler struct {
	activityUsecase *usecase.IncidentActivityUsecase
}

func NewIncidentActivityHandler(activityUsecase *usecase.IncidentActivityUsecase) *IncidentActivityHandler {
	return &IncidentActivityHandler{
		activityUsecase: activityUsecase,
	}
}

// AddComment godoc
// @Summary Add a comment to an incident
// @Description Add a comment to an incident
// @Tags incident-activities
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param comment body AddCommentRequest true "Comment"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/incidents/{id}/comments [post]
// @Security BearerAuth
func (h *IncidentActivityHandler) AddComment(c *gin.Context) {
	idStr := c.Param("id")
	incidentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	if err := h.activityUsecase.AddComment(uint(incidentID), userIDUint, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Comment added successfully"})
}

// GetActivities godoc
// @Summary Get activities for an incident
// @Description Get all activities (comments, status changes, etc.) for an incident
// @Tags incident-activities
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param limit query int false "Limit number of activities" default(50)
// @Success 200 {array} domain.IncidentActivity
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/incidents/{id}/activities [get]
// @Security BearerAuth
func (h *IncidentActivityHandler) GetActivities(c *gin.Context) {
	idStr := c.Param("id")
	incidentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	activities, err := h.activityUsecase.GetActivities(uint(incidentID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activities)
}

type AddCommentRequest struct {
	Comment string `json:"comment" binding:"required,min=1,max=5000"`
}

type AddTimelineEventRequest struct {
	EventType   string `json:"event_type" binding:"required,oneof=detected investigation_started root_cause_identified mitigation timeline_resolved other"`
	EventTime   string `json:"event_time" binding:"required"`
	Description string `json:"description" binding:"required,min=1,max=5000"`
}

// AddTimelineEvent godoc
// @Summary Add a timeline event to an incident
// @Description Add a timeline event (detected, investigation_started, root_cause_identified, mitigation, timeline_resolved, other) to an incident
// @Tags incident-activities
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param event body AddTimelineEventRequest true "Timeline Event"
// @Success 201 {object} domain.IncidentActivity
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/incidents/{id}/timeline [post]
// @Security BearerAuth
func (h *IncidentActivityHandler) AddTimelineEvent(c *gin.Context) {
	idStr := c.Param("id")
	incidentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	var req AddTimelineEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Parse event_time
	eventTime, err := time.Parse(time.RFC3339, req.EventTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_time format (expected RFC3339)"})
		return
	}

	activity, err := h.activityUsecase.AddTimelineEvent(
		uint(incidentID),
		userIDUint,
		domain.ActivityType(req.EventType),
		eventTime,
		req.Description,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, activity)
}

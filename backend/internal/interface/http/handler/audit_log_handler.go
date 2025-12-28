package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	auditLogUsecase usecase.AuditLogUsecase
}

func NewAuditLogHandler(u usecase.AuditLogUsecase) *AuditLogHandler {
	return &AuditLogHandler{auditLogUsecase: u}
}

func (h *AuditLogHandler) GetAll(c *gin.Context) {
	// Parse query parameters
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

	// Calculate pagination info
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

func (h *AuditLogHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audit log ID"})
		return
	}

	log, err := h.auditLogUsecase.GetByID(c.Request.Context(), uint(id))
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

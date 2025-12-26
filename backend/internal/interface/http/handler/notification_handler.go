package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationUsecase *usecase.NotificationUsecase
}

func NewNotificationHandler(notificationUsecase *usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{
		notificationUsecase: notificationUsecase,
	}
}

// GetMyNotificationSetting godoc
// @Summary Get my notification settings
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} domain.NotificationSetting
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /notifications/settings [get]
func (h *NotificationHandler) GetMyNotificationSetting(c *gin.Context) {
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

	setting, err := h.notificationUsecase.GetSettingByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// UpdateMyNotificationSetting godoc
// @Summary Update my notification settings
// @Tags notifications
// @Accept json
// @Produce json
// @Param setting body domain.NotificationSetting true "Notification Setting"
// @Success 200 {object} domain.NotificationSetting
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /notifications/settings [put]
func (h *NotificationHandler) UpdateMyNotificationSetting(c *gin.Context) {
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

	var req domain.NotificationSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.notificationUsecase.UpdateSetting(userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.notificationUsecase.GetSettingByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// GetUserNotificationSetting godoc
// @Summary Get user notification settings (Admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} domain.NotificationSetting
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /notifications/settings/:id [get]
func (h *NotificationHandler) GetUserNotificationSetting(c *gin.Context) {
	// Check if admin
	role, exists := c.Get("role")
	if !exists || role != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin only"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	setting, err := h.notificationUsecase.GetSettingByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

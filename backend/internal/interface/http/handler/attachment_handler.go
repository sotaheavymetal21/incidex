package handler

import (
	"incidex/internal/domain"
	"incidex/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AttachmentHandler struct {
	attachmentUsecase usecase.AttachmentUsecase
}

func NewAttachmentHandler(attachmentUsecase usecase.AttachmentUsecase) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentUsecase: attachmentUsecase,
	}
}

// Upload handles file upload for an incident
func (h *AttachmentHandler) Upload(c *gin.Context) {
	incidentIDStr := c.Param("id")
	incidentID, err := strconv.ParseUint(incidentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}

	// Get user from context (set by JWT middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Get file from form data
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Open the file
	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer fileReader.Close()

	// Get content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload attachment
	attachment, err := h.attachmentUsecase.UploadAttachment(
		c.Request.Context(),
		uint(incidentID),
		userID.(uint),
		file.Filename,
		file.Size,
		contentType,
		fileReader,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

// GetByIncidentID retrieves all attachments for an incident
func (h *AttachmentHandler) GetByIncidentID(c *gin.Context) {
	incidentIDStr := c.Param("id")
	incidentID, err := strconv.ParseUint(incidentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}

	attachments, err := h.attachmentUsecase.GetAttachmentsByIncidentID(c.Request.Context(), uint(incidentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attachments)
}

// Download handles file download
func (h *AttachmentHandler) Download(c *gin.Context) {
	attachmentIDStr := c.Param("attachmentId")
	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment ID"})
		return
	}

	// Get attachment metadata
	attachment, err := h.attachmentUsecase.GetAttachment(c.Request.Context(), uint(attachmentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}

	// Download file from storage
	reader, err := h.attachmentUsecase.DownloadAttachment(c.Request.Context(), uint(attachmentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Set headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+attachment.FileName)
	c.Header("Content-Type", attachment.MimeType)
	c.Header("Content-Length", strconv.FormatInt(attachment.FileSize, 10))

	// Stream the file
	c.DataFromReader(http.StatusOK, attachment.FileSize, attachment.MimeType, reader, nil)
}

// Delete handles attachment deletion
func (h *AttachmentHandler) Delete(c *gin.Context) {
	incidentIDStr := c.Param("id")
	_, err := strconv.ParseUint(incidentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}

	attachmentIDStr := c.Param("attachmentId")
	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment ID"})
		return
	}

	// Get user from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found"})
		return
	}

	// Delete attachment
	if err := h.attachmentUsecase.DeleteAttachment(
		c.Request.Context(),
		uint(attachmentID),
		userID.(uint),
		role.(domain.Role),
	); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "attachment deleted successfully"})
}

package handler

import (
	"fmt"
	"incidex/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AttachmentHandler は添付ファイル関連の HTTP handler を提供します
type AttachmentHandler struct {
	attachmentUsecase usecase.AttachmentUsecase
}

// NewAttachmentHandler は新しい AttachmentHandler を作成します
func NewAttachmentHandler(attachmentUsecase usecase.AttachmentUsecase) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentUsecase: attachmentUsecase,
	}
}

// Upload はインシデントへのファイルアップロードを処理します
func (h *AttachmentHandler) Upload(c *gin.Context) {
	incidentID, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	// context からユーザーを取得（JWT middleware で設定済み）
	userIDUint, err := GetUserIDFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	// フォームデータからファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// ファイルを開く
	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer fileReader.Close()

	// Content-Type を取得
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 添付ファイルをアップロード
	attachment, err := h.attachmentUsecase.UploadAttachment(
		c.Request.Context(),
		incidentID,
		userIDUint,
		file.Filename,
		file.Size,
		contentType,
		fileReader,
	)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

// GetByIncidentID はインシデントのすべての添付ファイルを取得します
func (h *AttachmentHandler) GetByIncidentID(c *gin.Context) {
	incidentID, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	attachments, err := h.attachmentUsecase.GetAttachmentsByIncidentID(c.Request.Context(), incidentID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, attachments)
}

// Download はファイルのダウンロードを処理します
func (h *AttachmentHandler) Download(c *gin.Context) {
	attachmentID, err := ParseIDParam(c, "attachmentId")
	if err != nil {
		HandleError(c, err)
		return
	}

	// 添付ファイルのメタデータを取得
	attachment, err := h.attachmentUsecase.GetAttachment(c.Request.Context(), attachmentID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// ストレージからファイルをダウンロード
	reader, err := h.attachmentUsecase.DownloadAttachment(c.Request.Context(), attachmentID)
	if err != nil {
		HandleError(c, err)
		return
	}
	defer reader.Close()

	// ダウンロード用の header を設定
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, attachment.FileName))
	c.Header("Content-Type", attachment.MimeType)
	c.Header("Content-Length", strconv.FormatInt(attachment.FileSize, 10))

	// ファイルをストリーミング
	c.DataFromReader(http.StatusOK, attachment.FileSize, attachment.MimeType, reader, nil)
}

// Delete は添付ファイルの削除を処理します
func (h *AttachmentHandler) Delete(c *gin.Context) {
	_, err := ParseIDParam(c, "id")
	if err != nil {
		HandleError(c, err)
		return
	}

	attachmentID, err := ParseIDParam(c, "attachmentId")
	if err != nil {
		HandleError(c, err)
		return
	}

	// context からユーザーを取得
	userIDUint, err := GetUserIDFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	role, err := GetUserRoleFromContext(c)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 添付ファイルを削除
	if err := h.attachmentUsecase.DeleteAttachment(
		c.Request.Context(),
		attachmentID,
		userIDUint,
		role,
	); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "attachment deleted successfully"})
}

package usecase

import (
	"context"
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/storage"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AttachmentUsecase interface {
	UploadAttachment(ctx context.Context, incidentID, userID uint, fileName string, fileSize int64, mimeType string, reader io.Reader) (*domain.Attachment, error)
	GetAttachmentsByIncidentID(ctx context.Context, incidentID uint) ([]*domain.Attachment, error)
	GetAttachment(ctx context.Context, id uint) (*domain.Attachment, error)
	DownloadAttachment(ctx context.Context, id uint) (io.ReadCloser, error)
	DeleteAttachment(ctx context.Context, id, userID uint, userRole domain.Role) error
}

type attachmentUsecase struct {
	attachmentRepo domain.AttachmentRepository
	incidentRepo   domain.IncidentRepository
	storage        *storage.MinIOStorage
}

func NewAttachmentUsecase(
	attachmentRepo domain.AttachmentRepository,
	incidentRepo domain.IncidentRepository,
	storage *storage.MinIOStorage,
) AttachmentUsecase {
	return &attachmentUsecase{
		attachmentRepo: attachmentRepo,
		incidentRepo:   incidentRepo,
		storage:        storage,
	}
}

// UploadAttachment uploads a file to MinIO and creates an attachment record
func (u *attachmentUsecase) UploadAttachment(ctx context.Context, incidentID, userID uint, fileName string, fileSize int64, mimeType string, reader io.Reader) (*domain.Attachment, error) {
	// Validate incident exists
	_, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, domain.ErrNotFound("Incident")
	}

	// Validate file size (max 50MB)
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if fileSize > maxFileSize {
		return nil, domain.ErrValidation("file size exceeds maximum allowed size of 50MB")
	}

	// Validate file extension
	if !isAllowedFileType(fileName) {
		return nil, domain.ErrValidation("file type not allowed")
	}

	// Generate unique storage key
	ext := filepath.Ext(fileName)
	storageKey := fmt.Sprintf("incidents/%d/%s%s", incidentID, uuid.New().String(), ext)

	// Upload to MinIO
	if err := u.storage.Upload(ctx, storageKey, reader, fileSize, mimeType); err != nil {
		return nil, domain.ErrInternal("failed to upload file", err)
	}

	// Create attachment record
	attachment := &domain.Attachment{
		IncidentID: incidentID,
		UserID:     userID,
		FileName:   fileName,
		FileSize:   fileSize,
		MimeType:   mimeType,
		StorageKey: storageKey,
		CreatedAt:  time.Now(),
	}

	if err := u.attachmentRepo.Create(attachment); err != nil {
		// Attempt to delete the uploaded file if database insert fails
		_ = u.storage.Delete(ctx, storageKey)
		return nil, domain.ErrInternal("failed to create attachment record", err)
	}

	// Reload to get user relation
	return u.attachmentRepo.FindByID(attachment.ID)
}

// GetAttachmentsByIncidentID retrieves all attachments for an incident
func (u *attachmentUsecase) GetAttachmentsByIncidentID(ctx context.Context, incidentID uint) ([]*domain.Attachment, error) {
	return u.attachmentRepo.FindByIncidentID(incidentID)
}

// GetAttachment retrieves an attachment by ID
func (u *attachmentUsecase) GetAttachment(ctx context.Context, id uint) (*domain.Attachment, error) {
	return u.attachmentRepo.FindByID(id)
}

// DownloadAttachment downloads a file from MinIO
func (u *attachmentUsecase) DownloadAttachment(ctx context.Context, id uint) (io.ReadCloser, error) {
	attachment, err := u.attachmentRepo.FindByID(id)
	if err != nil {
		return nil, domain.ErrNotFound("Attachment")
	}

	reader, err := u.storage.Download(ctx, attachment.StorageKey)
	if err != nil {
		return nil, domain.ErrInternal("failed to download file", err)
	}

	return reader, nil
}

// DeleteAttachment deletes an attachment (both from MinIO and database)
func (u *attachmentUsecase) DeleteAttachment(ctx context.Context, id, userID uint, userRole domain.Role) error {
	attachment, err := u.attachmentRepo.FindByID(id)
	if err != nil {
		return domain.ErrNotFound("Attachment")
	}

	// Check permissions: Only admin or the uploader can delete
	if userRole != domain.RoleAdmin && attachment.UserID != userID {
		return domain.ErrForbidden("you can only delete your own attachments")
	}

	// Delete from MinIO
	if err := u.storage.Delete(ctx, attachment.StorageKey); err != nil {
		// Log error but continue with database deletion
		fmt.Printf("Warning: failed to delete file from storage: %v\n", err)
	}

	// Delete from database
	return u.attachmentRepo.Delete(id)
}

// isAllowedFileType checks if the file extension is allowed
func isAllowedFileType(fileName string) bool {
	allowedExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", // Images
		".pdf", // PDF
		".txt", ".log", ".md", // Text files
		".json", ".xml", ".yaml", ".yml", // Config files
		".zip", ".tar", ".gz", // Archives
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

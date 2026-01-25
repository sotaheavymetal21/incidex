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

// UploadAttachment はファイルをMinIOにアップロードし、添付ファイルレコードを作成します
func (u *attachmentUsecase) UploadAttachment(ctx context.Context, incidentID, userID uint, fileName string, fileSize int64, mimeType string, reader io.Reader) (*domain.Attachment, error) {
	// インシデントが存在するかバリデーション
	_, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, domain.ErrNotFound("Incident")
	}

	// ファイルサイズをバリデーション（最大50MB）
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if fileSize > maxFileSize {
		return nil, domain.ErrValidation("file size exceeds maximum allowed size of 50MB")
	}

	// ファイル拡張子をバリデーション
	if !isAllowedFileType(fileName) {
		return nil, domain.ErrValidation("file type not allowed")
	}

	// 一意のストレージキーを生成
	ext := filepath.Ext(fileName)
	storageKey := fmt.Sprintf("incidents/%d/%s%s", incidentID, uuid.New().String(), ext)

	// MinIOにアップロード
	if err := u.storage.Upload(ctx, storageKey, reader, fileSize, mimeType); err != nil {
		return nil, domain.ErrInternal("failed to upload file", err)
	}

	// 添付ファイルレコードを作成
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
		// データベース挿入が失敗した場合、アップロードしたファイルの削除を試みる
		_ = u.storage.Delete(ctx, storageKey)
		return nil, domain.ErrInternal("failed to create attachment record", err)
	}

	// ユーザーリレーションを取得するためにリロード
	return u.attachmentRepo.FindByID(attachment.ID)
}

// GetAttachmentsByIncidentID はインシデントの全ての添付ファイルを取得します
func (u *attachmentUsecase) GetAttachmentsByIncidentID(ctx context.Context, incidentID uint) ([]*domain.Attachment, error) {
	return u.attachmentRepo.FindByIncidentID(incidentID)
}

// GetAttachment はIDで添付ファイルを取得します
func (u *attachmentUsecase) GetAttachment(ctx context.Context, id uint) (*domain.Attachment, error) {
	return u.attachmentRepo.FindByID(id)
}

// DownloadAttachment はMinIOからファイルをダウンロードします
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

// DeleteAttachment は添付ファイルを削除します（MinIOとデータベースの両方から）
func (u *attachmentUsecase) DeleteAttachment(ctx context.Context, id, userID uint, userRole domain.Role) error {
	attachment, err := u.attachmentRepo.FindByID(id)
	if err != nil {
		return domain.ErrNotFound("Attachment")
	}

	// 権限をチェック: 管理者またはアップロード者のみ削除可能
	if userRole != domain.RoleAdmin && attachment.UserID != userID {
		return domain.ErrForbidden("you can only delete your own attachments")
	}

	// MinIOから削除
	if err := u.storage.Delete(ctx, attachment.StorageKey); err != nil {
		// errorをログに記録するが、データベース削除は続行
		fmt.Printf("Warning: failed to delete file from storage: %v\n", err)
	}

	// データベースから削除
	return u.attachmentRepo.Delete(id)
}

// isAllowedFileType はファイル拡張子が許可されているかチェックします
func isAllowedFileType(fileName string) bool {
	allowedExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", // 画像
		".pdf", // PDF
		".txt", ".log", ".md", // テキストファイル
		".json", ".xml", ".yaml", ".yml", // 設定ファイル
		".zip", ".tar", ".gz", // アーカイブ
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

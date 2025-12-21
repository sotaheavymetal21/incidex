package persistence

import (
	"incidex/internal/domain"

	"gorm.io/gorm"
)

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) domain.AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) Create(attachment *domain.Attachment) error {
	return r.db.Create(attachment).Error
}

func (r *attachmentRepository) FindByID(id uint) (*domain.Attachment, error) {
	var attachment domain.Attachment
	err := r.db.Preload("User").First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *attachmentRepository) FindByIncidentID(incidentID uint) ([]*domain.Attachment, error) {
	var attachments []*domain.Attachment
	err := r.db.Where("incident_id = ?", incidentID).
		Preload("User").
		Order("created_at DESC").
		Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *attachmentRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Attachment{}, id).Error
}

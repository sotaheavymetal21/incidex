package domain

import "time"

type Attachment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	IncidentID uint      `gorm:"not null;index" json:"incident_id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	FileName   string    `gorm:"size:255;not null" json:"file_name"`
	FileSize   int64     `gorm:"not null" json:"file_size"`
	MimeType   string    `gorm:"size:100;not null" json:"mime_type"`
	StorageKey string    `gorm:"size:500;not null;unique" json:"storage_key"` // MinIO object key
	CreatedAt  time.Time `json:"created_at"`

	// Relations
	Incident *Incident `gorm:"foreignKey:IncidentID" json:"-"`
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type AttachmentRepository interface {
	Create(attachment *Attachment) error
	FindByID(id uint) (*Attachment, error)
	FindByIncidentID(incidentID uint) ([]*Attachment, error)
	Delete(id uint) error
}

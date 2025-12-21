package domain

import (
	"context"
	"time"
)

// IncidentTemplate represents a template for quickly creating incidents
type IncidentTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:200;not null;index" json:"name"`                  // テンプレート名
	Description string    `gorm:"type:text" json:"description"`                         // テンプレートの説明
	Title       string    `gorm:"size:500;not null" json:"title"`                       // インシデントタイトルのテンプレート
	Content     string    `gorm:"type:text;not null" json:"content"`                    // インシデント詳細のテンプレート
	Severity    Severity  `gorm:"size:20;not null" json:"severity"`                     // デフォルトの重要度
	ImpactScope string    `gorm:"size:500" json:"impact_scope"`                         // デフォルトの影響範囲
	CreatorID   uint      `gorm:"not null;index" json:"creator_id"`                     // テンプレート作成者
	IsPublic    bool      `gorm:"default:false;index" json:"is_public"`                 // 公開テンプレートか
	UsageCount  int       `gorm:"default:0" json:"usage_count"`                         // 使用回数
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Creator *User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Tags    []Tag `gorm:"many2many:template_tags" json:"tags,omitempty"` // デフォルトのタグ
}

// IncidentTemplateRepository defines the interface for incident template data access
type IncidentTemplateRepository interface {
	Create(ctx context.Context, template *IncidentTemplate) error
	FindAll(ctx context.Context, userID uint) ([]*IncidentTemplate, error) // ユーザーのテンプレート + 公開テンプレート
	FindByID(ctx context.Context, id uint) (*IncidentTemplate, error)
	Update(ctx context.Context, template *IncidentTemplate) error
	Delete(ctx context.Context, id uint) error
	IncrementUsageCount(ctx context.Context, id uint) error
}

package domain

import "time"

// NotificationSetting はユーザーごとの通知設定を表します
type NotificationSetting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// 通知チャネル
	EmailEnabled  bool   `gorm:"default:true" json:"email_enabled"`
	SlackEnabled  bool   `gorm:"default:false" json:"slack_enabled"`
	SlackWebhook  string `gorm:"type:varchar(512)" json:"slack_webhook,omitempty"`

	// 通知イベントの有効/無効
	NotifyOnIncidentCreated       bool `gorm:"default:true" json:"notify_on_incident_created"`
	NotifyOnAssigned              bool `gorm:"default:true" json:"notify_on_assigned"`
	NotifyOnComment               bool `gorm:"default:true" json:"notify_on_comment"`
	NotifyOnStatusChange          bool `gorm:"default:true" json:"notify_on_status_change"`
	NotifyOnSeverityChange        bool `gorm:"default:true" json:"notify_on_severity_change"`
	NotifyOnResolved              bool `gorm:"default:true" json:"notify_on_resolved"`
	NotifyOnEscalation            bool `gorm:"default:true" json:"notify_on_escalation"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationSettingRepository は通知設定のリポジトリインターフェース
type NotificationSettingRepository interface {
	Create(setting *NotificationSetting) error
	Update(setting *NotificationSetting) error
	GetByUserID(userID uint) (*NotificationSetting, error)
	Delete(userID uint) error
}

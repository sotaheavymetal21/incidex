package domain

import "time"

// ActivityType は発生したアクティビティのタイプを表します
type ActivityType string

const (
	ActivityTypeCreated         ActivityType = "created"
	ActivityTypeComment         ActivityType = "comment"
	ActivityTypeStatusChange    ActivityType = "status_change"
	ActivityTypeSeverityChange  ActivityType = "severity_change"
	ActivityTypeAssigneeChange  ActivityType = "assignee_change"
	ActivityTypeResolved        ActivityType = "resolved"
	ActivityTypeReopened        ActivityType = "reopened"
	// タイムラインイベントタイプ
	ActivityTypeDetected              ActivityType = "detected"
	ActivityTypeInvestigationStarted   ActivityType = "investigation_started"
	ActivityTypeRootCauseIdentified    ActivityType = "root_cause_identified"
	ActivityTypeMitigation             ActivityType = "mitigation"
	ActivityTypeTimelineResolved       ActivityType = "timeline_resolved"
	ActivityTypeOther                  ActivityType = "other"
)

// IncidentActivity はインシデントに関連するアクティビティまたはイベントを表します
type IncidentActivity struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	IncidentID  uint         `gorm:"not null;index" json:"incident_id"`
	UserID      uint         `gorm:"not null;index" json:"user_id"`
	ActivityType ActivityType `gorm:"size:50;not null;index" json:"activity_type"`
	Comment     string       `gorm:"type:text" json:"comment,omitempty"`
	OldValue    string       `gorm:"size:100" json:"old_value,omitempty"`
	NewValue    string       `gorm:"size:100" json:"new_value,omitempty"`
	CreatedAt   time.Time    `gorm:"index" json:"created_at"`

	// リレーション
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Incident *Incident `gorm:"foreignKey:IncidentID" json:"-"`
}

// IncidentActivityRepository はインシデントアクティビティデータアクセスのインターフェースを定義します
type IncidentActivityRepository interface {
	Create(activity *IncidentActivity) error
	FindByIncidentID(incidentID uint, limit int) ([]*IncidentActivity, error)
	FindRecent(limit int) ([]*IncidentActivity, error)
}

package domain

import "time"

// ActivityType represents the type of activity that occurred.
type ActivityType string

const (
	ActivityTypeCreated         ActivityType = "created"
	ActivityTypeComment         ActivityType = "comment"
	ActivityTypeStatusChange    ActivityType = "status_change"
	ActivityTypeSeverityChange  ActivityType = "severity_change"
	ActivityTypeAssigneeChange  ActivityType = "assignee_change"
	ActivityTypeResolved        ActivityType = "resolved"
	ActivityTypeReopened        ActivityType = "reopened"
)

// IncidentActivity represents an activity or event related to an incident.
type IncidentActivity struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	IncidentID  uint         `gorm:"not null;index" json:"incident_id"`
	UserID      uint         `gorm:"not null;index" json:"user_id"`
	ActivityType ActivityType `gorm:"size:50;not null;index" json:"activity_type"`
	Comment     string       `gorm:"type:text" json:"comment,omitempty"`
	OldValue    string       `gorm:"size:100" json:"old_value,omitempty"`
	NewValue    string       `gorm:"size:100" json:"new_value,omitempty"`
	CreatedAt   time.Time    `gorm:"index" json:"created_at"`

	// Relations
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Incident *Incident `gorm:"foreignKey:IncidentID" json:"-"`
}

// IncidentActivityRepository defines the interface for incident activity data access.
type IncidentActivityRepository interface {
	Create(activity *IncidentActivity) error
	FindByIncidentID(incidentID uint, limit int) ([]*IncidentActivity, error)
	FindRecent(limit int) ([]*IncidentActivity, error)
}

package persistence

import (
	"incidex/internal/domain"

	"gorm.io/gorm"
)

type incidentActivityRepository struct {
	db *gorm.DB
}

func NewIncidentActivityRepository(db *gorm.DB) domain.IncidentActivityRepository {
	return &incidentActivityRepository{db: db}
}

func (r *incidentActivityRepository) Create(activity *domain.IncidentActivity) error {
	return r.db.Create(activity).Error
}

func (r *incidentActivityRepository) FindByIncidentID(incidentID uint, limit int) ([]*domain.IncidentActivity, error) {
	var activities []*domain.IncidentActivity
	query := r.db.Where("incident_id = ?", incidentID).
		Preload("User").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *incidentActivityRepository) FindRecent(limit int) ([]*domain.IncidentActivity, error) {
	var activities []*domain.IncidentActivity
	query := r.db.Preload("User").
		Preload("Incident").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

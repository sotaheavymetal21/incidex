package persistence

import (
	"context"
	"incidex/internal/domain"

	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepository) FindAll(ctx context.Context, filters domain.AuditLogFilters) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Action != nil {
		query = query.Where("action = ?", *filters.Action)
	}
	if filters.ResourceType != nil {
		query = query.Where("resource_type = ?", *filters.ResourceType)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := filters.Page
	limit := filters.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Fetch logs
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *auditLogRepository) FindByID(ctx context.Context, id uint) (*domain.AuditLog, error) {
	var log domain.AuditLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

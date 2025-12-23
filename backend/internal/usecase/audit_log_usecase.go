package usecase

import (
	"context"
	"incidex/internal/domain"
)

type AuditLogUsecase interface {
	GetAll(ctx context.Context, filters domain.AuditLogFilters) ([]*domain.AuditLog, int64, error)
	GetByID(ctx context.Context, id uint) (*domain.AuditLog, error)
}

type auditLogUsecase struct {
	auditLogRepo domain.AuditLogRepository
}

func NewAuditLogUsecase(auditLogRepo domain.AuditLogRepository) AuditLogUsecase {
	return &auditLogUsecase{
		auditLogRepo: auditLogRepo,
	}
}

func (u *auditLogUsecase) GetAll(ctx context.Context, filters domain.AuditLogFilters) ([]*domain.AuditLog, int64, error) {
	return u.auditLogRepo.FindAll(ctx, filters)
}

func (u *auditLogUsecase) GetByID(ctx context.Context, id uint) (*domain.AuditLog, error) {
	return u.auditLogRepo.FindByID(ctx, id)
}

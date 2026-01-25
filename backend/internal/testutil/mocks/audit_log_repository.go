package mocks

import (
	"context"
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockAuditLogRepository は domain.AuditLogRepository のモック実装です
type MockAuditLogRepository struct {
	mock.Mock
}

func NewMockAuditLogRepository() *MockAuditLogRepository {
	return &MockAuditLogRepository{}
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditLogRepository) FindAll(ctx context.Context, filters domain.AuditLogFilters) ([]*domain.AuditLog, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.AuditLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditLogRepository) FindByID(ctx context.Context, id uint) (*domain.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

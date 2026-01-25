package mocks

import (
	"context"
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockIncidentRepository は domain.IncidentRepository のモック実装です
type MockIncidentRepository struct {
	mock.Mock
}

func NewMockIncidentRepository() *MockIncidentRepository {
	return &MockIncidentRepository{}
}

func (m *MockIncidentRepository) Create(ctx context.Context, incident *domain.Incident) error {
	args := m.Called(ctx, incident)
	return args.Error(0)
}

func (m *MockIncidentRepository) FindAll(ctx context.Context, filters domain.IncidentFilters, pagination domain.Pagination) ([]*domain.Incident, *domain.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var result *domain.PaginationResult
	if args.Get(1) != nil {
		result = args.Get(1).(*domain.PaginationResult)
	}
	return args.Get(0).([]*domain.Incident), result, args.Error(2)
}

func (m *MockIncidentRepository) FindByID(ctx context.Context, id uint) (*domain.Incident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Incident), args.Error(1)
}

func (m *MockIncidentRepository) Update(ctx context.Context, incident *domain.Incident) error {
	args := m.Called(ctx, incident)
	return args.Error(0)
}

func (m *MockIncidentRepository) UpdateAssignee(ctx context.Context, incidentID uint, assigneeID *uint) error {
	args := m.Called(ctx, incidentID, assigneeID)
	return args.Error(0)
}

func (m *MockIncidentRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIncidentRepository) Count(count *int64) error {
	args := m.Called(count)
	return args.Error(0)
}

func (m *MockIncidentRepository) CountBySeverity(severity domain.Severity, count *int64) error {
	args := m.Called(severity, count)
	return args.Error(0)
}

func (m *MockIncidentRepository) CountByStatus(status domain.Status, count *int64) error {
	args := m.Called(status, count)
	return args.Error(0)
}

func (m *MockIncidentRepository) FindRecent(limit int) ([]*domain.Incident, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Incident), args.Error(1)
}

func (m *MockIncidentRepository) GetAllIncidents() ([]*domain.Incident, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Incident), args.Error(1)
}

func (m *MockIncidentRepository) CountSLAViolated(count *int64) error {
	args := m.Called(count)
	return args.Error(0)
}

func (m *MockIncidentRepository) GetSLAMetrics() (*domain.SLAMetrics, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SLAMetrics), args.Error(1)
}

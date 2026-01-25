package mocks

import (
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockIncidentActivityRepository is a mock implementation of domain.IncidentActivityRepository
type MockIncidentActivityRepository struct {
	mock.Mock
}

func NewMockIncidentActivityRepository() *MockIncidentActivityRepository {
	return &MockIncidentActivityRepository{}
}

func (m *MockIncidentActivityRepository) Create(activity *domain.IncidentActivity) error {
	args := m.Called(activity)
	return args.Error(0)
}

func (m *MockIncidentActivityRepository) FindByIncidentID(incidentID uint, limit int) ([]*domain.IncidentActivity, error) {
	args := m.Called(incidentID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncidentActivity), args.Error(1)
}

func (m *MockIncidentActivityRepository) FindRecent(limit int) ([]*domain.IncidentActivity, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncidentActivity), args.Error(1)
}

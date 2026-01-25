package mocks

import (
	"context"
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockPostMortemRepository is a mock implementation of domain.PostMortemRepository
type MockPostMortemRepository struct {
	mock.Mock
}

func NewMockPostMortemRepository() *MockPostMortemRepository {
	return &MockPostMortemRepository{}
}

func (m *MockPostMortemRepository) Create(ctx context.Context, pm *domain.PostMortem) error {
	args := m.Called(ctx, pm)
	return args.Error(0)
}

func (m *MockPostMortemRepository) FindByID(ctx context.Context, id uint) (*domain.PostMortem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemRepository) FindByIncidentID(ctx context.Context, incidentID uint) (*domain.PostMortem, error) {
	args := m.Called(ctx, incidentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostMortem), args.Error(1)
}

func (m *MockPostMortemRepository) Update(ctx context.Context, pm *domain.PostMortem) error {
	args := m.Called(ctx, pm)
	return args.Error(0)
}

func (m *MockPostMortemRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostMortemRepository) FindAll(ctx context.Context, filters domain.PostMortemFilters, pagination domain.Pagination) ([]*domain.PostMortem, *domain.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var result *domain.PaginationResult
	if args.Get(1) != nil {
		result = args.Get(1).(*domain.PaginationResult)
	}
	return args.Get(0).([]*domain.PostMortem), result, args.Error(2)
}

package mocks

import (
	"context"
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockActionItemRepository は domain.ActionItemRepository のモック実装です
type MockActionItemRepository struct {
	mock.Mock
}

func NewMockActionItemRepository() *MockActionItemRepository {
	return &MockActionItemRepository{}
}

func (m *MockActionItemRepository) Create(ctx context.Context, item *domain.ActionItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockActionItemRepository) FindByID(ctx context.Context, id uint) (*domain.ActionItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ActionItem), args.Error(1)
}

func (m *MockActionItemRepository) FindByPostMortemID(ctx context.Context, postMortemID uint) ([]*domain.ActionItem, error) {
	args := m.Called(ctx, postMortemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ActionItem), args.Error(1)
}

func (m *MockActionItemRepository) Update(ctx context.Context, item *domain.ActionItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockActionItemRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockActionItemRepository) FindAll(ctx context.Context, filters domain.ActionItemFilters, pagination domain.Pagination) ([]*domain.ActionItem, *domain.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var result *domain.PaginationResult
	if args.Get(1) != nil {
		result = args.Get(1).(*domain.PaginationResult)
	}
	return args.Get(0).([]*domain.ActionItem), result, args.Error(2)
}

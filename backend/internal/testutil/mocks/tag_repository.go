package mocks

import (
	"context"
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockTagRepository は domain.TagRepository のモック実装です
type MockTagRepository struct {
	mock.Mock
}

func NewMockTagRepository() *MockTagRepository {
	return &MockTagRepository{}
}

func (m *MockTagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockTagRepository) FindAll(ctx context.Context) ([]*domain.Tag, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Tag), args.Error(1)
}

func (m *MockTagRepository) FindByID(ctx context.Context, id uint) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}

func (m *MockTagRepository) FindByIDs(ctx context.Context, ids []uint) ([]domain.Tag, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Tag), args.Error(1)
}

func (m *MockTagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockTagRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

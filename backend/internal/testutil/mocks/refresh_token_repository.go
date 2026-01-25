package mocks

import (
	"context"
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockRefreshTokenRepository is a mock implementation of domain.RefreshTokenRepository
type MockRefreshTokenRepository struct {
	mock.Mock
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{}
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

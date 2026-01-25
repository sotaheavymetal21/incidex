package mocks

import (
	"context"
	"incidex/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockPasswordResetTokenRepository は domain.PasswordResetTokenRepository のモック実装です
type MockPasswordResetTokenRepository struct {
	mock.Mock
}

func NewMockPasswordResetTokenRepository() *MockPasswordResetTokenRepository {
	return &MockPasswordResetTokenRepository{}
}

func (m *MockPasswordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockPasswordResetTokenRepository) FindByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetTokenRepository) MarkAsUsed(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPasswordResetTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPasswordResetTokenRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

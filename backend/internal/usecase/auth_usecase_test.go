package usecase

import (
	"context"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func createTestAuthUsecase(userRepo *mocks.MockUserRepository, refreshTokenRepo *mocks.MockRefreshTokenRepository) AuthUsecase {
	return NewAuthUsecase(userRepo, refreshTokenRepo, "test-secret-key-12345", 15*time.Minute)
}

func TestAuthUsecase_Register(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful registration with valid data", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		userRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := usecase.Register(ctx, "New User", "newuser@example.com", "StrongPass123!", "EMP-001", "Engineering")

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "New User", user.Name)
		assert.Equal(t, "newuser@example.com", user.Email)
		assert.Equal(t, domain.RoleViewer, user.Role)
		assert.True(t, user.IsActive)

		userRepo.AssertExpectations(t)
	})

	t.Run("registration fails when email already exists", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		existingUser := testutil.NewTestUser()
		userRepo.On("FindByEmail", ctx, "existing@example.com").Return(existingUser, nil)

		user, err := usecase.Register(ctx, "New User", "existing@example.com", "StrongPass123!", "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeConflict, domainErr.Code)

		userRepo.AssertExpectations(t)
	})

	t.Run("registration fails with weak password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		userRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, nil)

		user, err := usecase.Register(ctx, "New User", "newuser@example.com", "weak", "", "")

		require.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("registration fails with common password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		userRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, nil)

		user, err := usecase.Register(ctx, "New User", "newuser@example.com", "password123", "", "")

		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "common")
	})

	t.Run("registration fails with invalid name", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		user, err := usecase.Register(ctx, "", "newuser@example.com", "StrongPass123!", "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})

	t.Run("registration fails with invalid email", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		user, err := usecase.Register(ctx, "New User", "invalid-email", "StrongPass123!", "", "")

		require.Error(t, err)
		assert.Nil(t, user)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)
	})
}

func TestAuthUsecase_Login(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful login returns tokens", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
		user := testutil.NewTestUser(func(u *domain.User) {
			u.PasswordHash = string(hashedPassword)
			u.Email = "login@example.com"
		})

		userRepo.On("FindByEmail", ctx, "login@example.com").Return(user, nil)
		refreshTokenRepo.On("Create", ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

		response, err := usecase.Login(ctx, "login@example.com", "StrongPass123!")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, user, response.User)

		userRepo.AssertExpectations(t)
		refreshTokenRepo.AssertExpectations(t)
	})

	t.Run("login fails for non-existent user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		userRepo.On("FindByEmail", ctx, "nonexistent@example.com").Return(nil, nil)

		response, err := usecase.Login(ctx, "nonexistent@example.com", "AnyPassword123!")

		require.Error(t, err)
		assert.Nil(t, response)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeUnauthorized, domainErr.Code)
	})

	t.Run("login fails with incorrect password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass123!"), bcrypt.DefaultCost)
		user := testutil.NewTestUser(func(u *domain.User) {
			u.PasswordHash = string(hashedPassword)
		})

		userRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)

		response, err := usecase.Login(ctx, "test@example.com", "WrongPass123!")

		require.Error(t, err)
		assert.Nil(t, response)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeUnauthorized, domainErr.Code)
	})

	t.Run("login fails for disabled account", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
		user := testutil.NewTestUser(func(u *domain.User) {
			u.PasswordHash = string(hashedPassword)
			u.IsActive = false
		})

		userRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)

		response, err := usecase.Login(ctx, "test@example.com", "StrongPass123!")

		require.Error(t, err)
		assert.Nil(t, response)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})
}

func TestAuthUsecase_RefreshAccessToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful refresh with valid token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		user := testutil.NewTestUser()
		refreshToken := testutil.NewTestRefreshToken(user.ID)

		refreshTokenRepo.On("FindByToken", ctx, "test-refresh-token-12345").Return(refreshToken, nil)
		userRepo.On("FindByID", ctx, user.ID).Return(user, nil)
		refreshTokenRepo.On("Create", ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
		refreshTokenRepo.On("RevokeByToken", ctx, "test-refresh-token-12345").Return(nil)

		response, err := usecase.RefreshAccessToken(ctx, "test-refresh-token-12345")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotEqual(t, "test-refresh-token-12345", response.RefreshToken) // Should be rotated

		refreshTokenRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("refresh fails with invalid token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		refreshTokenRepo.On("FindByToken", ctx, "invalid-token").Return(nil, nil)

		response, err := usecase.RefreshAccessToken(ctx, "invalid-token")

		require.Error(t, err)
		assert.Nil(t, response)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeUnauthorized, domainErr.Code)
	})

	t.Run("refresh fails with expired token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		expiredToken := testutil.NewTestRefreshToken(1, func(rt *domain.RefreshToken) {
			rt.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired
		})

		refreshTokenRepo.On("FindByToken", ctx, "expired-token").Return(expiredToken, nil)

		response, err := usecase.RefreshAccessToken(ctx, "expired-token")

		require.Error(t, err)
		assert.Nil(t, response)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeUnauthorized, domainErr.Code)
	})

	t.Run("refresh fails with revoked token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		revokedAt := time.Now()
		revokedToken := testutil.NewTestRefreshToken(1, func(rt *domain.RefreshToken) {
			rt.RevokedAt = &revokedAt
		})

		refreshTokenRepo.On("FindByToken", ctx, "revoked-token").Return(revokedToken, nil)

		response, err := usecase.RefreshAccessToken(ctx, "revoked-token")

		require.Error(t, err)
		assert.Nil(t, response)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeUnauthorized, domainErr.Code)
	})

	t.Run("refresh fails for disabled user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.IsActive = false
		})
		refreshToken := testutil.NewTestRefreshToken(user.ID)

		refreshTokenRepo.On("FindByToken", ctx, "valid-token").Return(refreshToken, nil)
		userRepo.On("FindByID", ctx, user.ID).Return(user, nil)

		response, err := usecase.RefreshAccessToken(ctx, "valid-token")

		require.Error(t, err)
		assert.Nil(t, response)

		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeForbidden, domainErr.Code)
	})
}

func TestAuthUsecase_Logout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful logout revokes token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		refreshTokenRepo.On("RevokeByToken", ctx, "some-refresh-token").Return(nil)

		err := usecase.Logout(ctx, "some-refresh-token")

		require.NoError(t, err)
		refreshTokenRepo.AssertExpectations(t)
	})

	t.Run("logout with empty token does nothing", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		refreshTokenRepo := mocks.NewMockRefreshTokenRepository()
		usecase := createTestAuthUsecase(userRepo, refreshTokenRepo)

		err := usecase.Logout(ctx, "")

		require.NoError(t, err)
		// RevokeByToken should not be called
		refreshTokenRepo.AssertNotCalled(t, "RevokeByToken")
	})
}

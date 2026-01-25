package usecase

import (
	"context"
	"errors"
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

// EmailServiceInterface は EmailService のインターフェースを定義します
type EmailServiceInterface interface {
	SendPasswordResetEmail(to, userName, resetToken, frontendURL string) error
}

func createTestPasswordResetUsecase(
	userRepo *mocks.MockUserRepository,
	tokenRepo *mocks.MockPasswordResetTokenRepository,
	emailService EmailServiceInterface,
) *testPasswordResetUsecase {
	return &testPasswordResetUsecase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
		frontendURL:  "http://localhost:3000",
	}
}

// testPasswordResetUsecase は passwordResetUsecase のテスト用ラッパーです
type testPasswordResetUsecase struct {
	userRepo     domain.UserRepository
	tokenRepo    domain.PasswordResetTokenRepository
	emailService EmailServiceInterface
	frontendURL  string
}

func (u *testPasswordResetUsecase) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return domain.ErrDatabase("Failed to find user", err)
	}

	if user == nil {
		return nil
	}

	if !user.IsActive {
		return nil
	}

	if err := u.tokenRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return domain.ErrDatabase("Failed to delete existing tokens", err)
	}

	tokenString, err := domain.GenerateToken()
	if err != nil {
		return domain.ErrInternal("Failed to generate token", err)
	}

	resetToken := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(domain.PasswordResetTokenExpiration),
	}

	if err := u.tokenRepo.Create(ctx, resetToken); err != nil {
		return domain.ErrDatabase("Failed to create reset token", err)
	}

	if err := u.emailService.SendPasswordResetEmail(user.Email, user.Name, tokenString, u.frontendURL); err != nil {
		return domain.ErrInternal("Failed to send reset email", err)
	}

	return nil
}

func (u *testPasswordResetUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := u.tokenRepo.FindByToken(ctx, token)
	if err != nil {
		return domain.ErrDatabase("Failed to find token", err)
	}

	if resetToken == nil {
		return domain.ErrBadRequest("無効なトークンです")
	}

	if !resetToken.IsValid() {
		if resetToken.IsExpired() {
			return domain.ErrBadRequest("トークンの有効期限が切れています")
		}
		if resetToken.IsUsed() {
			return domain.ErrBadRequest("このトークンは既に使用されています")
		}
	}

	if err := domain.ValidatePasswordStrength(newPassword); err != nil {
		return domain.ErrValidation(err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternal("Failed to hash password", err)
	}

	if err := u.userRepo.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		return domain.ErrDatabase("Failed to update password", err)
	}

	if err := u.tokenRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return domain.ErrDatabase("Failed to mark token as used", err)
	}

	return nil
}

func (u *testPasswordResetUsecase) ValidateToken(ctx context.Context, token string) (bool, error) {
	resetToken, err := u.tokenRepo.FindByToken(ctx, token)
	if err != nil {
		return false, domain.ErrDatabase("Failed to find token", err)
	}

	if resetToken == nil {
		return false, nil
	}

	return resetToken.IsValid(), nil
}

func TestPasswordResetUsecase_RequestPasswordReset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful password reset request for active user", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.Email = "user@example.com"
			u.Name = "Test User"
			u.IsActive = true
		})

		userRepo.On("FindByEmail", ctx, "user@example.com").Return(user, nil)
		tokenRepo.On("DeleteByUserID", ctx, user.ID).Return(nil)
		tokenRepo.On("Create", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).Return(nil)
		emailService.On("SendPasswordResetEmail", "user@example.com", "Test User", mock.AnythingOfType("string"), "http://localhost:3000").Return(nil)

		err := usecase.RequestPasswordReset(ctx, "user@example.com")

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
		tokenRepo.AssertExpectations(t)
		emailService.AssertExpectations(t)
	})

	t.Run("returns success for non-existent user to prevent email enumeration", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		userRepo.On("FindByEmail", ctx, "nonexistent@example.com").Return(nil, nil)

		err := usecase.RequestPasswordReset(ctx, "nonexistent@example.com")

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
		// トークン作成とメール送信は呼ばれないはず
		tokenRepo.AssertNotCalled(t, "DeleteByUserID")
		tokenRepo.AssertNotCalled(t, "Create")
		emailService.AssertNotCalled(t, "SendPasswordResetEmail")
	})

	t.Run("returns success for inactive user without sending email", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.Email = "inactive@example.com"
			u.IsActive = false
		})

		userRepo.On("FindByEmail", ctx, "inactive@example.com").Return(user, nil)

		err := usecase.RequestPasswordReset(ctx, "inactive@example.com")

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
		// 非アクティブユーザーにはメール送信しない
		tokenRepo.AssertNotCalled(t, "DeleteByUserID")
		tokenRepo.AssertNotCalled(t, "Create")
		emailService.AssertNotCalled(t, "SendPasswordResetEmail")
	})

	t.Run("deletes existing tokens before creating new one", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		user := testutil.NewTestUser(func(u *domain.User) {
			u.ID = 123
			u.Email = "user@example.com"
		})

		userRepo.On("FindByEmail", ctx, "user@example.com").Return(user, nil)
		tokenRepo.On("DeleteByUserID", ctx, uint(123)).Return(nil)
		tokenRepo.On("Create", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).Return(nil)
		emailService.On("SendPasswordResetEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := usecase.RequestPasswordReset(ctx, "user@example.com")

		require.NoError(t, err)
		tokenRepo.AssertCalled(t, "DeleteByUserID", ctx, uint(123))
	})

	t.Run("fails when user lookup returns database error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		dbErr := errors.New("database connection lost")
		userRepo.On("FindByEmail", ctx, "user@example.com").Return(nil, dbErr)

		err := usecase.RequestPasswordReset(ctx, "user@example.com")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})

	t.Run("fails when token deletion returns error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		user := testutil.NewTestUser()
		dbErr := errors.New("delete failed")

		userRepo.On("FindByEmail", ctx, "user@example.com").Return(user, nil)
		tokenRepo.On("DeleteByUserID", ctx, user.ID).Return(dbErr)

		err := usecase.RequestPasswordReset(ctx, "user@example.com")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})

	t.Run("fails when token creation returns error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		user := testutil.NewTestUser()
		dbErr := errors.New("create failed")

		userRepo.On("FindByEmail", ctx, "user@example.com").Return(user, nil)
		tokenRepo.On("DeleteByUserID", ctx, user.ID).Return(nil)
		tokenRepo.On("Create", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).Return(dbErr)

		err := usecase.RequestPasswordReset(ctx, "user@example.com")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})

	t.Run("fails when email sending returns error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		user := testutil.NewTestUser()
		emailErr := errors.New("SMTP connection failed")

		userRepo.On("FindByEmail", ctx, "user@example.com").Return(user, nil)
		tokenRepo.On("DeleteByUserID", ctx, user.ID).Return(nil)
		tokenRepo.On("Create", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).Return(nil)
		emailService.On("SendPasswordResetEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(emailErr)

		err := usecase.RequestPasswordReset(ctx, "user@example.com")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeInternal, domainErr.Code)
	})
}

func TestPasswordResetUsecase_ResetPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successfully resets password with valid token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "valid-token-123"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = nil
		})

		tokenRepo.On("FindByToken", ctx, "valid-token-123").Return(resetToken, nil)
		userRepo.On("UpdatePassword", ctx, uint(1), mock.AnythingOfType("string")).Return(nil)
		tokenRepo.On("MarkAsUsed", ctx, resetToken.ID).Return(nil)

		err := usecase.ResetPassword(ctx, "valid-token-123", "NewStrongPass123!")

		require.NoError(t, err)
		tokenRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)

		// パスワードがハッシュ化されて保存されていることを確認
		userRepo.AssertCalled(t, "UpdatePassword", ctx, uint(1), mock.MatchedBy(func(hashedPassword string) bool {
			// bcryptハッシュであることを確認（$2a$で始まる）
			return len(hashedPassword) > 0 && hashedPassword[:4] == "$2a$"
		}))
	})

	t.Run("fails with invalid token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		tokenRepo.On("FindByToken", ctx, "invalid-token").Return(nil, nil)

		err := usecase.ResetPassword(ctx, "invalid-token", "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeBadRequest, domainErr.Code)
		assert.Contains(t, err.Error(), "無効なトークン")

		userRepo.AssertNotCalled(t, "UpdatePassword")
		tokenRepo.AssertNotCalled(t, "MarkAsUsed")
	})

	t.Run("fails with expired token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "expired-token"
			rt.ExpiresAt = time.Now().Add(-30 * time.Minute) // 30分前に期限切れ
			rt.UsedAt = nil
		})

		tokenRepo.On("FindByToken", ctx, "expired-token").Return(resetToken, nil)

		err := usecase.ResetPassword(ctx, "expired-token", "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeBadRequest, domainErr.Code)
		assert.Contains(t, err.Error(), "有効期限が切れています")

		userRepo.AssertNotCalled(t, "UpdatePassword")
		tokenRepo.AssertNotCalled(t, "MarkAsUsed")
	})

	t.Run("fails with already used token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		usedAt := time.Now().Add(-10 * time.Minute)
		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "used-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = &usedAt
		})

		tokenRepo.On("FindByToken", ctx, "used-token").Return(resetToken, nil)

		err := usecase.ResetPassword(ctx, "used-token", "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeBadRequest, domainErr.Code)
		assert.Contains(t, err.Error(), "既に使用されています")

		userRepo.AssertNotCalled(t, "UpdatePassword")
		tokenRepo.AssertNotCalled(t, "MarkAsUsed")
	})

	t.Run("fails with weak password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "valid-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = nil
		})

		tokenRepo.On("FindByToken", ctx, "valid-token").Return(resetToken, nil)

		err := usecase.ResetPassword(ctx, "valid-token", "weak")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)

		userRepo.AssertNotCalled(t, "UpdatePassword")
		tokenRepo.AssertNotCalled(t, "MarkAsUsed")
	})

	t.Run("fails with common password", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "valid-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = nil
		})

		tokenRepo.On("FindByToken", ctx, "valid-token").Return(resetToken, nil)

		err := usecase.ResetPassword(ctx, "valid-token", "password123")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeValidation, domainErr.Code)

		userRepo.AssertNotCalled(t, "UpdatePassword")
		tokenRepo.AssertNotCalled(t, "MarkAsUsed")
	})

	t.Run("fails when token lookup returns database error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		dbErr := errors.New("database error")
		tokenRepo.On("FindByToken", ctx, "some-token").Return(nil, dbErr)

		err := usecase.ResetPassword(ctx, "some-token", "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})

	t.Run("fails when password update returns error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "valid-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = nil
		})

		dbErr := errors.New("update failed")
		tokenRepo.On("FindByToken", ctx, "valid-token").Return(resetToken, nil)
		userRepo.On("UpdatePassword", ctx, uint(1), mock.AnythingOfType("string")).Return(dbErr)

		err := usecase.ResetPassword(ctx, "valid-token", "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)

		tokenRepo.AssertNotCalled(t, "MarkAsUsed")
	})

	t.Run("fails when marking token as used returns error", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "valid-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = nil
		})

		dbErr := errors.New("mark as used failed")
		tokenRepo.On("FindByToken", ctx, "valid-token").Return(resetToken, nil)
		userRepo.On("UpdatePassword", ctx, uint(1), mock.AnythingOfType("string")).Return(nil)
		tokenRepo.On("MarkAsUsed", ctx, resetToken.ID).Return(dbErr)

		err := usecase.ResetPassword(ctx, "valid-token", "NewStrongPass123!")

		require.Error(t, err)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})

	t.Run("verifies password is properly hashed with bcrypt", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "valid-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = nil
		})

		var capturedHash string
		tokenRepo.On("FindByToken", ctx, "valid-token").Return(resetToken, nil)
		userRepo.On("UpdatePassword", ctx, uint(1), mock.AnythingOfType("string")).
			Run(func(args mock.Arguments) {
				capturedHash = args.Get(2).(string)
			}).
			Return(nil)
		tokenRepo.On("MarkAsUsed", ctx, resetToken.ID).Return(nil)

		newPassword := "MyNewSecurePass456!"
		err := usecase.ResetPassword(ctx, "valid-token", newPassword)

		require.NoError(t, err)

		// bcryptハッシュが元のパスワードと一致することを確認
		err = bcrypt.CompareHashAndPassword([]byte(capturedHash), []byte(newPassword))
		assert.NoError(t, err, "Hashed password should match the original password")
	})
}

func TestPasswordResetUsecase_ValidateToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns true for valid token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "valid-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = nil
		})

		tokenRepo.On("FindByToken", ctx, "valid-token").Return(resetToken, nil)

		isValid, err := usecase.ValidateToken(ctx, "valid-token")

		require.NoError(t, err)
		assert.True(t, isValid)
		tokenRepo.AssertExpectations(t)
	})

	t.Run("returns false for non-existent token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		tokenRepo.On("FindByToken", ctx, "nonexistent-token").Return(nil, nil)

		isValid, err := usecase.ValidateToken(ctx, "nonexistent-token")

		require.NoError(t, err)
		assert.False(t, isValid)
		tokenRepo.AssertExpectations(t)
	})

	t.Run("returns false for expired token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "expired-token"
			rt.ExpiresAt = time.Now().Add(-30 * time.Minute)
			rt.UsedAt = nil
		})

		tokenRepo.On("FindByToken", ctx, "expired-token").Return(resetToken, nil)

		isValid, err := usecase.ValidateToken(ctx, "expired-token")

		require.NoError(t, err)
		assert.False(t, isValid)
		tokenRepo.AssertExpectations(t)
	})

	t.Run("returns false for used token", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		usedAt := time.Now().Add(-10 * time.Minute)
		resetToken := testutil.NewTestPasswordResetToken(1, func(rt *domain.PasswordResetToken) {
			rt.Token = "used-token"
			rt.ExpiresAt = time.Now().Add(30 * time.Minute)
			rt.UsedAt = &usedAt
		})

		tokenRepo.On("FindByToken", ctx, "used-token").Return(resetToken, nil)

		isValid, err := usecase.ValidateToken(ctx, "used-token")

		require.NoError(t, err)
		assert.False(t, isValid)
		tokenRepo.AssertExpectations(t)
	})

	t.Run("returns error when database lookup fails", func(t *testing.T) {
		t.Parallel()

		userRepo := mocks.NewMockUserRepository()
		tokenRepo := mocks.NewMockPasswordResetTokenRepository()
		emailService := mocks.NewMockEmailService()
		usecase := createTestPasswordResetUsecase(userRepo, tokenRepo, emailService)

		dbErr := errors.New("database connection lost")
		tokenRepo.On("FindByToken", ctx, "some-token").Return(nil, dbErr)

		isValid, err := usecase.ValidateToken(ctx, "some-token")

		require.Error(t, err)
		assert.False(t, isValid)
		domainErr, ok := domain.AsDomainError(err)
		require.True(t, ok)
		assert.Equal(t, domain.ErrCodeDatabaseError, domainErr.Code)
	})
}

package usecase

import (
	"context"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/notification"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type PasswordResetUsecase interface {
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ValidateToken(ctx context.Context, token string) (bool, error)
}

type passwordResetUsecase struct {
	userRepo       domain.UserRepository
	tokenRepo      domain.PasswordResetTokenRepository
	emailService   *notification.EmailService
	frontendURL    string
}

func NewPasswordResetUsecase(
	userRepo domain.UserRepository,
	tokenRepo domain.PasswordResetTokenRepository,
	emailService *notification.EmailService,
	frontendURL string,
) PasswordResetUsecase {
	return &passwordResetUsecase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
		frontendURL:  frontendURL,
	}
}

func (u *passwordResetUsecase) RequestPasswordReset(ctx context.Context, email string) error {
	// Find user by email
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return domain.ErrDatabase("Failed to find user", err)
	}

	// Always return success to prevent email enumeration attacks
	// but only send email if user exists
	if user == nil {
		return nil
	}

	// Check if user is active
	if !user.IsActive {
		return nil
	}

	// Delete any existing tokens for this user
	if err := u.tokenRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return domain.ErrDatabase("Failed to delete existing tokens", err)
	}

	// Generate new token
	tokenString, err := domain.GenerateToken()
	if err != nil {
		return domain.ErrInternal("Failed to generate token", err)
	}

	// Create token record
	resetToken := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(domain.PasswordResetTokenExpiration),
	}

	if err := u.tokenRepo.Create(ctx, resetToken); err != nil {
		return domain.ErrDatabase("Failed to create reset token", err)
	}

	// Send email
	if err := u.emailService.SendPasswordResetEmail(user.Email, user.Name, tokenString, u.frontendURL); err != nil {
		return domain.ErrInternal("Failed to send reset email", err)
	}

	return nil
}

func (u *passwordResetUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Find token
	resetToken, err := u.tokenRepo.FindByToken(ctx, token)
	if err != nil {
		return domain.ErrDatabase("Failed to find token", err)
	}

	if resetToken == nil {
		return domain.ErrBadRequest("無効なトークンです")
	}

	// Check if token is valid
	if !resetToken.IsValid() {
		if resetToken.IsExpired() {
			return domain.ErrBadRequest("トークンの有効期限が切れています")
		}
		if resetToken.IsUsed() {
			return domain.ErrBadRequest("このトークンは既に使用されています")
		}
	}

	// Validate password strength
	if err := domain.ValidatePasswordStrength(newPassword); err != nil {
		return domain.ErrValidation(err.Error())
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternal("Failed to hash password", err)
	}

	// Update password
	if err := u.userRepo.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		return domain.ErrDatabase("Failed to update password", err)
	}

	// Mark token as used
	if err := u.tokenRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return domain.ErrDatabase("Failed to mark token as used", err)
	}

	return nil
}

func (u *passwordResetUsecase) ValidateToken(ctx context.Context, token string) (bool, error) {
	resetToken, err := u.tokenRepo.FindByToken(ctx, token)
	if err != nil {
		return false, domain.ErrDatabase("Failed to find token", err)
	}

	if resetToken == nil {
		return false, nil
	}

	return resetToken.IsValid(), nil
}

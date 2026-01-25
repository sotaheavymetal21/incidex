package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"incidex/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *domain.User `json:"user"`
}

type AuthUsecase interface {
	Register(ctx context.Context, name, email, password, employeeNumber, department string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authUsecase struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	jwtSecret        []byte
	jwtExpiry        time.Duration
	refreshExpiry    time.Duration
}

func NewAuthUsecase(userRepo domain.UserRepository, refreshTokenRepo domain.RefreshTokenRepository, jwtSecret string, jwtExpiry time.Duration) AuthUsecase {
	return &authUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        []byte(jwtSecret),
		jwtExpiry:        jwtExpiry,
		refreshExpiry:    7 * 24 * time.Hour, // refresh tokenは7日間有効
	}
}

func (u *authUsecase) Register(ctx context.Context, name, email, password, employeeNumber, department string) (*domain.User, error) {
	// ユーザー入力をバリデーション
	if err := domain.ValidateUserInput(name, email, employeeNumber, department); err != nil {
		return nil, err
	}

	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrDatabase("Failed to check existing user", err)
	}
	if existingUser != nil {
		return nil, domain.ErrConflict("Email already exists")
	}

	// パスワード強度をバリデーション
	if err := domain.ValidatePasswordStrength(password); err != nil {
		return nil, domain.ErrValidation(err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal("Failed to hash password", err)
	}

	user := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleViewer, // デフォルトのロール
		IsActive:     true,
	}
	// 空でない場合のみオプションフィールドを設定
	if employeeNumber != "" {
		user.EmployeeNumber = &employeeNumber
	}
	if department != "" {
		user.Department = &department
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, domain.ErrDatabase("Failed to create user", err)
	}

	return user, nil
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrDatabase("Failed to find user", err)
	}
	if user == nil {
		return nil, domain.ErrUnauthorized("Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized("Invalid credentials")
	}

	// ユーザーがアクティブかチェック
	if !user.IsActive {
		return nil, domain.ErrForbidden("Account is disabled")
	}

	// access token（JWT）を生成
	accessToken, err := u.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// refresh tokenを生成
	refreshToken, err := u.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (u *authUsecase) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*AuthResponse, error) {
	// refresh tokenを検索
	refreshToken, err := u.refreshTokenRepo.FindByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, domain.ErrDatabase("Failed to find refresh token", err)
	}
	if refreshToken == nil {
		return nil, domain.ErrUnauthorized("Invalid refresh token")
	}

	// refresh tokenをバリデーション
	if !refreshToken.IsValid() {
		return nil, domain.ErrUnauthorized("Refresh token is expired or revoked")
	}

	// ユーザーを取得
	user, err := u.userRepo.FindByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, domain.ErrDatabase("Failed to find user", err)
	}
	if user == nil {
		return nil, domain.ErrUnauthorized("User not found")
	}

	// ユーザーがアクティブかチェック
	if !user.IsActive {
		return nil, domain.ErrForbidden("Account is disabled")
	}

	// 新しいaccess tokenを生成
	accessToken, err := u.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// オプション: 新しいrefresh tokenを生成し、古いものを無効化（ローテーション）
	newRefreshToken, err := u.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 古いrefresh tokenを無効化
	if err := u.refreshTokenRepo.RevokeByToken(ctx, refreshTokenStr); err != nil {
		return nil, domain.ErrDatabase("Failed to revoke old refresh token", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

func (u *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil // 無効化するものがない
	}

	// refresh tokenを無効化
	if err := u.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return domain.ErrDatabase("Failed to revoke refresh token", err)
	}

	return nil
}

// generateAccessToken は新しいJWT access tokenを作成します
func (u *authUsecase) generateAccessToken(user *domain.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(u.jwtExpiry).Unix(),
	})

	tokenString, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", domain.ErrInternal("Failed to generate access token", err)
	}

	return tokenString, nil
}

// generateRefreshToken は新しいrefresh tokenを作成し、データベースに保存します
func (u *authUsecase) generateRefreshToken(ctx context.Context, userID uint) (string, error) {
	// セキュアなランダムtokenを生成
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", domain.ErrInternal("Failed to generate refresh token", err)
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	// データベースにrefresh tokenを作成
	refreshToken := &domain.RefreshToken{
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: time.Now().Add(u.refreshExpiry),
	}

	if err := u.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return "", domain.ErrDatabase("Failed to save refresh token", err)
	}

	return tokenString, nil
}

package usecase

import (
	"context"
	"errors"
	"incidex/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
}

type authUsecase struct {
	userRepo  domain.UserRepository
	jwtSecret []byte
	jwtExpiry time.Duration
}

func NewAuthUsecase(userRepo domain.UserRepository, jwtSecret string, jwtExpiry time.Duration) AuthUsecase {
	return &authUsecase{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
	}
}

func (u *authUsecase) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleViewer, // Default role
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(u.jwtExpiry).Unix(),
	})

	tokenString, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

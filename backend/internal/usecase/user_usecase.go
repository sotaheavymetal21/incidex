package usecase

import (
	"context"
	"incidex/internal/domain"
)

type UserUsecase interface {
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
}

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return u.userRepo.FindAll(ctx)
}

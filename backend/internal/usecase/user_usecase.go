package usecase

import (
	"context"
	"errors"
	"incidex/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, id uint, name, email string, role domain.Role) (*domain.User, error)
	UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string) error
	Delete(ctx context.Context, id uint) error
}

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return nil, errors.New("user has been deleted")
	}
	return user, nil
}

func (u *userUsecase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := u.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	// Filter out deleted users
	activeUsers := make([]*domain.User, 0)
	for _, user := range users {
		if user.DeletedAt == nil {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers, nil
}

func (u *userUsecase) Update(ctx context.Context, id uint, name, email string, role domain.Role) (*domain.User, error) {
	// Find existing user
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return nil, errors.New("cannot update deleted user")
	}

	// Check if email is being changed to one that already exists
	if user.Email != email {
		existingUser, err := u.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already exists")
		}
	}

	// Update user fields
	user.Name = name
	user.Email = email
	user.Role = role

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return errors.New("cannot update password for deleted user")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)

	return u.userRepo.Update(ctx, user)
}

func (u *userUsecase) Delete(ctx context.Context, id uint) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.DeletedAt != nil {
		return errors.New("user already deleted")
	}

	return u.userRepo.Delete(ctx, id)
}

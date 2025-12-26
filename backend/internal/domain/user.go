package domain

import (
	"context"
	"time"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type User struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Email          string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string     `gorm:"not null" json:"-"`
	Name           string     `gorm:"not null" json:"name"`
	EmployeeNumber string     `gorm:"uniqueIndex" json:"employee_number,omitempty"`
	Department     string     `json:"department,omitempty"`
	Role           Role       `gorm:"not null;default:'viewer'" json:"role"`
	IsActive       bool       `gorm:"default:true;not null" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, id uint, passwordHash string) error
	Delete(ctx context.Context, id uint) error
	ToggleActive(ctx context.Context, id uint, isActive bool) error
}

package domain

import (
	"context"
	"time"
)

type User struct {
	ID            string    `json:"id"`
	FullName      string    `json:"fullName"`
	Email         string    `json:"email" validate:"required"`
	PhoneNumber   string    `json:"phoneNumber" validate:"required"`
	PasswordHash  string    `json:"password" validate:"required"`
	Role          string    `json:"role"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type GetData struct {
	Email string `json:"email"`
}

type UserRepository interface {
	GetUserByUsername(ctx context.Context, email string, phoneNumber string) (*User, error)
	CreateUser(ctx context.Context, user *User) error

	GetDataByToken(ctx context.Context, token string) (*GetData, error)
	MarkEmailVerified(ctx context.Context, userID string) error
}

type VerificateRepository interface {
	CreateToken(ctx context.Context, userID string, expiresAt time.Time) (string, error)
	GetUserByToken(ctx context.Context, token string) (string, error)
	DeleteToken(ctx context.Context, token string) error
}

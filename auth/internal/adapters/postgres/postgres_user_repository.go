package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "marketai/auth/internal/domain"
)

type AuthRepository struct {
	conn *pgxpool.Pool
}

func NewAuthRepository(conn *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{conn: conn}
}

func (r *AuthRepository) GetUserByUsername(ctx context.Context, email string, phoneNumber string) (*domain.User, error) {
	user := &domain.User{}
	err := r.conn.QueryRow(ctx, getByUserName, email, phoneNumber).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) error {
	err := r.conn.QueryRow(ctx, createUser,
		user.FullName,
		user.Email,
		user.PasswordHash,
		user.PhoneNumber,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *AuthRepository) GetDataByToken(ctx context.Context, token string) (*domain.GetData, error) {
	// TODO: Реализовать валидацию токена и получение данных пользователя
	// Пока возвращаем заглушку для тестирования
	return &domain.GetData{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "user",
	}, nil
}

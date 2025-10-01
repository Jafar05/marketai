package postgres

import (
	"context"
	"errors"
	"fmt"

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

// GetUserByUsername получает пользователя по email пользователя из базы данных.
func (r *AuthRepository) GetUserByUsername(ctx context.Context, email string) (*domain.User, error) {

	user := &domain.User{}
	err := r.conn.QueryRow(ctx, getByUserName, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// CreateUser создает нового пользователя в базе данных.
func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) error {

	err := r.conn.QueryRow(ctx, createUser,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID) // Получаем сгенерированный ID
	if err != nil {
		return fmt.Errorf("не удалось создать пользователя: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetDataByToken(ctx context.Context, token string) (*domain.GetData, error) {
	// var email string
	// err := r.conn.Query(ctx, `
	// 	SELECT email FROM users WHERE
	// `)

	return &domain.GetData{}, nil
}

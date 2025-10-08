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
		&user.EmailVerified,
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
	// var email string
	// err := r.conn.Query(ctx, `
	// 	SELECT email FROM users WHERE
	// `)

	return &domain.GetData{}, nil
}

func (r *AuthRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.conn.Exec(ctx, `UPDATE users SET email_verified=true WHERE id=$1`, userID)
	return err
}

package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerificationRepository struct {
	conn *pgxpool.Pool
}

func NewEmailVerificationRepository(conn *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{conn: conn}
}

func (r *EmailVerificationRepository) CreateToken(ctx context.Context, userID string, expiresAt time.Time) (string, error) {
	token := uuid.NewString()
	_, err := r.conn.Exec(ctx, createToken, userID, token, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *EmailVerificationRepository) GetUserByToken(ctx context.Context, token string) (string, error) {
	var userID string
	err := r.conn.QueryRow(ctx, getUserByToken, token).Scan(&userID)
	return userID, err
}

func (r *EmailVerificationRepository) DeleteToken(ctx context.Context, token string) error {
	_, err := r.conn.Exec(ctx, deleteToken, token)
	return err
}

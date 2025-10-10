package postgres

import (
	"context"
	"fmt"
	"marketai/cards/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardRepository struct {
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) CreateCard(ctx context.Context, card *domain.Card) error {
	query := `
		INSERT INTO cards (id, user_id, photo_url, short_description, title, description, tags, image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if card.ID == "" {
		card.ID = uuid.New().String()
	}

	_, err := r.db.Exec(ctx, query,
		card.ID,
		card.UserID,
		card.PhotoURL,
		card.ShortDescription,
		card.Title,
		card.Description,
		card.Tags,
		card.Image,
		card.CreatedAt,
		card.UpdatedAt,
	)

	return err
}

func (r *CardRepository) GetCardsByUserID(ctx context.Context, userID string) ([]*domain.Card, error) {
	query := `
		SELECT id, user_id, photo_url, short_description, title, description, tags, image, created_at, updated_at
		FROM cards
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		card := &domain.Card{}
		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.PhotoURL,
			&card.ShortDescription,
			&card.Title,
			&card.Description,
			&card.Tags,
			&card.Image,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *CardRepository) GetCardByID(ctx context.Context, id string) (*domain.Card, error) {
	query := `
		SELECT id, user_id, photo_url, short_description, title, description, tags, image, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	card := &domain.Card{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&card.ID,
		&card.UserID,
		&card.PhotoURL,
		&card.ShortDescription,
		&card.Title,
		&card.Description,
		&card.Tags,
		&card.Image,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("card not found")
		}
		return nil, err
	}

	return card, nil
}

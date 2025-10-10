package domain

import (
	"context"
	"time"
)

type Card struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	PhotoURL         string    `json:"photo_url"`
	ShortDescription string    `json:"short_description"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Tags             []string  `json:"tags"`
	Image            string    `json:"image"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CardRepository interface {
	CreateCard(ctx context.Context, card *Card) error
	GetCardsByUserID(ctx context.Context, userID string) ([]*Card, error)
	GetCardByID(ctx context.Context, id string) (*Card, error)
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)
}

type UserInfo struct {
	UserID string
	Role   string
}

type AIService interface {
	GenerateCardContent(ctx context.Context, photoURL, description string) (*GeneratedCard, error)
}

type GeneratedCard struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Image       string   `json:"image"`
}

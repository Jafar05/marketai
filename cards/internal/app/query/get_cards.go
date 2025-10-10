package query

import (
	"context"
	"marketai/cards/internal/domain"
)

type GetCardsByUserQuery struct {
	UserID string
}

type GetCardsByUserResult struct {
	Cards []*domain.Card
}

type GetCardsByUserHandler interface {
	Handle(ctx context.Context, query GetCardsByUserQuery) (*GetCardsByUserResult, error)
}

type getCardsByUserHandler struct {
	cardRepo domain.CardRepository
}

func NewGetCardsByUserHandler(cardRepo domain.CardRepository) *getCardsByUserHandler {
	return &getCardsByUserHandler{
		cardRepo: cardRepo,
	}
}

func (h *getCardsByUserHandler) Handle(ctx context.Context, query GetCardsByUserQuery) (*GetCardsByUserResult, error) {
	cards, err := h.cardRepo.GetCardsByUserID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	return &GetCardsByUserResult{Cards: cards}, nil
}

type GetCardByIDQuery struct {
	CardID string
}

type GetCardByIDResult struct {
	Card *domain.Card
}

type GetCardByIDHandler interface {
	Handle(ctx context.Context, query GetCardByIDQuery) (*GetCardByIDResult, error)
}

type getCardByIDHandler struct {
	cardRepo domain.CardRepository
}

func NewGetCardByIDHandler(cardRepo domain.CardRepository) *getCardByIDHandler {
	return &getCardByIDHandler{
		cardRepo: cardRepo,
	}
}

func (h *getCardByIDHandler) Handle(ctx context.Context, query GetCardByIDQuery) (*GetCardByIDResult, error) {
	card, err := h.cardRepo.GetCardByID(ctx, query.CardID)
	if err != nil {
		return nil, err
	}

	return &GetCardByIDResult{Card: card}, nil
}

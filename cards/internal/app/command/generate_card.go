package command

import (
	"context"
	"fmt"
	"marketai/cards/internal/domain"
	"time"

	"github.com/google/uuid"
)

type GenerateCardCommand struct {
	UserID           string
	PhotoURL         string
	ShortDescription string
}

type GenerateCardResult struct {
	Card *domain.Card
}

type GenerateCardHandler interface {
	Handle(ctx context.Context, cmd GenerateCardCommand) (*GenerateCardResult, error)
}

type generateCardHandler struct {
	cardRepo  domain.CardRepository
	aiService domain.AIService
}

func NewGenerateCardHandler(cardRepo domain.CardRepository, aiService domain.AIService) *generateCardHandler {
	return &generateCardHandler{
		cardRepo:  cardRepo,
		aiService: aiService,
	}
}

func (h *generateCardHandler) Handle(ctx context.Context, cmd GenerateCardCommand) (*GenerateCardResult, error) {
	// Генерируем контент карточки через AI
	generatedContent, err := h.aiService.GenerateCardContent(ctx, cmd.PhotoURL, cmd.ShortDescription)
	if err != nil {
		return nil, fmt.Errorf("failed to generate card content: %w", err)
	}

	// Создаем карточку
	card := &domain.Card{
		ID:               uuid.New().String(),
		UserID:           cmd.UserID,
		PhotoURL:         cmd.PhotoURL,
		ShortDescription: cmd.ShortDescription,
		Title:            generatedContent.Title,
		Description:      generatedContent.Description,
		Tags:             generatedContent.Tags,
		Image:            generatedContent.Image,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Сохраняем в базу данных
	if err := h.cardRepo.CreateCard(ctx, card); err != nil {
		return nil, fmt.Errorf("failed to save card: %w", err)
	}

	return &GenerateCardResult{Card: card}, nil
}

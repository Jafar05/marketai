package app

import (
	"marketai/cards/internal/adapters"
	"marketai/cards/internal/adapters/postgres"
	"marketai/cards/internal/app/command"
	"marketai/cards/internal/app/query"
)

type Commands struct {
	GenerateCard command.GenerateCardHandler
}

type Queries struct {
	GetCardsByUser query.GetCardsByUserHandler
	GetCardByID    query.GetCardByIDHandler
}

type AppCQRS struct {
	Commands Commands
	Queries  Queries
}

func NewAppCQRS(
	cardRepo *postgres.CardRepository,
	authService *adapters.AuthGRPCService,
	aiService *adapters.OpenAIService,
) *AppCQRS {
	return &AppCQRS{
		Commands: Commands{
			GenerateCard: command.NewGenerateCardHandler(cardRepo, aiService),
		},
		Queries: Queries{
			GetCardsByUser: query.NewGetCardsByUserHandler(cardRepo),
			GetCardByID:    query.NewGetCardByIDHandler(cardRepo),
		},
	}
}

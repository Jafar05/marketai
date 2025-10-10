package app

import (
	"marketai/auth/internal/adapters/postgres"
	"marketai/auth/internal/app/command"
	"marketai/auth/internal/app/query"
	"marketai/auth/internal/config"
)

type Commands struct {
	Register command.RegisterCommandHandler
}

type Queries struct {
	Login          query.LoginCommandHandler
	GetUserByToken query.GetDataByTokenHandler
}

type AppCQRS struct {
	Commands Commands
	Queries  Queries
}

func NewAppCQRS(
	userRepo *postgres.AuthRepository,
	cfg *config.Config,
) *AppCQRS {
	return &AppCQRS{
		Commands: Commands{
			Register: command.NewRegisterUserCommandHandler(userRepo, cfg),
		},
		Queries: Queries{
			Login:          query.NewLoginCommandHandler(userRepo, cfg),
			GetUserByToken: query.NewGetDataByTokenHandler(userRepo),
		},
	}
}

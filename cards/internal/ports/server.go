package ports

import (
	"marketai/cards/internal/adapters"
	"marketai/cards/internal/adapters/postgres"
	"marketai/cards/internal/app"
	"marketai/cards/internal/config"
	"marketai/cards/internal/domain"
	"marketai/pkg/bootstrap"
	"marketai/pkg/postgresql"

	"go.uber.org/fx"
)

func App() fx.Option {
	return fx.Options(
		bootstrap.AppOptions[*config.Config](
			bootstrap.WithSecrets[*config.Secrets](config.MapConfig),
			bootstrap.WithEcho[*config.Config](registerRoutes),
			postgresql.Connection[*config.Config](),
			fx.Provide(
				app.NewAppCQRS,
				postgres.NewCardRepository,
				adapters.NewAuthGRPCService,
				adapters.NewOpenAIService,
				fx.Annotate(
					func(authService *adapters.AuthGRPCService) domain.AuthService {
						return authService
					},
					fx.As(new(domain.AuthService)),
				),
			),
		),
	)
}

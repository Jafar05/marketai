package ports

import (
	"marketai/auth/internal/adapters/postgres"
	"marketai/auth/internal/adapters/postgres/migrations"
	"marketai/auth/internal/app"
	"marketai/auth/internal/config"
	auth_grpc_api "marketai/auth/proto/generated-source"

	"github.com/Jafar05/pkg/bootstrap"
	"github.com/Jafar05/pkg/grpc"
	"github.com/Jafar05/pkg/postgresql"
	"go.uber.org/fx"
)

func App() fx.Option {
	return bootstrap.AppOptions[*config.Config](
		bootstrap.WithSecrets[*config.Secrets](config.MapSecrets),

		bootstrap.WithEcho[*config.Config](registerRoutes),

		grpc.Server[*config.Config](
			grpc.WithListener(),
			grpc.WithRegisterFunc(auth_grpc_api.RegisterAuthServiceServer),
		),

		postgresql.Connection[*config.Config](
			postgresql.WithBindataMigrate(
				migrations.AssetNames(),
				migrations.Asset,
			),
		),

		fx.Provide(
			app.NewAppCQRS,
			postgres.NewAuthRepository,
			newGrpcServer,
		),
	)
}

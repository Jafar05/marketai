package ports

import (
	"marketai/auth/internal/adapters/postgres"
	"marketai/auth/internal/adapters/postgres/migrations"
	"marketai/auth/internal/app"
	"marketai/auth/internal/config"
	auth_grpc_api "marketai/auth/proto/generated-source"
	"marketai/pkg/bootstrap"
	"marketai/pkg/grpc"
	"marketai/pkg/logger"
	"marketai/pkg/postgresql"

	"github.com/go-playground/validator"
	"go.uber.org/fx"
)

func AppOptionsCustom(cfg *config.Config, opts ...fx.Option) fx.Option {
	return fx.Options(
		fx.Provide(func() *config.Config {
			return cfg
		}),
		fx.Options(opts...),
	)
}

func App(cfg *config.Config) fx.Option {
	return AppOptionsCustom(cfg,
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
			func() logger.AppLog {
				return logger.NewNop()
			},
			func() *validator.Validate {
				return validator.New()
			},
		),
	)
}

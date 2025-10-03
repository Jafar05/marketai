package postgresql

import (
	"context"
	"errors"
	"marketai/pkg/probes"

	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	GetConfig interface {
		PostgresConfig() *PostgresCfg
	}
)

func MapSecrets(
	c *PostgresCfg,
	s *Secrets,
) *PostgresCfg {

	if s.Password != "" {
		c.Password = s.Password
	}

	if s.User != "" {
		c.User = s.User
	}

	if s.MigratePassword != "" {
		c.MigratePassword = s.MigratePassword
	}

	if s.MigrateUser != "" {
		c.MigrateUser = s.MigrateUser
	}

	return c
}

func config[Config GetConfig](
	c Config,
) *PostgresCfg {
	return c.PostgresConfig()
}

func Connection[Config GetConfig](
	opts ...fx.Option,
) fx.Option {
	return fx.Options(

		fx.Provide(
			config[Config],
			newPgxPool,
		),
		fx.Options(opts...),
		probes.WithLiveCheck(
			postgresCheck,
			fx.ParamTags(probes.ContextAnnotation, "", ""),
		),
		probes.WithReadyCheck(
			postgresCheck,
			fx.ParamTags(probes.ContextAnnotation, "", ""),
		),
	)
}

func postgresCheck(
	ctx context.Context,
	conn *pgxpool.Pool,
	logger *zap.Logger,
) *probes.Check {
	return &probes.Check{
		Name: conn.Config().ConnConfig.Database,
		Func: func() error {
			l := logger.With(
				zap.String("conn", conn.Config().ConnConfig.Database),
			)

			err := conn.Ping(ctx)

			if err != nil && errors.Is(err, context.Canceled) {
				l.Debug("context.Canceled")
				return nil
			}

			if err != nil {
				l.Info(
					"check failed",
					zap.Error(err),
				)
				return err
			}

			return nil
		},
	}
}

func WithBindataMigrate(
	names []string,
	assetFunc bindata.AssetFunc,
) fx.Option {
	return fx.Options(
		fx.Provide(
			func() *bindata.AssetSource {
				return bindata.Resource(names, assetFunc)
			},
			newMigrate,
		),
		fx.Invoke(
			func(m *Migrate, logger *zap.Logger) (err error) {
				logger.Info("migrate start")
				defer func() {
					if err == nil {
						logger.Info("migrate done")
					}
				}()

				err = m.Run()
				return
			},
		),
	)
}

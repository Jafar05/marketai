package config

import (
	"marketai/pkg/bootstrap"
	"marketai/pkg/grpc"
	"marketai/pkg/logger"
	"marketai/pkg/postgresql"
	"marketai/pkg/probes"
)

type (
	Config struct {
		Http       *bootstrap.HttpConfig    `mapstructure:"http" validate:"required"`
		Logger     *logger.Config           `mapstructure:"logger" validate:"required"`
		Postgres   *postgresql.PostgresCfg  `mapstructure:"postgres"`
		Probes     *probes.ProbeCfg         `mapstructure:"probes" validate:"required"`
		Tracing    *bootstrap.TracingConfig `mapstructure:"tracing" validate:"required"`
		GrpcServer *grpc.ServerConfig       `mapstructure:"grpc" validate:"required"`

		JWTSecret string `mapstructure:"jwt_secret" env:"JWT_SECRET"`

		SMTP struct {
			MailerSendAPIKey string `mapstructure:"smtp_mailer_send_api_key"`
			From             string `mapstructure:"smtp_from"`
		} `mapstructure:"smtp"`
	}

	ServerConfig struct {
		Logger   *logger.Config
		Probes   *probes.ProbeCfg
		Postgres *postgresql.PostgresCfg
	}
)

func (c *Config) PostgresConfig() *postgresql.PostgresCfg {
	return c.Postgres
}

func (c *Config) LoggerConfig() *logger.Config {
	return c.Logger
}

func (c *Config) HttpConfig() *bootstrap.HttpConfig {
	return c.Http
}

func (c *Config) ProbesConfig() *probes.ProbeCfg {
	return c.Probes
}

func (c *Config) TracingConfig() *bootstrap.TracingConfig {
	return c.Tracing
}

func (c *Config) GrpcConfig() *grpc.ServerConfig {
	return c.GrpcServer
}

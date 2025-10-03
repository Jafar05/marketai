package config

import "marketai/pkg/postgresql"

type Secrets struct {
	Postgres *postgresql.Secrets `mapstructure:"pg" validate:"required"`
}

func MapSecrets(config *Config, secrets *Secrets) *Config {
	config.Postgres = postgresql.MapSecrets(config.Postgres, secrets.Postgres)
	return config
}

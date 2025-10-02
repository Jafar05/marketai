package config

import "github.com/Jafar05/pkg/postgresql"

type Secrets struct {
	Postgres *postgresql.Secrets `mapstructure:"pg" validate:"required"`
}

func MapSecrets(config *Config, secrets *Secrets) *Config {
	config.Postgres = postgresql.MapSecrets(config.Postgres, secrets.Postgres)
	return config
}

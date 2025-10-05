package config

import (
	"marketai/pkg/postgresql"
	"os"
)

type Secrets struct {
	Postgres *postgresql.Secrets `mapstructure:"pg" validate:"required"`
}

func MapSecrets(config *Config, secrets *Secrets) *Config {
	var dbPort string
	// if os.Getenv("HTTP_PORT") != "production" {
	// 	err := godotenv.Load("/Users/jafar/GolandProjects/MarketAI/auth/.env")

	// }

	dbPort = os.Getenv("HTTP_PORT")
	config.Http.Port = dbPort
	config.Postgres = postgresql.MapSecrets(config.Postgres, secrets.Postgres)
	return config
}

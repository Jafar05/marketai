package config

import (
	"log"
	"marketai/pkg/postgresql"
	"os"

	"github.com/joho/godotenv"
)

type Secrets struct {
	Postgres *postgresql.Secrets `mapstructure:"pg" validate:"required"`
}

func MapSecrets(config *Config, secrets *Secrets) *Config {
	var dbPort string
	if os.Getenv("HTTP_PORT") != "production" {
		err := godotenv.Load("/Users/jafar/GolandProjects/MarketAI/auth/.env")
		if err != nil {
			log.Fatalf("Ошибка загрузки .env: %v", err)
		}

	}

	dbPort = os.Getenv("HTTP_PORT")
	config.Http.Port = dbPort
	config.Postgres = postgresql.MapSecrets(config.Postgres, secrets.Postgres)
	return config
}

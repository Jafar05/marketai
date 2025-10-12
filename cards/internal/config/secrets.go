package config

import (
	"marketai/pkg/postgresql"
	"os"

	"github.com/joho/godotenv"
)

type Secrets struct {
	Postgres *postgresql.Secrets `mapstructure:"pg" validate:"required"`
}

func MapConfig(config *Config, secrets *Secrets) *Config {
	mapEnv(config, secrets)
	config.Postgres = postgresql.MapSecrets(config.Postgres, secrets.Postgres)
	return config
}

func mapEnv(config *Config, secrets *Secrets) {
	if os.Getenv("ENV") != "production" {
		godotenv.Load("/Users/jafar/GolandProjects/MarketAI/cards/.env")
	}

	serverPort := os.Getenv("CARDS_HTTP_PORT")
	postgresHost := os.Getenv("CARDS_POSTGRES_HOST")
	postgresPort := os.Getenv("CARDS_POSTGRES_PORT")
	postgresDbName := os.Getenv("CARDS_POSTGRES_DB_NAME")
	postgresDbUser := os.Getenv("CARDS_POSTGRES_DB_USER")
	postgresDbPassword := os.Getenv("CARDS_POSTGRES_DB_PASSWORD")

	deepseekApiKey := os.Getenv("DEEPSEEK_API_KEY")

	config.Http.Port = serverPort

	config.Postgres.Host = postgresHost
	config.Postgres.Port = postgresPort
	config.Postgres.DBName = postgresDbName
	secrets.Postgres.User = postgresDbUser
	secrets.Postgres.Password = postgresDbPassword

	config.AI.DeepseekAPIKey = deepseekApiKey
}

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

	serviceName := os.Getenv("SERVICE_NAME")

	grpcPort := os.Getenv("GRPC_PORT")

	authGrpcEndpoint := os.Getenv("AUTH_GRPC_ENDPOINT")

	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	aiModel := os.Getenv("AI_MODEL")

	if serverPort != "" {
		config.Http.Port = serverPort
	}

	if postgresHost != "" {
		config.Postgres.Host = postgresHost
	}
	if postgresPort != "" {
		config.Postgres.Port = postgresPort
	}
	if postgresDbName != "" {
		config.Postgres.DBName = postgresDbName
	}
	if postgresDbUser != "" {
		secrets.Postgres.User = postgresDbUser
	}
	if postgresDbPassword != "" {
		secrets.Postgres.Password = postgresDbPassword
	}

	if serviceName != "" {
		config.Tracing.ServiceName = serviceName
	}

	if grpcPort != "" {
		config.GrpcServer.Port = grpcPort
	}

	if authGrpcEndpoint != "" {
		config.Auth.GRPCEndpoint = authGrpcEndpoint
	}

	if openaiApiKey != "" {
		config.AI.OpenAIAPIKey = openaiApiKey
	}
	if aiModel != "" {
		config.AI.Model = aiModel
	}
}

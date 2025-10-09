package config

import (
	"github.com/joho/godotenv"
	"marketai/pkg/postgresql"
	"os"
	"strconv"
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
		godotenv.Load("/Users/jafar/GolandProjects/MarketAI/auth/.env")

	}

	httpPort := os.Getenv("HTTP_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	grpcPort := os.Getenv("GRPC_SERVER_PORT")
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresDbName := os.Getenv("POSTGRES_DB_NAME")
	postgresDbUser := os.Getenv("POSTGRES_DB_USER")
	postgresDbPassword := os.Getenv("POSTGRES_DB_PASSWORD")

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")
	smtPortValue, _ := strconv.Atoi(smtpPort)

	config.Http.Port = httpPort
	config.GrpcServer.Port = grpcPort
	config.JWTSecret = jwtSecret
	config.Postgres.Host = postgresHost
	config.Postgres.Port = postgresPort
	config.Postgres.DBName = postgresDbName
	secrets.Postgres.User = postgresDbUser
	secrets.Postgres.Password = postgresDbPassword

	config.SMTP.Host = smtpHost
	config.SMTP.Port = smtPortValue
	config.SMTP.Username = smtpUser
	config.SMTP.Password = smtpPassword
	config.SMTP.From = smtpFrom
}

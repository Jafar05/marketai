package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
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

func WithServerConfig(c *Config) *ServerConfig {
	return &ServerConfig{
		Logger:   c.Logger,
		Postgres: c.Postgres,
		Probes:   c.Probes,
	}
}

func LoadConfig() *Config {
	viper.AutomaticEnv()

	viper.BindEnv("http.port", "HTTP_PORT")
	// Добавьте дополнительные BindEnv для других полей, если нужно, например:
	// viper.BindEnv("jwt_secret", "JWT_SECRET")

	viper.SetConfigFile("configs/auth/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Ошибка разбора конфига: %v", err)
	}

	// Пример обработки переопределения для http.port (если нужно)
	port := viper.GetString("http.port")
	fmt.Println("port==", port)
	if port != "" {
		// Если нужно перезаписать в cfg, но поскольку Unmarshal уже сделал это, это может быть избыточно
		cfg.Http.Port = port // Предполагая, что HttpConfig имеет поле Port типа string
	}

	// Здесь можно добавить валидацию, если есть пакет validator
	// validate := validator.New()
	// if err := validate.Struct(&cfg); err != nil {
	// 	log.Fatalf("Ошибка валидации конфига: %v", err)
	// }

	return &cfg
}

package bootstrap

import (
	"errors"
	"flag"
	"fmt"
	"marketai/pkg/logger"
	"marketai/pkg/probes"
	"os"
	"runtime"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrEmptyConfigPath  = errors.New("config path flag value is empty")
	ErrEmptySecretsPath = errors.New("secrets path flag value is empty")
	ErrNoEchoRouting    = errors.New("no echo routing function defined")
)

func adjustMaxprocs(logger *zap.Logger) {
	opt := maxprocs.Logger(logger.Sugar().Infof)

	if _, err := maxprocs.Set(opt); err != nil {
		logger.Info("cannot adjust maxprocs", zap.Error(err))
	}

	logger.Info("GOMAXPROCS", zap.Int("val", runtime.GOMAXPROCS(0)))

}

type (
	appFlags struct {
		secretsPath string
		configPath  string
	}
)

const (
	flagConfig  = "config"
	flagSecrets = "secrets"
	envConfig   = "SBERCODE_CONFIG"
	envSecrets  = "SBERCODE_SECRETS"
)

func initFlags() *appFlags {
	configPath := os.Getenv(envConfig)
	secretsPath := os.Getenv(envSecrets)

	if configPath == "" {
		flag.StringVar(&configPath, flagConfig, "", "microservice config file")
	}

	if secretsPath == "" {
		flag.StringVar(&secretsPath, flagSecrets, "", "microservice secrets file")
	}

	flag.Parse()

	return &appFlags{
		secretsPath: secretsPath,
		configPath:  configPath,
	}
}

type (
	Config interface {
		LoggerConfig() *logger.Config
		TracingConfig() *TracingConfig
		ProbesConfig() *probes.ProbeCfg
	}
)

const (
	yaml = "yaml"
)

func readConfigFile[AppConfig any](
	path string,
	v *validator.Validate,
) (config AppConfig, err error) {
	viper.SetConfigType(yaml)
	viper.SetConfigFile(path)

	if err = viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("fail to read %s: %w", path, err)
	}

	config = *new(AppConfig)

	if err = viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("fail to unmarshal %s: %w", path, err)
	}

	if err = v.Struct(config); err != nil {
		return config, fmt.Errorf("fail to validate %s: %w", path, err)
	}

	return
}

func initConfig[AppConfig Config](
	f *appFlags,
	v *validator.Validate,
) (config AppConfig, err error) {
	if f.configPath == "" {
		return config, ErrEmptyConfigPath
	}

	return readConfigFile[AppConfig](f.configPath, v)
}

func initSecrets[Secrets any](
	f *appFlags,
	v *validator.Validate,
) (secrets Secrets, err error) {
	if f.secretsPath == "" {
		return secrets, ErrEmptySecretsPath
	}

	return readConfigFile[Secrets](f.secretsPath, v)
}

func WithSecrets[Secrets any](decorator interface{}) fx.Option {
	return fx.Options(
		fx.Decorate(decorator),
		fx.Provide(
			initSecrets[Secrets],
		),
	)
}

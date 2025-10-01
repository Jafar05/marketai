package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Config struct {
		ServiceName string       `mapstructure:"serviceName" validate:"required"`
		LogLevel    string       `mapstructure:"level"`
		DevMode     bool         `mapstructure:"devMode"`
		NoOp        bool         `mapstructure:"noOp"`
		KafkaLogger *KafkaConfig `mapstructure:"kafkaLogger"`
	}

	KafkaConfig struct {
		Brokers      []string `mapstructure:"brokers" validate:"required"`
		Topic        string   `mapstructure:"topic" validate:"required"`
		KafkaVersion string   `mapstructure:"version"`
		Namespace    string   `mapstructure:"ns" validate:"required"`
		Debug        bool     `mapstructure:"debug"`
	}

	AppParams struct {
		fx.Out

		Logger    *zap.Logger
		Sugar     *zap.SugaredLogger
		AppLogger AppLog
	}
)

func buildLogger(
	c *Config,
	loggerConfig zap.Config,
	lc fx.Lifecycle,
) (l *zap.Logger, err error) {

	if c.KafkaLogger == nil {
		return loggerConfig.Build()
	}

	consoleConfig := zapConfig(c.DevMode)

	if c.KafkaLogger.Debug {
		consoleConfig.Level.SetLevel(zap.DebugLevel)
	}

	console, err := consoleConfig.Build()

	if err != nil {
		return nil, err
	}

	kafka, err := newKafkaWriter(c.KafkaLogger, console)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.StopHook(kafka.Close))

	writer := zapcore.AddSync(kafka)

	return loggerConfig.Build(zap.WrapCore(
		func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(
				core,
				zapcore.NewCore(
					zapcore.NewJSONEncoder(loggerConfig.EncoderConfig),
					writer,
					loggerConfig.Level,
				).With([]zap.Field{
					zap.String("ns", c.KafkaLogger.Namespace),
				}),
			)
		},
	))

}

func InitLogger(c *Config, lc fx.Lifecycle) (AppParams, error) {
	if c.NoOp {
		noOp := zap.NewNop()
		return AppParams{
			Logger: noOp,
			Sugar:  noOp.Sugar(),
		}, nil
	}

	z := zapConfig(c.DevMode)

	lvl, err := zapcore.ParseLevel(level(c.LogLevel, c.DevMode))
	if err != nil {
		return AppParams{}, err
	}

	z.Level.SetLevel(lvl)

	res, err := buildLogger(c, z, lc)
	if err != nil {
		return AppParams{}, err
	}

	if c.ServiceName != "" {
		res = res.Named(c.ServiceName)
	}

	return AppParams{
		Logger:    res,
		Sugar:     res.Sugar(),
		AppLogger: NewAppLogger(res),
	}, nil
}

func zapConfig(devMode bool) (z zap.Config) {

	defer func() {
		z.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		z.EncoderConfig.MessageKey = "message"
	}()

	if devMode {
		return zap.NewDevelopmentConfig()
	}

	return zap.NewProductionConfig()
}

func level(l string, devMode bool) string {
	if l == "" && devMode {
		return zapcore.DebugLevel.String()
	}

	if l == "" {
		return zapcore.InfoLevel.String()
	}

	return l
}

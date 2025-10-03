package bootstrap

import (
	"marketai/pkg/logger"
	"marketai/pkg/probes"

	"github.com/go-playground/validator"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func AppOptions[AppConfig Config](options ...fx.Option) fx.Option {
	return fx.Options(
		basicProvides[AppConfig](),
		fx.Options(commonOptions()),
		fx.Options(options...),
		fx.Invoke(probes.Invoke),
	)

}

func commonOptions() fx.Option {
	return fx.Options(
		fx.Invoke(
			loggerInfo,
			adjustMaxprocs,
			invokeOtel,
		),
		fx.WithLogger(withLogger),
	)
}

func loggerConfig[AppConfig Config](c AppConfig) *logger.Config {
	return c.LoggerConfig()
}

func probesConfig[AppConfig Config](c AppConfig) *probes.ProbeCfg {
	return c.ProbesConfig()
}

func tracingConfig[AppConfig Config](
	c AppConfig,
) *TracingConfig {
	return c.TracingConfig()
}

func basicProvides[AppConfig Config]() fx.Option {
	return fx.Provide(
		validator.New,
		initFlags,
		initConfig[AppConfig],
		loggerConfig[AppConfig],
		logger.InitLogger,
		tracingConfig[AppConfig],
		withTraceProvider,
		probesConfig[AppConfig],
		probes.WithProbes,
	)
}

func loggerInfo(logger *zap.Logger) {
	logger.Info("logger ready")
}

func withLogger(log *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: log}
}

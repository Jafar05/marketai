package bootstrap

import (
	"marketai/pkg/http"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	maxHeaderBytes = 1 << 20
	bodyLimit      = "2M"
	readTimeout    = 15 * time.Second
)

type (
	HttpConfig struct {
		Port               string        `mapstructure:"port" validate:"required"`
		ApiBasePath        string        `mapstructure:"apiBasePath"`
		ReadTimeout        time.Duration `mapstructure:"readTimeout"`
		WriteTimeout       time.Duration `mapstructure:"writeTimeout"`
		BodyLimitSkipPaths []string      `mapstructure:"bodyLimitSkipPaths"`
	}

	GetHttpConfig interface {
		HttpConfig() *HttpConfig
	}

	HttpParams struct {
		fx.In

		Lifecycle  fx.Lifecycle
		Config     *HttpConfig
		Logger     *zap.Logger
		Echo       *echo.Echo
		Shutdowner fx.Shutdowner
	}

	withEchoParams struct {
		fx.In

		Config         *HttpConfig
		Logger         *zap.Logger
		TracerProvider trace.TracerProvider `optional:"true"`
	}
)

func httpConfig[Config GetHttpConfig](c Config) *HttpConfig {
	return c.HttpConfig()
}

func WithEcho[Config GetHttpConfig](initRouting ...interface{}) fx.Option {
	if len(initRouting) == 0 {
		return fx.Error(ErrNoEchoRouting)
	}

	return fx.Options(
		fx.Provide(
			httpConfig[Config],
			withEcho,
		),
		fx.Invoke(initRouting...),
		fx.Invoke(invokeEcho),
	)
}

// DeprecatedDirectEchoNew нужно для странных нетиповых сценариев где надо
// по не очевидной причине поднимать еще один инстанс http сервера на
// отдельном порту. Прежде чем использовать этот вызов, убедитесь, что нет
// других альтернативных возможностей реализации
func DeprecatedDirectEchoNew(
	c *HttpConfig,
	logger *zap.Logger,
	tp trace.TracerProvider,
) *echo.Echo {

	return newEcho(c, logger, tp)
}

// DeprecatedDirectInvokeEcho нужно для странных нетиповых сценариев где надо
// по не очевидной причине поднимать еще один инстанс http сервера на
// отдельном порту. Прежде чем использовать этот вызов, убедитесь, что нет
// других альтернативных возможностей реализации
func DeprecatedDirectInvokeEcho(p HttpParams) {
	invokeEcho(p)
}

func useTraceProvider(
	e *echo.Echo,
	logger *zap.Logger,
	tp trace.TracerProvider,
) {

	if tp == nil {
		logger.Warn("No trace provider found, echo will not propagate traces")
		return
	}

	logger.Info("use trace provider to propagate traces")

	e.Use(otelecho.Middleware("", otelecho.WithTracerProvider(tp)))
}

func withEcho(p withEchoParams) *echo.Echo {
	return newEcho(p.Config, p.Logger, p.TracerProvider)
}

func newEcho(
	c *HttpConfig,
	logger *zap.Logger,
	tp trace.TracerProvider,
) *echo.Echo {

	e := echo.New()

	if len(c.BodyLimitSkipPaths) > 0 {
		e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
			Limit: bodyLimit,
			Skipper: func(ctx echo.Context) bool {
				return slices.ContainsFunc(c.BodyLimitSkipPaths, func(path string) bool {
					return strings.HasPrefix(ctx.Request().URL.Path, path)
				})
			},
		}))
	} else {
		e.Use(middleware.BodyLimit(bodyLimit))
	}

	useTraceProvider(e, logger, tp)

	// по умолчанию ReadTimeout ставим 15 секунд, см readTimeout
	e.Server.ReadTimeout = readTimeout
	if c.ReadTimeout > 0 {
		e.Server.ReadTimeout = time.Second * c.ReadTimeout
	}
	logger.Info("read timeout", zap.Duration("val", e.Server.ReadTimeout))

	// по умолчанию WriteTimeout ставим ставим в 0 те бесконечность
	e.Server.WriteTimeout = 0
	if c.WriteTimeout > 0 {
		e.Server.WriteTimeout = time.Second * c.WriteTimeout
	}
	logger.Info("write timeout", zap.Duration("val", e.Server.WriteTimeout))

	e.Server.MaxHeaderBytes = maxHeaderBytes
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = http.GetErrorHandler(logger)

	return e
}

func invokeEcho(p HttpParams) {
	s := http.NewEchoFx(
		p.Shutdowner,
		p.Echo,
		p.Logger,
		p.Config.Port,
	)

	p.Lifecycle.Append(fx.StartStopHook(
		s.Start,
		p.Echo.Shutdown,
	))
}

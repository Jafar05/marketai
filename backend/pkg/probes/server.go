package probes

import (
	"context"
	"time"

	"github.com/Jafar05/pkg/http"

	"github.com/heptiolabs/healthcheck"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	Check struct {
		Name string
		Func healthcheck.Check
	}

	ProbeParams struct {
		fx.In

		Lifecycle fx.Lifecycle
		Config    *ProbeCfg
		Logger    *zap.Logger
		Echo      *echo.Echo `optional:"true"`
	}

	Probe struct {
		fx.Out

		Echo    *echo.Echo `name:"echoProbes"`
		Handler healthcheck.Handler
		Context context.Context `name:"probeContext"`
	}

	InvokeParams struct {
		fx.In

		Lifecycle   fx.Lifecycle
		Shutdowner  fx.Shutdowner
		Logger      *zap.Logger
		Config      *ProbeCfg
		Echo        *echo.Echo `name:"echoProbes"`
		LiveChecks  []*Check   `group:"liveChecks"`
		ReadyChecks []*Check   `group:"readyChecks"`
		Handler     healthcheck.Handler
	}
)

const (
	ContextAnnotation = `name:"probeContext"`
)

const (
	writeTimeout = 15 * time.Second
	readTimeout  = 15 * time.Second
)

const (
	bodyLimit = "2M"
)

func WithLiveCheckCtx(newCheck interface{}) fx.Option {
	return WithLiveCheck(
		newCheck,
		fx.ParamTags(ContextAnnotation),
	)
}

func WithReadyCheckCtx(newCheck interface{}) fx.Option {
	return WithReadyCheck(
		newCheck,
		fx.ParamTags(ContextAnnotation),
	)
}

func WithLiveCheck(newCheck interface{}, a ...fx.Annotation) fx.Option {
	return fx.Provide(
		fx.Annotate(
			newCheck,
			append(a, fx.ResultTags(`group:"liveChecks"`))...,
		),
	)

}

func WithReadyCheck(newCheck interface{}, a ...fx.Annotation) fx.Option {
	return fx.Provide(
		fx.Annotate(
			newCheck,
			append(a, fx.ResultTags(`group:"readyChecks"`))...,
		),
	)

}

func WithProbes(p ProbeParams) Probe {
	ctx, cancel := context.WithCancel(context.Background())

	p.Lifecycle.Append(fx.StopHook(cancel))

	if p.Echo != nil {
		p.Echo.Use(echoprometheus.NewMiddleware("echo"))
		p.Logger.Info("using main echo server")
	}

	if p.Echo == nil {
		p.Logger.Info("standalone")
	}

	metrics := echo.New()

	metrics.HideBanner = true
	metrics.HidePort = true

	metrics.Server.ReadTimeout = readTimeout
	metrics.Server.WriteTimeout = writeTimeout

	metrics.Use(middleware.BodyLimit(bodyLimit))

	metrics.GET(p.Config.PrometheusPath, echoprometheus.NewHandler())

	handler := healthcheck.NewHandler()

	metrics.GET(p.Config.LivenessPath, func(c echo.Context) error {
		handler.LiveEndpoint(c.Response(), c.Request())
		return nil
	})

	metrics.GET(p.Config.ReadinessPath, func(c echo.Context) error {
		handler.ReadyEndpoint(c.Response(), c.Request())
		return nil
	})

	return Probe{
		Echo:    metrics,
		Handler: handler,
		Context: ctx,
	}
}

func Invoke(p InvokeParams) {
	s := http.NewEchoFx(
		p.Shutdowner,
		p.Echo,
		p.Logger,
		p.Config.Port,
	)

	addLiveChecks(p)
	addReadyChecks(p)

	p.Lifecycle.Append(fx.StartStopHook(
		s.Start,
		p.Echo.Shutdown,
	))

}

func addLiveChecks(p InvokeParams) {
	if len(p.LiveChecks) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	d := time.Duration(p.Config.CheckIntervalSeconds) * time.Second

	for _, c := range p.LiveChecks {
		p.Handler.AddLivenessCheck(
			c.Name,
			healthcheck.AsyncWithContext(ctx, c.Func, d),
		)
		p.Logger.Info("add live check", zap.String("name", c.Name))
	}

	p.Lifecycle.Append(fx.StopHook(cancel))
}

func addReadyChecks(p InvokeParams) {
	if len(p.ReadyChecks) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	d := time.Duration(p.Config.CheckIntervalSeconds) * time.Second

	for _, c := range p.ReadyChecks {
		p.Handler.AddReadinessCheck(
			c.Name,
			healthcheck.AsyncWithContext(ctx, c.Func, d),
		)
		p.Logger.Info("add ready check", zap.String("name", c.Name))
	}

	p.Lifecycle.Append(fx.StopHook(cancel))
}

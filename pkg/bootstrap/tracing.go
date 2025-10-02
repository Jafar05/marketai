package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/zapr"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pkgGrpc "github.com/Jafar05/pkg/grpc"
)

const (
	TracerName = "bitbucket.pcbltools.ru/bitbucket/scm/edupower/eduterm.git/pkg/app"
)

type (
	TracingConfig struct {
		Enable      bool   `mapstructure:"enable"`
		ServiceName string `mapstructure:"serviceName"`
		HostPort    string `mapstructure:"hostPort"`
		LogSpans    bool   `mapstructure:"logSpans"`
	}

	TracingParams struct {
		fx.In

		Config    *TracingConfig
		Logger    *zap.Logger
		Lifecycle fx.Lifecycle
	}
)

func withTraceProvider(p TracingParams) (trace.TracerProvider, error) {

	logger := p.Logger.Named("otel")

	otel.SetLogger(zapr.NewLogger(logger))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
		logger.Info("internal", zap.Error(cause))
	}))

	if !p.Config.Enable {
		p.Logger.Info("tracing is disabled")

		return noop.NewTracerProvider(), nil
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName(p.Config.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(
		ctx,
		time.Second,
	)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		p.Config.HostPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(
				//interceptorLogger(p.Logger),
				pkgGrpc.InterceptorLogger(p.Logger),
				logging.WithLevels(pkgGrpc.CodeToLevel),
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create gRPC connection to collector: %w",
			err,
		)
	}

	// Set up a trace exporter
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithGRPCConn(conn),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	p.Lifecycle.Append(fx.StopHook(exporter.Shutdown))

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	p.Lifecycle.Append(fx.StopHook(tp.Shutdown))

	return tp, nil

}

func invokeOtel(p trace.TracerProvider) {

	otel.SetTracerProvider(p)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
			b3.New(
				b3.WithInjectEncoding(b3.B3MultipleHeader),
			),
		),
	)
}

package grpc

import (
	"context"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/trace"
)

type (
	ServerConfig struct {
		Port string `mapstructure:"port" validate:"required"`
	}

	GetServerConfig interface {
		GrpcConfig() *ServerConfig
	}
)

// see https://github.com/grpc-ecosystem/go-grpc-middleware/blob/71d7422112b1d7fadd4b8bf12a6f33ba6d22e98e/examples/server/main.go
func server(l *zap.Logger) (*grpc.Server, *grpcprom.ServerMetrics) {

	buckets := []float64{
		0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120,
	}

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets(buckets),
		),
	)

	prometheus.MustRegister(srvMetrics)

	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}

		return nil
	}

	// Setup metric for panic recoveries.
	panicsTotal := promauto.With(
		prometheus.DefaultRegisterer,
	).NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})

	grpcPanicRecoveryHandler := func(info any) (err error) {
		panicsTotal.Inc()
		l.Error("recovered from panic")
		return status.Errorf(codes.Internal, "%s", info)
	}

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			// open telemetry interceptor
			srvMetrics.UnaryServerInterceptor(
				grpcprom.WithExemplarFromContext(exemplarFromContext),
			),
			logging.UnaryServerInterceptor(
				InterceptorLogger(l),
				logging.WithFieldsFromContext(logTraceID),
				logging.WithLevels(CodeToLevel),
			),
			recovery.UnaryServerInterceptor(
				recovery.WithRecoveryHandler(grpcPanicRecoveryHandler),
			),
		),
	)

	//	srvMetrics.InitializeMetrics(srv)

	return srv, srvMetrics
}

func initMetrics(srv *grpc.Server, metrics *grpcprom.ServerMetrics) {
	metrics.InitializeMetrics(srv)
}

func invokeServer(
	srv *grpc.Server,
	listener net.Listener,
	logger *zap.Logger,
	lc fx.Lifecycle,
	shutdowner fx.Shutdowner,
) {
	lc.Append(fx.StartStopHook(
		func() {
			go func() {
				err := srv.Serve(listener)
				if err == nil {
					logger.Info("server stopped")
					return
				}

				logger.Error("server error", zap.Error(err))

				if err := shutdowner.Shutdown(); err != nil {
					logger.Error("unable to shutdown", zap.Error(err))
				}
			}()
		},
		func() {
			srv.GracefulStop()
			srv.Stop()
		},
	))
}

func serverConfig[Config GetServerConfig](c Config) *ServerConfig {
	return c.GrpcConfig()
}

func WithListener() fx.Option {
	return fx.Provide(
		func(config *ServerConfig) (net.Listener, error) {
			return net.Listen("tcp", config.Port)
		},
	)
}

func WithRegisterFuncGrpcServer(f interface{}) fx.Option {
	return fx.Invoke(
		f,
	)
}

func WithRegisterFunc(f interface{}) fx.Option {
	return fx.Invoke(
		fx.Annotate(
			f,
			fx.From(new(*grpc.Server)),
		),
	)
}

func Server[Config GetServerConfig](
	opts ...fx.Option,
) fx.Option {
	return fx.Options(
		fx.Provide(
			serverConfig[Config],
			server,
		),
		fx.Options(opts...),
		fx.Invoke(initMetrics),
		fx.Invoke(invokeServer),
	)
}

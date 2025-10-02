package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNoClientConnectionFound = errors.New("")
)

const (
	defaultMaxTimeout  = time.Minute * 5
	fxGrpcConnAnnotate = `name:"grpcClientConn%s"`
)

type (
	ClientConfig struct {
		Target string `mapstructure:"port" validate:"required,gt=0"`
		// if multiple services bind to same address use port names
		Name           string        `mapstructure:"name" validate:"required,gt=0"`
		RequestTimeout time.Duration `mapstructure:"timeout"`
	}

	ClientConnections struct {
		Map map[string]*grpc.ClientConn
	}

	GetClientConfig interface {
		GrpcClientConfig() []*ClientConfig
	}
)

func clients(
	logger *zap.Logger,
	config GetClientConfig,
	opts ...grpc.DialOption,
) (*ClientConnections, error) {
	connections := make(map[string]*grpc.ClientConn)

	for _, configItem := range config.GrpcClientConfig() {
		client, err := grpcClient(logger, configItem.Target, opts...)
		if err != nil {
			return nil, err
		}
		connections[configItem.Name] = client
	}

	return &ClientConnections{
		Map: connections,
	}, nil

}

func clientConfig[Config GetClientConfig](
	c Config,
) GetClientConfig {
	return c
}

func stopClients(
	lc fx.Lifecycle,
	log *zap.Logger,
	conns *ClientConnections,
) {

	lc.Append(
		fx.StopHook(func() {
			const msg = "grpc client conn close"

			for name, conn := range conns.Map {
				err := conn.Close()
				if err != nil {
					log.Info(msg, zap.String("name", name), zap.Error(err))
					continue
				}

				log.Info(msg, zap.String("name", name))
			}
		}),
	)

}

func Client[Config GetClientConfig](opts ...fx.Option) fx.Option {
	return fx.Options(
		fx.Provide(
			clientConfig[Config],
		),
		fx.Provide(
			fx.Annotate(
				clients,
				fx.ParamTags("", "", `group:"grpcDialOptions"`),
			),
		),
		fx.Options(opts...),
		fx.Invoke(
			stopClients,
		),
	)
}

func WithDialOption(opt grpc.DialOption) fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() grpc.DialOption {
					return opt
				},
				fx.ResultTags(`group:"grpcDialOptions"`),
			),
		),
	)
}

func WithNamedService(name string, newService interface{}) fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func(
					connections *ClientConnections,
				) (grpc.ClientConnInterface, error) {

					conn, ok := connections.Map[name]

					if !ok {
						return nil, fmt.Errorf(
							"%w: name=%s",
							ErrNoClientConnectionFound,
							name,
						)
					}

					return conn, nil
				},

				fx.ResultTags(fmt.Sprintf(fxGrpcConnAnnotate, name)),
			),
		),
		fx.Provide(
			fx.Annotate(
				newService,
				fx.ParamTags(fmt.Sprintf(fxGrpcConnAnnotate, name)),
			),
		),
	)
}

func grpcClient(
	logger *zap.Logger,
	port string,
	opts ...grpc.DialOption,
) (*grpc.ClientConn, error) {

	buckets := []float64{
		0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120,
	}

	reg := prometheus.NewRegistry()
	clMetrics := grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets(buckets),
		),
	)

	reg.MustRegister(clMetrics)
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	return grpc.Dial(
		port,
		append(
			opts,
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithChainUnaryInterceptor(
				timeout.UnaryClientInterceptor(defaultMaxTimeout),
				clMetrics.UnaryClientInterceptor(
					grpcprom.WithExemplarFromContext(exemplarFromContext),
				),
				logging.UnaryClientInterceptor(
					InterceptorLogger(logger),
					logging.WithFieldsFromContext(logTraceID),
					logging.WithLevels(CodeToLevel),
				),
			),
		)...,
	)
}

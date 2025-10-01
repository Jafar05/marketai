package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
)

type (
	decorateMessage struct {
		method string
		err    string
	}
)

func (e *decorateMessage) withField(key, val string) {

	switch key {
	case "grpc.method":
		e.method = val
	case "grpc.error":
		e.err = val
	}

}

func (e *decorateMessage) format(msg string) string {

	if e.err == "" && e.method == "" {
		return msg
	}

	if e.err == "" {
		return fmt.Sprintf("%s %s", e.method, msg)
	}

	return fmt.Sprintf("%s %s with error: err=%s", e.method, msg, e.err)

}

func logTraceID(ctx context.Context) logging.Fields {
	if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
		return logging.Fields{"traceID", span.TraceID().String()}
	}
	return nil
}

func CodeToLevel(code codes.Code) logging.Level {
	if code == codes.OK {
		return logging.LevelDebug
	}

	return logging.LevelError
}

func checkLevel(lvl logging.Level, zapLvl zapcore.Level) bool {
	if lvl == logging.LevelError {
		return false
	}

	return zapLvl != zap.DebugLevel

}

func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(
		ctx context.Context,
		lvl logging.Level,
		msg string,
		fields ...any,
	) {

		if checkLevel(lvl, l.Level()) {
			return
		}

		f := make([]zap.Field, 0, len(fields)/2)

		var d decorateMessage

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]
			f = append(f, zap.Any(key.(string), value))

			d.withField(key.(string), value.(string))
		}

		logger := l.WithOptions(zap.AddCallerSkip(2)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(d.format(msg))
		// case logging.LevelInfo:
		// 	logger.Info(d.format(msg))
		// case logging.LevelWarn:
		// 	logger.Warn(d.format(msg))
		case logging.LevelError:
			logger.Error(d.format(msg))
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

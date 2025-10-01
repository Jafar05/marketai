package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type (
	appLogger struct {
		base      *zap.Logger
		baseSugar *zap.SugaredLogger
	}
)

func NewNop() AppLog {
	return NewAppLogger(zap.NewNop())
}

func NewAppLogger(base *zap.Logger) AppLog {
	log := base.WithOptions(zap.AddCallerSkip(1))
	return &appLogger{
		base:      log,
		baseSugar: log.Sugar(),
	}
}

func (l *appLogger) Info(msg string) {
	l.base.Info(msg)
}

func (l *appLogger) Debug(msg string) {
	l.base.Debug(msg)
}

func (l *appLogger) Warn(msg string) {
	l.base.Warn(msg)
}

func (l *appLogger) withErrorMsg(msg string, err error) string {
	if err == nil {
		return msg
	}

	return fmt.Sprintf("%s: err=%s", msg, err.Error())
}

func (l *appLogger) InfoWithError(msg string, err error) {
	l.base.Info(l.withErrorMsg(msg, err), zap.Error(err))
}

func (l *appLogger) DebugWithError(msg string, err error) {

	if l.base.Level() != zap.DebugLevel {
		return
	}

	l.base.Debug(l.withErrorMsg(msg, err), zap.Error(err))
}

func (l *appLogger) WarnWithError(msg string, err error) {
	l.base.Warn(l.withErrorMsg(msg, err), zap.Error(err))
}

func (l *appLogger) Error(msg string, err error) {
	l.base.Error(l.withErrorMsg(msg, err), zap.Error(err))
}

func (l *appLogger) Infof(template string, args ...interface{}) {
	l.baseSugar.Infof(template, args...)
}

func (l *appLogger) Debugf(template string, args ...interface{}) {

	if l.base.Level() != zap.DebugLevel {
		return
	}

	l.baseSugar.Debugf(template, args...)
}

func (l *appLogger) Warnf(template string, args ...interface{}) {
	l.baseSugar.Warnf(template, args...)
}

func (l *appLogger) InfofWithError(
	template string,
	err error,
	args ...interface{},
) {

	l.base.Info(
		l.withErrorMsg(fmt.Sprintf(template, args...), err),
		zap.Error(err),
	)

}

func (l *appLogger) DebugfWithError(
	template string,
	err error,
	args ...interface{},
) {

	if l.base.Level() != zap.DebugLevel {
		return
	}

	l.base.Debug(
		l.withErrorMsg(fmt.Sprintf(template, args...), err),
		zap.Error(err),
	)

}

func (l *appLogger) WarnfWithError(template string, err error, args ...interface{}) {

	l.base.Warn(
		l.withErrorMsg(fmt.Sprintf(template, args...), err),
		zap.Error(err),
	)

}

func (l *appLogger) Errorf(template string, err error, args ...interface{}) {

	l.base.Error(
		l.withErrorMsg(fmt.Sprintf(template, args...), err),
		zap.Error(err),
	)

}

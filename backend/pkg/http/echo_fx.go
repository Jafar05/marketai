package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	EchoFx struct {
		shutdowner fx.Shutdowner
		echo       *echo.Echo
		logger     *zap.Logger
		port       string
	}
)

func NewEchoFx(
	shutdowner fx.Shutdowner,
	echo *echo.Echo,
	logger *zap.Logger,
	port string,
) *EchoFx {
	return &EchoFx{
		shutdowner: shutdowner,
		echo:       echo,
		logger:     logger,
		port:       port,
	}
}

func (s *EchoFx) Start() {
	go s.run()
}

func (s *EchoFx) run() {
	err := s.echo.Start(s.port)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {

		s.logger.Error("unexpected error", zap.Error(err))

		// do immediate shutdown if unexpected error occur
		shutdownErr := s.shutdowner.Shutdown()

		if shutdownErr != nil {
			s.logger.Error(
				"failed to shutdown",
				zap.Error(shutdownErr),
			)
		}
	}

	s.logger.Info("server stopped")
}

func httpError(err error) *echo.HTTPError {
	he, ok := err.(*echo.HTTPError)

	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				return herr
			}
		}

		return he
	}

	return echo.ErrInternalServerError.WithInternal(err)

}

func statusMessage(msg interface{}) interface{} {

	switch m := msg.(type) {
	case string:
		return msg
	case json.Marshaler:
		return msg
	case error:
		return m.Error()
	}

	return msg
}

// GetErrorHandler returns modified copy of echo default error handler
// returns error message echo.Map{"message": m, "error": err.Error()} to client
func GetErrorHandler(logger *zap.Logger) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		he := httpError(err)

		if c.Request().Method == http.MethodHead {
			c.NoContent(he.Code)
			return
		}

		if he.Internal == nil {

			c.JSON(he.Code, echo.Map{
				"statusMessage": statusMessage(he.Message),
				"timestamp":     time.Now(),
			})

			return
		}

		c.JSON(he.Code, echo.Map{
			"message":       he.Internal.Error(),
			"statusMessage": statusMessage(he.Message),
			"timestamp":     time.Now(),
		})

	}
}

package ports

import (
	"fmt"
	"log"
	"marketai/auth/internal/app"
	"marketai/auth/internal/app/dto"
	"marketai/auth/internal/config"
	"marketai/auth/internal/domain"
	"marketai/pkg/logger"
	"marketai/pkgAuth/jwt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
)

type httpServer struct {
	fx.In

	Config    *config.Config
	Echo      *echo.Echo
	Logger    logger.AppLog
	Validator *validator.Validate
}

// UserContextKey - ключ для хранения информации о пользователе в контексте запроса.
type UserContextKey string

func registerRoutes(s httpServer, a *app.AppCQRS) {
	s.Echo.Use(middleware.CORS())
	withAuth := s.Echo.Group(s.Config.Http.ApiBasePath)

	withAuth.Add(http.MethodPost, "/login", s.loginHandler(a))
	withAuth.Add(http.MethodPost, "/register", s.registerHandler(a))

	withAuth.Add(http.MethodPost, "/validate", s.validateTokenHandler())

}

// @Summary		Аутентификация пользователя
// @Description	Вход пользователя в систему и получение JWT токена.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			input	body		dto.LoginCommand	true	"Данные для входа"
// @Success		200		{object}	map[string]string	"Успешный вход, возвращает JWT токен"
// @Failure		400		{string}	string				"Неверный формат запроса"
// @Failure		401		{string}	string				"Неверные учетные данные"
// @Router			/login [post]
func (rc *httpServer) loginHandler(a *app.AppCQRS) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var req dto.LoginCommand
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
		}

		result, err := a.Queries.Login.Handle(ctx, req)
		if err != nil {
			log.Printf("Ошибка при входе пользователя %s: %v", req.Email, err)
			return echo.NewHTTPError(http.StatusBadRequest, "Неверные учетные данные")
		}

		return c.JSON(http.StatusOK, map[string]string{"token": result.Token})
	}
}

// @Summary		Регистрация нового пользователя
// @Description	Регистрация нового пользователя в системе.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			input	body		dto.RegisterUserCommand	true	"Данные для регистрации"
// @Success		200		{object}	map[string]string		"Пользователь успешно зарегистрирован"
// @Failure		400		{string}	string					"Неверный формат запроса"
// @Failure		409		{string}	string					"Пользователь с таким именем уже существует"
// @Router			/register [post]
func (rc *httpServer) registerHandler(a *app.AppCQRS) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var req domain.User

		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
		}

		if req.Role == "" {
			req.Role = "user"
		}

		result, err := a.Commands.Register.Handle(ctx, req)
		if err != nil {
			log.Printf("Ошибка при регистрации пользователя %s: %v", req.Email, err)
			return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("ошибка: %s", err.Error()), err)
		}

		return c.JSON(http.StatusOK, map[string]string{"user_id": result.UserID, "message": "Пользователь успешно зарегистрирован"})
	}
}

func (rc *httpServer) validateTokenHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.ValidateTokenRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "неверный формат запроса")
		}

		claims, err := jwt.ValidateToken(req.Token, rc.Config.JWTSecret)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "неверный токен")
		}

		response := dto.ValidateTokenResponse{
			Valid:  true,
			UserID: claims.UserID,
			Role:   claims.Role,
		}

		return c.JSON(http.StatusOK, response)
	}
}

package ports

import (
	"context"
	"log"
	"marketai/cards/internal/app"
	"marketai/cards/internal/app/command"
	"marketai/cards/internal/app/dto"
	"marketai/cards/internal/app/query"
	"marketai/cards/internal/config"
	"marketai/cards/internal/domain"
	"marketai/pkg/logger"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
)

type httpServer struct {
	fx.In

	Config      *config.Config
	Echo        *echo.Echo
	Logger      logger.AppLog
	Validator   *validator.Validate
	App         *app.AppCQRS
	AuthService domain.AuthService
}

type HTTPServer struct {
	echo *echo.Echo
	cfg  *config.Config
}

func NewHTTPServer(srv httpServer) *HTTPServer {
	registerRoutes(srv, srv.App, srv.AuthService)
	return &HTTPServer{
		echo: srv.Echo,
		cfg:  srv.Config,
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	return s.echo.Start(":" + s.cfg.Http.Port)
}

func registerRoutes(s httpServer, a *app.AppCQRS, authService domain.AuthService) {
	s.Echo.Use(middleware.CORS())
	s.Echo.Use(middleware.Logger())
	s.Echo.Use(middleware.Recover())

	// Middleware для проверки JWT токена
	//authMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
	//	return func(c echo.Context) error {
	//		authHeader := c.Request().Header.Get("Authorization")
	//		if authHeader == "" {
	//			return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
	//		}
	//
	//		token := strings.TrimPrefix(authHeader, "Bearer ")
	//		if token == authHeader {
	//			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format")
	//		}
	//
	//		userInfo, err := authService.ValidateToken(c.Request().Context(), token)
	//		if err != nil {
	//			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	//		}
	//
	//		c.Set("user_id", userInfo.UserID)
	//		c.Set("user_role", userInfo.Role)
	//
	//		return next(c)
	//	}
	//}

	api := s.Echo.Group(s.Config.Http.ApiBasePath)
	//api.Use(authMiddleware) // Временно отключено для тестирования
	//s.Echo.POST("/generate", s.generateCardHandler(a))
	api.POST("/generate", s.generateCardHandler(a))
	api.GET("/history", s.getCardsHistoryHandler(a))
	api.GET("/:id", s.getCardByIDHandler(a))
}

// @Summary		Генерация карточки товара
// @Description	Генерирует карточку товара на основе фото и описания с помощью AI
// @Tags			cards
// @Accept			json
// @Produce		json
// @Param			input	body		dto.GenerateCardRequest	true	"Данные для генерации карточки"
// @Success		200		{object}	dto.GenerateCardResponse	"Карточка успешно сгенерирована"
// @Failure		400		{string}	string					"Неверный формат запроса"
// @Failure		401		{string}	string					"Неавторизованный доступ"
// @Router			/generate [post]
func (rc *httpServer) generateCardHandler(a *app.AppCQRS) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var req dto.GenerateCardRequest

		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
		}

		if err := rc.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Неверные данные запроса")
		}

		userID := "test-user" // Временно для тестирования

		result, err := a.Commands.GenerateCard.Handle(ctx, command.GenerateCardCommand{
			UserID:           userID,
			PhotoURL:         req.PhotoURL,
			ShortDescription: req.ShortDescription,
		})
		if err != nil {
			log.Printf("Ошибка при генерации карточки для пользователя %s: %v", userID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка при генерации карточки")
		}

		response := dto.GenerateCardResponse{
			ID:          result.Card.ID,
			Title:       result.Card.Title,
			Description: result.Card.Description,
			Tags:        result.Card.Tags,
			Image:       result.Card.Image,
		}

		return c.JSON(http.StatusOK, response)
	}
}

// @Summary		История карточек пользователя
// @Description	Возвращает список всех карточек пользователя
// @Tags			cards
// @Produce		json
// @Success		200		{object}	dto.CardHistoryResponse	"Список карточек"
// @Failure		401		{string}	string					"Неавторизованный доступ"
// @Router			/history [get]
func (rc *httpServer) getCardsHistoryHandler(a *app.AppCQRS) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		userID := "test-user" // Временно для тестирования

		result, err := a.Queries.GetCardsByUser.Handle(ctx, query.GetCardsByUserQuery{
			UserID: userID,
		})
		if err != nil {
			log.Printf("Ошибка при получении истории карточек для пользователя %s: %v", userID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка при получении истории")
		}

		var cards []dto.CardInfo
		for _, card := range result.Cards {
			cards = append(cards, dto.CardInfo{
				ID:               card.ID,
				PhotoURL:         card.PhotoURL,
				ShortDescription: card.ShortDescription,
				Title:            card.Title,
				Description:      card.Description,
				Tags:             card.Tags,
				Image:            card.Image,
				CreatedAt:        card.CreatedAt.Format(time.RFC3339),
			})
		}

		response := dto.CardHistoryResponse{Cards: cards}
		return c.JSON(http.StatusOK, response)
	}
}

// @Summary		Получение карточки по ID
// @Description	Возвращает детальную информацию о карточке
// @Tags			cards
// @Produce		json
// @Param			id	path		string	true	"ID карточки"
// @Success		200	{object}	dto.CardDetailResponse	"Детали карточки"
// @Failure		401	{string}	string					"Неавторизованный доступ"
// @Failure		404	{string}	string					"Карточка не найдена"
// @Router			/{id} [get]
func (rc *httpServer) getCardByIDHandler(a *app.AppCQRS) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		cardID := c.Param("id")

		result, err := a.Queries.GetCardByID.Handle(ctx, query.GetCardByIDQuery{
			CardID: cardID,
		})
		if err != nil {
			log.Printf("Ошибка при получении карточки %s: %v", cardID, err)
			return echo.NewHTTPError(http.StatusNotFound, "Карточка не найдена")
		}

		response := dto.CardDetailResponse{
			ID:               result.Card.ID,
			PhotoURL:         result.Card.PhotoURL,
			ShortDescription: result.Card.ShortDescription,
			Title:            result.Card.Title,
			Description:      result.Card.Description,
			Tags:             result.Card.Tags,
			Image:            result.Card.Image,
			CreatedAt:        result.Card.CreatedAt.Format(time.RFC3339),
		}

		return c.JSON(http.StatusOK, response)
	}
}

package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"marketai/auth/internal/app/dto"
	"marketai/auth/internal/config"
	domain "marketai/auth/internal/domain"
	"marketai/pkgAuth/jwt"
)

type LoginCommandResult struct {
	Token    string
	UserID   string
	FullName string
}

type LoginCommandHandlerResult struct {
	userRepo domain.UserRepository
	config   *config.Config
}

type LoginCommandHandler interface {
	Handle(ctx context.Context, cmd dto.LoginCommand) (*LoginCommandResult, error)
}

func NewLoginCommandHandler(userRepo domain.UserRepository, cfg *config.Config) *LoginCommandHandlerResult {
	return &LoginCommandHandlerResult{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (h *LoginCommandHandlerResult) Handle(ctx context.Context, cmd dto.LoginCommand) (*LoginCommandResult, error) {
	user, err := h.userRepo.GetUserByUsername(ctx, cmd.Email, cmd.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}
	if user == nil {
		return nil, errors.New("неверные учетные данные")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password))

	if err != nil {
		return nil, fmt.Errorf("неверные учетные данные: %w", err)
	}

	token, err := jwt.GenerateToken(user.ID, user.Role, h.config.JWTSecret, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("ошибка при генерации JWT токена: %w", err)
	}

	return &LoginCommandResult{
		Token:    token,
		UserID:   user.ID,
		FullName: user.FullName,
	}, nil
}

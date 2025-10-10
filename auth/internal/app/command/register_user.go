package command

import (
	"context"
	"errors"
	"fmt"
	"marketai/auth/internal/config"
	"time"

	"github.com/jackc/pgx/v5"

	"golang.org/x/crypto/bcrypt"

	domain "marketai/auth/internal/domain"
	"marketai/pkgAuth/jwt"
)

type RegisterUserCommandResult struct {
	UserID string
	Token  string
	User   *domain.User
}

type RegisterCommandHandler interface {
	Handle(ctx context.Context, cmd domain.User) (*RegisterUserCommandResult, error)
}

type registerUserCommandHandler struct {
	pgRepo domain.UserRepository
	cfg    *config.Config
}

func NewRegisterUserCommandHandler(
	userRepo domain.UserRepository,
	cfg *config.Config,
) *registerUserCommandHandler {
	return &registerUserCommandHandler{
		pgRepo: userRepo,
		cfg:    cfg,
	}
}

func (h *registerUserCommandHandler) Handle(ctx context.Context, cmd domain.User) (*RegisterUserCommandResult, error) {
	existingUser, err := h.pgRepo.GetUserByUsername(ctx, cmd.Email, cmd.PhoneNumber)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	newUser := &domain.User{
		FullName:     cmd.FullName,
		Email:        cmd.Email,
		PasswordHash: string(hashedPassword),
		PhoneNumber:  cmd.PhoneNumber,
		Role:         cmd.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.pgRepo.CreateUser(ctx, newUser); err != nil {
		return nil, fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	// Генерируем JWT токен для нового пользователя
	token, err := jwt.GenerateToken(newUser.ID, newUser.Role, h.cfg.JWTSecret, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("ошибка при генерации JWT токена: %w", err)
	}

	return &RegisterUserCommandResult{
		UserID: newUser.ID,
		Token:  token,
		User:   newUser,
	}, nil
}

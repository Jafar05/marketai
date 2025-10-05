package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	domain "marketai/auth/internal/domain"
)

type RegisterUserCommandResult struct {
	UserID string
}

type RegisterCommandHandler interface {
	Handle(ctx context.Context, cmd domain.User) (*RegisterUserCommandResult, error)
}

type registerUserCommandHandler struct {
	pgRepo domain.UserRepository
}

func NewRegisterUserCommandHandler(userRepo domain.UserRepository) *registerUserCommandHandler {
	return &registerUserCommandHandler{
		pgRepo: userRepo,
	}
}

func (h *registerUserCommandHandler) Handle(ctx context.Context, cmd domain.User) (*RegisterUserCommandResult, error) {

	existingUser, err := h.pgRepo.GetUserByUsername(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}

	if existingUser != nil {
		return nil, errors.New("пользователь с таким именем уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	newUser := &domain.User{
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

	return &RegisterUserCommandResult{UserID: newUser.ID}, nil
}

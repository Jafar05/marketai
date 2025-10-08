package command

import (
	"context"
	"fmt"
	"marketai/auth/internal/adapters/postgres"
)

type VerifyEmailCommandHandler interface {
	Handle(ctx context.Context, token string) (string, error)
}

type verifyEmailCommandHandler struct {
	emailRepo *postgres.EmailVerificationRepository
	userRepo  *postgres.AuthRepository
}

func NewVerifyEmailCommandHandler(
	emailRepo *postgres.EmailVerificationRepository,
	userRepo *postgres.AuthRepository,
) *verifyEmailCommandHandler {
	return &verifyEmailCommandHandler{
		emailRepo: emailRepo,
		userRepo:  userRepo,
	}
}

func (h *verifyEmailCommandHandler) Handle(ctx context.Context, token string) (string, error) {
	userID, err := h.emailRepo.GetUserByToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("неверный или просроченный токен")
	}

	if err := h.userRepo.MarkEmailVerified(ctx, userID); err != nil {
		return "", fmt.Errorf("не удалось подтвердить email: %w", err)
	}

	_ = h.emailRepo.DeleteToken(ctx, token)
	return userID, nil
}

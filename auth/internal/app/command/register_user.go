package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"gopkg.in/gomail.v2"
	"marketai/auth/internal/adapters/postgres"
	"marketai/auth/internal/config"
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
	pgRepo    domain.UserRepository
	emailRepo *postgres.EmailVerificationRepository
	cfg       *config.Config
}

func NewRegisterUserCommandHandler(
	userRepo domain.UserRepository,
	emailRepo *postgres.EmailVerificationRepository,
	cfg *config.Config,
) *registerUserCommandHandler {
	return &registerUserCommandHandler{
		pgRepo:    userRepo,
		emailRepo: emailRepo,
		cfg:       cfg,
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

	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := h.emailRepo.CreateToken(ctx, newUser.ID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании токена подтверждения: %w", err)
	}

	link := fmt.Sprintf("https://marketai-backend-production.up.railway.app/verify-email?token=%s", token)
	body := fmt.Sprintf("<h3>Подтвердите ваш email</h3><p><a href='%s'>Нажмите для подтверждения</a></p>", link)
	if err := SendEmail(
		h.cfg.SMTP.Host,
		h.cfg.SMTP.Port,
		h.cfg.SMTP.Username,
		h.cfg.SMTP.Password,
		h.cfg.SMTP.From,
		newUser.Email,
		"Подтверждение email",
		body,
	); err != nil {
		return nil, fmt.Errorf("не удалось отправить письмо подтверждения: %w", err)
	}

	return &RegisterUserCommandResult{UserID: newUser.ID}, nil
}

func SendEmail(host string, port int, username, password, from, to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(host, port, username, password)
	return d.DialAndSend(m)
}

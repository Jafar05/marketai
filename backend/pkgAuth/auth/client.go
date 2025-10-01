package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ValidateToken отправляет запрос к auth сервису для валидации токена
func (c *Client) ValidateToken(token string) (*ValidateTokenResponse, error) {
	reqBody := ValidateTokenRequest{
		Token: token,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/validate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Обрабатываем различные статус коды
	switch resp.StatusCode {
	case http.StatusOK:
		var response ValidateTokenResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
		}
		return &response, nil

	case http.StatusUnauthorized:
		return nil, fmt.Errorf("недействительный токен")

	case http.StatusBadRequest:
		return nil, fmt.Errorf("неверный формат запроса")

	default:
		return nil, fmt.Errorf("неуспешный статус ответа: %d, тело: %s", resp.StatusCode, string(body))
	}

}

// ValidateTokenWithRetry отправляет запрос с повторными попытками
func (c *Client) ValidateTokenWithRetry(token string, maxRetries int) (*ValidateTokenResponse, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		response, err := c.ValidateToken(token)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Если это последняя попытка, не ждем
		if i == maxRetries {
			break
		}

		// Экспоненциальная задержка: 100ms, 200ms, 400ms, ...
		delay := time.Duration(100*(1<<i)) * time.Millisecond
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("все попытки не удались, последняя ошибка: %w", lastErr)
}

// IsTokenValid проверяет валидность токена (упрощенная версия)
func (c *Client) IsTokenValid(token string) (bool, error) {
	response, err := c.ValidateToken(token)
	if err != nil {
		return false, err
	}
	return response.Valid, nil
}

// GetUserInfo возвращает информацию о пользователе по токену
func (c *Client) GetUserInfo(token string) (string, string, error) {
	response, err := c.ValidateToken(token)
	if err != nil {
		return "", "", err
	}

	if !response.Valid {
		return "", "", fmt.Errorf("токен недействителен")
	}

	return response.UserID, response.Role, nil
}

// Ping проверяет доступность auth сервиса
func (c *Client) Ping() error {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/health", nil)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервис недоступен, статус: %d", resp.StatusCode)
	}

	return nil
}

// SetTimeout устанавливает таймаут для HTTP клиента
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetBaseURL изменяет базовый URL клиента
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// Примеры использования в сервисах:
/*
// Локальная валидация JWT
protected := s.echo.Group("/api/v1/profile")
protected.Use(s.jwtMiddleware()) // Использует s/jwt

// Или через auth сервис
authClient := auth.NewClient("http://localhost:8080")
response, err := authClient.ValidateToken(token)
*/

/*
Этот файл содержит HTTP клиент для взаимодействия с auth сервисом:
Основные функции:
ValidateToken - валидация токена через HTTP запрос к auth сервису
ValidateTokenWithRetry - валидация с повторными попытками
IsTokenValid - упрощенная проверка валидности
GetUserInfo - получение информации о пользователе
Ping - проверка доступности auth сервиса
SetTimeout - установка таймаута
SetBaseURL - изменение базового URL
Использование:

// Создание клиента
authClient := auth.NewClient("http://localhost:8080")

// Валидация токена
response, err := authClient.ValidateToken("jwt-token-here")
if err != nil {
    // Обработка ошибки
}

if response.Valid {
    userID := response.UserID
    role := response.Role
    // Использование информации о пользователе
}

// Проверка доступности сервиса
if err := authClient.Ping(); err != nil {
    // Auth сервис недоступен
}
*/

/*
Преимущества этой реализации:

Auth Client:
Надежность - обработка различных HTTP статус кодов
Отказоустойчивость - поддержка повторных попыток
Гибкость - настройка таймаутов и URL
Мониторинг - проверка доступности сервиса
*/

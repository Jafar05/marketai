package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	header = `{"alg":"HS256","typ":"JWT"}`
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"` // Срок действия токена (Unix timestamp)
	Iat    int64  `json:"iat"` // Время выдачи токена (Unix timestamp)
}

func GenerateToken(userID, role, secret string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		Exp:    now.Add(expiration).Unix(),
		Iat:    now.Unix(),
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString([]byte(header))
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", echo.ErrBadRequest.WithInternal(fmt.Errorf("ошибка при маршалинге claims: %w", err))
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerEncoded + "." + claimsEncoded
	signature := signMessage(message, secret)

	return message + "." + signature, nil
}

func ValidateToken(token, secret string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("неверный формат токена")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signature := parts[2]

	message := headerEncoded + "." + claimsEncoded
	expectedSignature := signMessage(message, secret)

	if signature != expectedSignature {
		return nil, errors.New("неверная подпись токена")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, echo.ErrBadRequest.WithInternal(fmt.Errorf("ошибка при декодировании claims: %w", err))
	}
	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, echo.ErrBadRequest.WithInternal(fmt.Errorf("ошибка при демаршалинге claims: %w", err))
	}

	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("токен просрочен")
	}

	return &claims, nil

}

func signMessage(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

package jwt

import (
	"context"
	"net/http"
	"strings"
)

type UserContextKey string

const userContextKey UserContextKey = "user"

// JWTAuthMiddleware создает middleware для валидации JWT токенов
func JWTAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Требуется токен авторизации", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Неверный формат токена авторизации", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			claims, err := ValidateToken(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, "Недействительный токен", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext извлекает информацию о пользователе из контекста
func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(userContextKey).(*Claims)
	return claims, ok
}

// RequireAuth создает middleware, который требует аутентификации
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return JWTAuthMiddleware(jwtSecret)
}

// RequireRole создает middleware, который требует определенной роли
func RequireRole(jwtSecret, requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Сначала валидируем токен
			authMiddleware := JWTAuthMiddleware(jwtSecret)
			authMiddleware(next).ServeHTTP(w, r)

			// Проверяем роль
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Ошибка контекста пользователя", http.StatusInternalServerError)
				return
			}

			if claims.Role != requiredRole {
				http.Error(w, "Недостаточно прав", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole создает middleware, который требует одну из указанных ролей
func RequireAnyRole(jwtSecret string, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Сначала валидируем токен
			authMiddleware := JWTAuthMiddleware(jwtSecret)
			authMiddleware(next).ServeHTTP(w, r)

			// Проверяем роль
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Ошибка контекста пользователя", http.StatusInternalServerError)
				return
			}

			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Недостаточно прав", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth создает middleware, который добавляет информацию о пользователе, если токен есть
func OptionalAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString := parts[1]
					if claims, err := ValidateToken(tokenString, jwtSecret); err == nil {
						ctx := context.WithValue(r.Context(), userContextKey, claims)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Если токена нет или он невалидный, продолжаем без аутентификации
			next.ServeHTTP(w, r)
		})
	}
}

// Примеры использования в сервисах:
/*
// Использование middleware
mux := http.NewServeMux()
mux.Handle("/api/users", jwt.RequireRole(jwtSecret, "admin")(usersHandler))
mux.Handle("/api/profile", jwt.RequireAuth(jwtSecret)(profileHandler))
*/

/*
Описание реализаций
1. s/jwt/middleware.go
Этот файл содержит HTTP middleware для работы с JWT токенами:
Основные функции:
JWTAuthMiddleware - основная функция валидации JWT токенов
GetUserFromContext - извлечение информации о пользователе из контекста
RequireAuth - middleware, требующий аутентификации
RequireRole - middleware, требующий определенной роли
RequireAnyRole - middleware, требующий одну из указанных ролей
OptionalAuth - опциональная аутентификация


// В server.go или http_routes.go
mux := http.NewServeMux()
mux.Handle("/protected", jwt.RequireAuth(jwtSecret)(protectedHandler))
mux.Handle("/admin", jwt.RequireRole(jwtSecret, "admin")(adminHandler))
mux.Handle("/public", jwt.OptionalAuth(jwtSecret)(publicHandler))
*/

/*
Преимущества этой реализации:
JWT Middleware:
Гибкость - разные типы middleware для разных сценариев
Безопасность - правильная обработка ошибок и статус кодов
Производительность - локальная валидация без сетевых вызовов
Удобство - простой API для использования
*/

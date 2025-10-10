# Общий Dockerfile для MarketAI

Этот проект использует один общий Dockerfile для сборки всех Go сервисов.

## Как это работает

Dockerfile использует build arguments для определения какой сервис собирать:

- `SERVICE_NAME` - имя сервиса (auth, cards)
- `BUILD_PATH` - путь к main.go файлу сервиса
- `BINARY_NAME` - имя итогового бинарного файла

## Использование

### Через docker-compose (рекомендуется)

```bash
docker-compose up --build
```

### Ручная сборка

```bash
# Auth сервис
docker build \
  --build-arg SERVICE_NAME=auth \
  --build-arg BUILD_PATH=./auth/cmd/auth \
  --build-arg BINARY_NAME=auth-service \
  -t marketai-auth .

# Cards сервис
docker build \
  --build-arg SERVICE_NAME=cards \
  --build-arg BUILD_PATH=./cards/cmd/cards \
  --build-arg BINARY_NAME=cards-service \
  -t marketai-cards .
```

## Преимущества

1. **Единообразие** - все сервисы собираются одинаково
2. **DRY принцип** - нет дублирования кода
3. **Простота поддержки** - изменения в одном месте
4. **Кэширование** - общие слои кэшируются между сервисами
5. **Безопасность** - единые настройки безопасности

## Структура образа

- **Builder stage**: Go 1.21-alpine с зависимостями
- **Runtime stage**: Alpine Linux с минимальными зависимостями
- **Конфигурация**: Автоматически копируются все configs/ и .env файлы
- **Безопасность**: Запуск от non-root пользователя

# Multi-stage Dockerfile for MarketAI services

# -----------------------
# Builder stage
# -----------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Кэшируем зависимости Go
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Аргументы для выбора сервиса
ARG SERVICE_NAME=auth
ARG BINARY_NAME=service

# Определяем путь к main для выбранного сервиса
ARG BUILD_PATH=./${SERVICE_NAME}/cmd/${SERVICE_NAME}

RUN echo "=== Building MarketAI Service ===" && \
    echo "SERVICE_NAME: ${SERVICE_NAME}" && \
    echo "BINARY_NAME: ${BINARY_NAME}" && \
    echo "BUILD_PATH: ${BUILD_PATH}" && \
    echo "Checking if main.go exists..." && \
    ls -la ${BUILD_PATH} && \
    echo "================================"

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w" \
    -o /out/${BINARY_NAME} \
    ${BUILD_PATH}

# -----------------------
# Runtime stage
# -----------------------
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем бинарник из билд-стадии
COPY --from=builder /out/${BINARY_NAME} ./

# Копируем конфиги
COPY --from=builder /app/configs/ ./configs/

# Переменные окружения
ENV CONFIG_PATH=/app/configs

# Динамический порт в зависимости от сервиса
ARG SERVICE_NAME=auth
ENV SERVICE_NAME=${SERVICE_NAME}

# Устанавливаем порт в зависимости от сервиса
RUN if [ "${SERVICE_NAME}" = "auth" ]; then \
      echo "8080" > /app/port && \
      echo "Auth service will run on port 8080"; \
    else \
      echo "8081" > /app/port && \
      echo "Cards service will run on port 8081"; \
    fi

# Экспонируем оба порта для гибкости
EXPOSE 8080 8081

# CMD запускает выбранный сервис
CMD ["sh", "-c", "./${BINARY_NAME} --config /app/configs/${SERVICE_NAME}/config.yaml --secrets /app/configs/${SERVICE_NAME}/secrets.yaml"]

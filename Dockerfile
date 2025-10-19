# ---------------------------
# Builder stage
# ---------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Установим зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Кэшируем go mod
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Аргументы для выбора сервиса
ARG SERVICE_NAME=auth
# Определяем путь к main.go и имя бинаря динамически
RUN if [ "$SERVICE_NAME" = "auth" ]; then \
      BUILD_PATH=./auth/cmd/auth && BINARY_NAME=auth-service; \
    else \
      BUILD_PATH=./cards/cmd/cards && BINARY_NAME=cards-service; \
    fi && \
    echo "Building $SERVICE_NAME service from $BUILD_PATH -> /out/$BINARY_NAME" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/$BINARY_NAME $BUILD_PATH

# ---------------------------
# Runtime stage
# ---------------------------
FROM alpine:3.19

# Установим runtime зависимости
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем бинарь из билд-стадии
COPY --from=builder /out/ ./

# Копируем конфиги
COPY --from=builder /app/configs/ ./configs/

# Аргументы для выбора сервиса
ARG SERVICE_NAME=auth

# Определяем имя бинаря по SERVICE_NAME
RUN if [ "$SERVICE_NAME" = "auth" ]; then \
      BINARY_NAME=auth-service; \
    else \
      BINARY_NAME=cards-service; \
    fi && \
    echo "Runtime SERVICE_NAME=$SERVICE_NAME, BINARY=$BINARY_NAME" && \
    ln -s /app/$BINARY_NAME /app/service

# Экспонируем порты (оба, для гибкости)
EXPOSE 8080 8081

# CMD запускает сервис через симлинк ./service
CMD ["sh", "-c", "./service --config /app/configs/${SERVICE_NAME}/config.yaml --secrets /app/configs/${SERVICE_NAME}/secrets.yaml"]

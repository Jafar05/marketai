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

RUN echo "Building ${SERVICE_NAME} from ${BUILD_PATH} as ${BINARY_NAME}"

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
COPY --from=builder /out/service ./

# Копируем конфиги
COPY --from=builder /app/configs/ ./configs/

# Переменные окружения
ENV CONFIG_PATH=/app/configs
ENV PORT=8080

EXPOSE 8080

# CMD запускает выбранный сервис
ARG SERVICE_NAME=auth
CMD ["sh", "-c", "./service --config /app/configs/${SERVICE_NAME}/config.yaml --secrets /app/configs/${SERVICE_NAME}/secrets.yaml"]

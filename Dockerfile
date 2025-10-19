FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME=auth
ARG BUILD_PATH=./auth/cmd/auth
ARG BINARY_NAME=auth-service

# Build the service binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w" \
    -o /out/${BINARY_NAME} \
    ${BUILD_PATH}

# ---------------------------
# Runtime stage
# ---------------------------
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/ ./

COPY --from=builder /app/configs/ ./configs/

ENV PORT=8080
ENV CONFIG_PATH=/app/configs
ENV SERVICE_NAME=auth
ENV BINARY_NAME=auth-service

EXPOSE 8080

RUN if [ "${SERVICE_NAME}" = "cards" ]; then \
      ln -s /app/cards-service /app/service; \
    else \
      ln -s /app/auth-service /app/service; \
    fi

CMD ["sh", "-c", "./service --config /app/configs/${SERVICE_NAME}/config.yaml --secrets /app/configs/${SERVICE_NAME}/secrets.yaml"]

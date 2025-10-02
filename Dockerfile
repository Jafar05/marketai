# Multi-stage Dockerfile for Go backend (auth service)

FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build deps
RUN apk add --no-cache git

# Cache go mod deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the sources
COPY . .

# Build the auth service binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/server ./auth/cmd/auth

# Runtime stage
FROM alpine:3.20
WORKDIR /app

ENV PORT=8080
EXPOSE 8080

COPY --from=builder /out/server /app/server

# Copy runtime configs from repo root
COPY ./configs/auth /app/configs/auth

USER nobody

# Rewrite http.port in config to $PORT at container start and run the server
ENTRYPOINT ["sh","-c","sed -i -E 's/^  port: \\".*\\"/  port: ":'\\\"${PORT}\\\"'"/' /app/configs/auth/config.yaml && exec /app/server -config=/app/configs/auth/config.yaml -secrets=/app/configs/auth/secrets.yaml"]





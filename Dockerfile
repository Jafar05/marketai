FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache go mod dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for service selection
ARG SERVICE_NAME=auth
ARG BUILD_PATH=./auth/cmd/auth
ARG BINARY_NAME=auth-service



# Build the service binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w" \
    -o /out/${BINARY_NAME} \
    ${BUILD_PATH}

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /out/ ./

# Copy configuration files
COPY --from=builder /app/configs/ ./configs/

# Note: .env files are not copied as configuration should be provided via environment variables

# Set default environment variables
ENV PORT=8080
ENV CONFIG_PATH=/app/configs
ENV SERVICE_NAME=auth
ENV BINARY_NAME=auth-service

# Expose default port (can be overridden)
EXPOSE 8080

# Default command (can be overridden)
CMD ["./auth-service", "--config", "/app/configs/auth/config.yaml", "--secrets", "/app/configs/auth/secrets.yaml"]
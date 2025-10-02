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

# Build the service binary (select service by build arg)
ARG SERVICE_NAME=auth
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s - w" -o /out/server ./${SERVICE_NAME}/cmd/${SERVICE_NAME}

# Runtime stage
FROM alpine:3.20
WORKDIR /app

RUN adduser -D -H -u 10001 appuser \
    && apk add --no-cache ca-certificates

ENV PORT=8080
EXPOSE 8080

COPY --from=builder /out/server /app/server
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

USER appuser

ENV SERVICE_NAME=${SERVICE_NAME}

ENTRYPOINT ["/app/entrypoint.sh"]
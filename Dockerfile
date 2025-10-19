FROM golang:1.24-alpine AS builder

WORKDIR /app


RUN apk add --no-cache git ca-certificates tzdata


COPY go.mod go.sum ./
RUN go mod download


COPY . .

# Default build args (can be overridden by Railway)
ARG SERVICE_NAME=auth
ARG BUILD_PATH=./${SERVICE_NAME}/cmd/${SERVICE_NAME}
ARG BINARY_NAME=${SERVICE_NAME}-service

RUN echo "Building ${SERVICE_NAME} from ${BUILD_PATH}"


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w" \
    -o /out/${BINARY_NAME} \
    ${BUILD_PATH}


FROM alpine:latest


RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app


COPY --from=builder /out/ ./


COPY --from=builder /app/configs/ ./configs/





ENV CONFIG_PATH=/app/configs
ENV PORT=8080



EXPOSE 8080

# Universal command – uses SERVICE_NAME to find config paths dynamically
ARG SERVICE_NAME=auth
ARG BINARY_NAME=${SERVICE_NAME}-service
CMD ["sh", "-c", "./${BINARY_NAME} --config /app/configs/${SERVICE_NAME}/config.yaml --secrets /app/configs/${SERVICE_NAME}/secrets.yaml"]
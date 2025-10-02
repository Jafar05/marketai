#!/bin/sh
set -eu

# Унифицированная поддержка нескольких сервисов и переменных
SERVICE_NAME="${SERVICE_NAME:-auth}"
CONFIG_DIR="${CONFIG_DIR:-/app/configs/$SERVICE_NAME}"

mkdir -p "$CONFIG_DIR"

# Если переменные заданы — восстанавливаем файлы из Railway
CONFIG_VALUE="${CONFIG_YAML}"
SECRETS_VALUE="${SECRETS_YAML}"

if [ "$CONFIG_VALUE" != "" ]; then
  printf "%s" "$CONFIG_VALUE" > "$CONFIG_DIR/config.yaml"
fi

if [ "$SECRETS_VALUE" != "" ]; then
  printf "%s" "$SECRETS_VALUE" > "$CONFIG_DIR/secrets.yaml"
fi

exec /app/server -config="$CONFIG_DIR/config.yaml" -secrets="$CONFIG_DIR/secrets.yaml"
#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="aurora-hypervisor"
INSTALL_BIN_PATH="/usr/local/bin/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"
ENV_FILE_PATH="${CONFIG_DIR}/.env"
SYSTEMD_SERVICE_DEST="/etc/systemd/system/${SERVICE_NAME}.service"
ENV_FILE_PATH=""
TLS_SRC_DIR=".local/tls"
TLS_DIR="${CONFIG_DIR}/tls"
GO_BIN="${GO_BIN:-go}"

usage() {
  echo "Usage: $0 -e <path-to-env-file>"
  echo "  -e    Path to environment config file (required)"
  exit 1
}

while getopts "e:" opt; do
  case "$opt" in
    e) ENV_FILE_PATH="$OPTARG" ;;
    *) usage ;;
  esac
done

if [ -z "$ENV_FILE_PATH" ]; then
  echo "Error: Environment file is required." >&2
  usage
fi
if [ ! -f "$ENV_FILE_PATH" ]; then
  echo "Error: Environment file not found at $ENV_FILE_PATH" >&2
  exit 1
fi
mkdir -p bin
$GO_BIN build -o "bin/${SERVICE_NAME}" ./cmd/server

if ! getent group aurora >/dev/null 2>&1; then
  sudo groupadd --system aurora
fi
if ! id "$SERVICE_NAME" >/dev/null 2>&1; then
  sudo useradd -r -s /bin/false -g aurora "$SERVICE_NAME"
fi

sudo mkdir -p "$CONFIG_DIR" "$TLS_DIR"

echo "Copying environment file..."
sudo cp "$ENV_FILE_PATH" "${CONFIG_DIR}/.env"

if [ -d "$TLS_SRC_DIR" ]; then
  echo "Copying TLS certificates..."
  sudo rm -rf "${TLS_DIR}/app"
  sudo mkdir -p "${TLS_DIR}/app"
  sudo cp -R "${TLS_SRC_DIR}"/. "${TLS_DIR}/app"/
fi

sudo chown -R "${SERVICE_NAME}:aurora" "$CONFIG_DIR"
sudo chmod 750 "$CONFIG_DIR"
sudo chmod 640 "${CONFIG_DIR}/.env"

sudo systemctl stop "${SERVICE_NAME}.service" >/dev/null 2>&1 || true
sudo install -m 755 "bin/${SERVICE_NAME}" "$INSTALL_BIN_PATH"
sudo cp "package/${SERVICE_NAME}.service" "$SYSTEMD_SERVICE_DEST"
sudo systemctl daemon-reload
sudo systemctl enable "${SERVICE_NAME}.service"
sudo systemctl restart "${SERVICE_NAME}.service"
sudo systemctl status "$SERVICE_NAME" --no-pager -l

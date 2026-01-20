#!/usr/bin/env bash
set -euo pipefail

# Apache Tika Server (full, includes Tesseract OCR) via Podman.
# Commands:
#   ./tika-service.sh start
#   ./tika-service.sh stop
#   ./tika-service.sh status
#   ./tika-service.sh logs

NAME="tika-ocr"
IMAGE="docker.io/apache/tika:latest-full"
PORT_HOST="9998"
PORT_CONT="9998"

start() {
  podman pull "$IMAGE" >/dev/null

  # Remove any old container with same name
  if podman container exists "$NAME"; then
    podman rm -f "$NAME" >/dev/null
  fi

  # Run detached
  podman run -d --name "$NAME" \
    -p "${PORT_HOST}:${PORT_CONT}" \
    --restart=always \
    "$IMAGE" >/dev/null

  # Generate and enable a systemd user service (no root required)
  mkdir -p "${HOME}/.config/systemd/user"
  podman generate systemd --name "$NAME" --files --new >/dev/null
  mv "container-${NAME}.service" "${HOME}/.config/systemd/user/container-${NAME}.service"

  systemctl --user daemon-reload
  systemctl --user enable --now "container-${NAME}.service"

  echo "Started: http://127.0.0.1:${PORT_HOST}/tika"
  echo "Test:   curl -T file.pdf http://127.0.0.1:${PORT_HOST}/tika --header 'Accept: text/plain'"
}

stop() {
  systemctl --user disable --now "container-${NAME}.service" 2>/dev/null || true
  systemctl --user daemon-reload || true

  if podman container exists "$NAME"; then
    podman rm -f "$NAME" >/dev/null
  fi

  rm -f "${HOME}/.config/systemd/user/container-${NAME}.service" || true

  echo "Stopped."
}

status() {
  systemctl --user status "container-${NAME}.service" --no-pager || true
  echo
  podman ps --filter "name=${NAME}" || true
}

logs() {
  podman logs -f "$NAME"
}

case "${1:-}" in
  start) start ;;
  stop) stop ;;
  status) status ;;
  logs) logs ;;
  *)
    echo "Usage: $0 {start|stop|status|logs}" >&2
    exit 1
    ;;
esac

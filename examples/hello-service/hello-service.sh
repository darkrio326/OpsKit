#!/usr/bin/env sh
set -eu

PORT="${SERVICE_PORT:-18080}"

echo "[hello-service] starting on port ${PORT}"

if command -v python3 >/dev/null 2>&1; then
  exec python3 -m http.server "${PORT}" --bind 0.0.0.0
fi

if command -v busybox >/dev/null 2>&1; then
  exec busybox httpd -f -p "${PORT}" -h .
fi

echo "[hello-service] python3/busybox not found, running keepalive loop"
while true; do
  sleep 3600
done

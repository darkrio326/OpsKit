#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${IMAGE:-kylinv10/kylin:b09}"
OUTPUT="${OUTPUT:-${ROOT}/.tmp/generic-e2e-kylin-v10}"
MOUNT_ROOT="${MOUNT_ROOT:-${ROOT}/.tmp/kylin-v10-mounts}"
BIN_DIR="${BIN_DIR:-${ROOT}/.tmp/bin}"
GOCACHE_DIR="${GOCACHE_DIR:-${ROOT}/.tmp/gocache-kylin-v10}"
DOCKER_PLATFORM="${DOCKER_PLATFORM:-}"
DRY_RUN="${DRY_RUN:-0}"

if [[ "${1:-}" == "--help" ]]; then
  cat <<'EOF'
Run OpsKit generic-manage AF verification inside a clean Kylin V10 container.

Environment variables:
  IMAGE            container image (default: kylinv10/kylin:b09)
  OUTPUT           host output dir (default: <repo>/.tmp/generic-e2e-kylin-v10)
  MOUNT_ROOT       host mount root for /opt /data /logs
  BIN_DIR          host dir for built opskit binary
  GOCACHE_DIR      host GOCACHE dir for build
  DOCKER_PLATFORM  optional docker --platform override
  DRY_RUN          pass through dry-run mode to run-af.sh (0 or 1)

Example:
  OUTPUT=$PWD/.tmp/generic-e2e-kylin ./examples/generic-manage/run-af-kylin-v10-docker.sh
EOF
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 2
fi
if ! command -v go >/dev/null 2>&1; then
  echo "go is required to build a Linux opskit binary" >&2
  exit 2
fi

image_arch="$(docker image inspect "${IMAGE}" --format '{{.Architecture}}' 2>/dev/null || true)"
if [[ -z "${image_arch}" ]]; then
  echo "pulling image ${IMAGE} ..."
  docker pull "${IMAGE}" >/dev/null
  image_arch="$(docker image inspect "${IMAGE}" --format '{{.Architecture}}')"
fi

case "${image_arch}" in
  arm64|aarch64) goarch="arm64" ;;
  amd64|x86_64) goarch="amd64" ;;
  *)
    echo "unsupported image architecture: ${image_arch}" >&2
    exit 2
    ;;
esac

mkdir -p "${BIN_DIR}" "${GOCACHE_DIR}" "${OUTPUT}" "${MOUNT_ROOT}/opt" "${MOUNT_ROOT}/data" "${MOUNT_ROOT}/logs"

binary_host="${BIN_DIR}/opskit-linux-${goarch}"
binary_container="/workspace/.tmp/bin/opskit-linux-${goarch}"

echo "== Build opskit for Linux/${goarch} =="
(
  cd "${ROOT}"
  GOCACHE="${GOCACHE_DIR}" GOOS=linux GOARCH="${goarch}" CGO_ENABLED=0 go build -o "${binary_host}" ./cmd/opskit
)
chmod +x "${binary_host}"

echo "== Run OpsKit AF in Kylin V10 container =="
echo "IMAGE=${IMAGE}"
echo "OUTPUT=${OUTPUT}"
echo "DRY_RUN=${DRY_RUN}"

if [[ -n "${DOCKER_PLATFORM}" ]]; then
  docker run --rm \
    --platform "${DOCKER_PLATFORM}" \
    -v "${ROOT}:/workspace" \
    -v "${OUTPUT}:/out" \
    -v "${MOUNT_ROOT}/opt:/opt" \
    -v "${MOUNT_ROOT}/data:/data" \
    -v "${MOUNT_ROOT}/logs:/logs" \
    -w /workspace \
    "${IMAGE}" \
    bash -lc "BIN='${binary_container}' OUTPUT=/out DRY_RUN='${DRY_RUN}' ./examples/generic-manage/run-af.sh"
else
  docker run --rm \
    -v "${ROOT}:/workspace" \
    -v "${OUTPUT}:/out" \
    -v "${MOUNT_ROOT}/opt:/opt" \
    -v "${MOUNT_ROOT}/data:/data" \
    -v "${MOUNT_ROOT}/logs:/logs" \
    -w /workspace \
    "${IMAGE}" \
    bash -lc "BIN='${binary_container}' OUTPUT=/out DRY_RUN='${DRY_RUN}' ./examples/generic-manage/run-af.sh"
fi

echo ""
echo "== Done =="
echo "summary: ${OUTPUT}/summary.json"
echo "ui:      ${OUTPUT}/ui/index.html"

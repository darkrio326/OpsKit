#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Run OpsKit release readiness checks.

Usage:
  scripts/release-check.sh [options]

Options:
  -o, --output <dir>    Output root for dry-run execution (default: ./.tmp/release-check)
      --skip-tests      Skip go test
      --skip-run        Skip run A/D/accept dry-run checks
  -h, --help            Show help

Environment:
  GO_CACHE_DIR          Go build cache dir (default: ./.tmp/gocache-release-check)
USAGE
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/.tmp/release-check"
GO_CACHE_DIR="${GO_CACHE_DIR:-${ROOT_DIR}/.tmp/gocache-release-check}"
SKIP_TESTS=0
SKIP_RUN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --skip-tests)
      SKIP_TESTS=1
      shift
      ;;
    --skip-run)
      SKIP_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

run() {
  echo "==> $*"
  "$@"
}

mkdir -p "${GO_CACHE_DIR}"
mkdir -p "${OUTPUT_DIR}"

cd "${ROOT_DIR}"

if [[ "${SKIP_TESTS}" == "0" ]]; then
  run env GOCACHE="${GO_CACHE_DIR}" go test ./...
fi

run env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit template validate assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json
run env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit template validate assets/templates/demo-hello-service.json --vars-file ./examples/vars/demo-hello-service.env

if [[ "${SKIP_RUN}" == "0" ]]; then
  DEMO_OUT="${OUTPUT_DIR}/demo-audit"
  run env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit run A --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit status --output "${DEMO_OUT}"
fi

echo ""
echo "release-check passed"
echo "output: ${OUTPUT_DIR}"

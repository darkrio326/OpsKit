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
      --with-offline-validate
                         Run offline validation script as an extra gate
      --offline-bin <path>
                         Binary path for offline validation (default: <output>/opskit-release-check)
      --offline-output <dir>
                         Output dir for offline validation (default: <output>/offline-validate)
      --offline-json-status-file <path>
                         status --json output file (default: <offline-output>/status.json)
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
WITH_OFFLINE_VALIDATE=0
OFFLINE_BIN=""
OFFLINE_OUTPUT=""
OFFLINE_JSON_STATUS_FILE=""
STEP_COUNT=0
TOTAL_SECONDS=0
STEP_LINES=()

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
    --with-offline-validate)
      WITH_OFFLINE_VALIDATE=1
      shift
      ;;
    --offline-bin)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OFFLINE_BIN="$2"
      shift 2
      ;;
    --offline-output)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OFFLINE_OUTPUT="$2"
      shift 2
      ;;
    --offline-json-status-file)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OFFLINE_JSON_STATUS_FILE="$2"
      shift 2
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

now_s() {
  date +%s
}

run_step() {
  local label="$1"
  shift
  local started ended elapsed
  started="$(now_s)"
  echo "==> ${label}"
  "$@"
  ended="$(now_s)"
  elapsed=$((ended - started))
  STEP_COUNT=$((STEP_COUNT + 1))
  TOTAL_SECONDS=$((TOTAL_SECONDS + elapsed))
  STEP_LINES+=("${label}|${elapsed}")
  echo "    done: ${elapsed}s"
}

mkdir -p "${GO_CACHE_DIR}"
mkdir -p "${OUTPUT_DIR}"

if [[ -z "${OFFLINE_OUTPUT}" ]]; then
  OFFLINE_OUTPUT="${OUTPUT_DIR}/offline-validate"
fi
if [[ -z "${OFFLINE_BIN}" ]]; then
  OFFLINE_BIN="${OUTPUT_DIR}/opskit-release-check"
fi
if [[ -z "${OFFLINE_JSON_STATUS_FILE}" ]]; then
  OFFLINE_JSON_STATUS_FILE="${OFFLINE_OUTPUT}/status.json"
fi

cd "${ROOT_DIR}"

if [[ "${SKIP_TESTS}" == "0" ]]; then
  run_step "go test ./..." env GOCACHE="${GO_CACHE_DIR}" go test ./...
fi

run_step "template validate demo-server-audit" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit template validate assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json
run_step "template validate demo-hello-service" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit template validate assets/templates/demo-hello-service.json --vars-file ./examples/vars/demo-hello-service.env

if [[ "${SKIP_RUN}" == "0" ]]; then
  DEMO_OUT="${OUTPUT_DIR}/demo-audit"
  run_step "run A dry-run" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit run A --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run_step "run D dry-run" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run_step "accept dry-run" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run_step "status refresh" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit status --output "${DEMO_OUT}"
fi

if [[ "${WITH_OFFLINE_VALIDATE}" == "1" ]]; then
  run_step "build offline validation binary" env GOCACHE="${GO_CACHE_DIR}" go build -o "${OFFLINE_BIN}" ./cmd/opskit
  run_step "offline validation gate" ./scripts/kylin-offline-validate.sh --bin "${OFFLINE_BIN}" --output "${OFFLINE_OUTPUT}" --json-status-file "${OFFLINE_JSON_STATUS_FILE}" --clean
fi

echo ""
echo "release-check summary"
echo "- steps: ${STEP_COUNT}"
for line in "${STEP_LINES[@]}"; do
  label="${line%%|*}"
  elapsed="${line##*|}"
  echo "  - ${label}: ${elapsed}s"
done
echo "- total duration: ${TOTAL_SECONDS}s"
echo "- output: ${OUTPUT_DIR}"
if [[ "${WITH_OFFLINE_VALIDATE}" == "1" ]]; then
  echo "- offline output: ${OFFLINE_OUTPUT}"
  echo "- offline status json: ${OFFLINE_JSON_STATUS_FILE}"
fi
echo "release-check passed"

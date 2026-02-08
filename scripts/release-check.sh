#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Run OpsKit release readiness checks.

Usage:
  scripts/release-check.sh [options]

Options:
  -o, --output <dir>    Output root for checks (default: ./.tmp/release-check)
      --summary-json-file <path>
                        Save machine-readable summary JSON
                        (default: <output>/summary.json)
      --skip-tests      Skip `go test ./...`
      --skip-template-json-contract
                         Skip template validate --json contract gate
      --skip-run        Skip `run A/D/accept/status` dry-run checks
      --with-offline-validate
                         Run offline validation script as an extra gate
      --offline-bin <path>
                         Binary path for offline validation (default: <output>/opskit-release-check)
      --offline-output <dir>
                         Output dir for offline validation (default: <output>/offline-validate)
      --offline-json-status-file <path>
                         status --json output file (default: <offline-output>/status.json)
      --offline-summary-json-file <path>
                         offline summary output file (default: <offline-output>/summary.json)
      --offline-strict-exit
                         Require offline run A/D/accept/status exit code to be 0
  -h, --help            Show help

Environment:
  GO_CACHE_DIR          Go build cache dir (default: ./.tmp/gocache-release-check)
USAGE
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/.tmp/release-check"
GO_CACHE_DIR="${GO_CACHE_DIR:-${ROOT_DIR}/.tmp/gocache-release-check}"
SUMMARY_JSON_FILE=""
SKIP_TESTS=0
SKIP_TEMPLATE_JSON_CONTRACT=0
SKIP_RUN=0
WITH_OFFLINE_VALIDATE=0
OFFLINE_BIN=""
OFFLINE_OUTPUT=""
OFFLINE_JSON_STATUS_FILE=""
OFFLINE_SUMMARY_JSON_FILE=""
OFFLINE_STRICT_EXIT=0
STEP_COUNT=0
TOTAL_SECONDS=0
STEP_LINES=()
START_TS="$(date +%s)"
FINAL_RESULT="pass"
PRIMARY_REASON_CODE="ok"
RECOMMENDED_ACTION="continue_release"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --summary-json-file)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      SUMMARY_JSON_FILE="$2"
      shift 2
      ;;
    --skip-tests)
      SKIP_TESTS=1
      shift
      ;;
    --skip-template-json-contract)
      SKIP_TEMPLATE_JSON_CONTRACT=1
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
    --offline-summary-json-file)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OFFLINE_SUMMARY_JSON_FILE="$2"
      shift 2
      ;;
    --offline-strict-exit)
      OFFLINE_STRICT_EXIT=1
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

now_s() {
  date +%s
}

now_iso() {
  date -u +%Y-%m-%dT%H:%M:%SZ
}

json_escape() {
  local s="$1"
  s=${s//\\/\\\\}
  s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}
  s=${s//$'\r'/\\r}
  s=${s//$'\t'/\\t}
  printf '%s' "${s}"
}

bool_json() {
  if [[ "$1" == "1" ]]; then
    printf 'true'
  else
    printf 'false'
  fi
}

write_summary_json() {
  local tmp_file end_ts duration i
  end_ts="$(now_s)"
  duration=$((end_ts - START_TS))
  tmp_file="${SUMMARY_JSON_FILE}.tmp"

  {
    echo "{"
    echo "  \"schemaVersion\": \"v1\","
    echo "  \"generatedAt\": \"$(now_iso)\","
    echo "  \"result\": \"${FINAL_RESULT}\","
    echo "  \"reasonCode\": \"${PRIMARY_REASON_CODE}\","
    echo "  \"recommendedAction\": \"${RECOMMENDED_ACTION}\","
    echo "  \"output\": \"$(json_escape "${OUTPUT_DIR}")\","
    echo "  \"goCacheDir\": \"$(json_escape "${GO_CACHE_DIR}")\","
    echo "  \"skipTests\": $(bool_json "${SKIP_TESTS}"),"
    echo "  \"skipTemplateJsonContract\": $(bool_json "${SKIP_TEMPLATE_JSON_CONTRACT}"),"
    echo "  \"skipRun\": $(bool_json "${SKIP_RUN}"),"
    echo "  \"withOfflineValidate\": $(bool_json "${WITH_OFFLINE_VALIDATE}"),"
    echo "  \"offlineStrictExit\": $(bool_json "${OFFLINE_STRICT_EXIT}"),"
    echo "  \"steps\": ${STEP_COUNT},"
    echo "  \"totalDurationSeconds\": ${duration},"
    echo "  \"stepResults\": ["

    for i in "${!STEP_LINES[@]}"; do
      local line label elapsed rc reason comma
      IFS='|' read -r label elapsed rc reason <<< "${STEP_LINES[$i]}"
      comma=","
      if [[ "${i}" -eq $((${#STEP_LINES[@]} - 1)) ]]; then
        comma=""
      fi
      echo "    {\"name\":\"$(json_escape "${label}")\",\"elapsedSeconds\":${elapsed},\"exitCode\":${rc},\"reasonCode\":\"$(json_escape "${reason}")\"}${comma}"
    done

    echo "  ],"
    echo "  \"offline\": {"
    echo "    \"output\": \"$(json_escape "${OFFLINE_OUTPUT}")\","
    echo "    \"jsonStatusFile\": \"$(json_escape "${OFFLINE_JSON_STATUS_FILE}")\","
    echo "    \"summaryJsonFile\": \"$(json_escape "${OFFLINE_SUMMARY_JSON_FILE}")\""
    echo "  }"
    echo "}"
  } > "${tmp_file}"

  mv "${tmp_file}" "${SUMMARY_JSON_FILE}"
}

run_step() {
  local label="$1"
  local reason="$2"
  shift
  shift
  local rc started ended elapsed effective_reason
  started="$(now_s)"
  echo "==> ${label}"
  set +e
  "$@"
  rc=$?
  set -e
  ended="$(now_s)"
  elapsed=$((ended - started))
  STEP_COUNT=$((STEP_COUNT + 1))
  TOTAL_SECONDS=$((TOTAL_SECONDS + elapsed))
  effective_reason="ok"
  if [[ "${rc}" != "0" ]]; then
    effective_reason="${reason}"
  fi
  STEP_LINES+=("${label}|${elapsed}|${rc}|${effective_reason}")
  if [[ "${rc}" == "0" ]]; then
    echo "    done: ${elapsed}s"
    return 0
  fi

  FINAL_RESULT="fail"
  PRIMARY_REASON_CODE="${reason}"
  RECOMMENDED_ACTION="block_release"
  write_summary_json

  echo "    failed: ${elapsed}s rc=${rc}" >&2
  echo ""
  echo "release-check failed"
  echo "- reason: ${PRIMARY_REASON_CODE}"
  echo "- recommended action: ${RECOMMENDED_ACTION}"
  echo "- summary json: ${SUMMARY_JSON_FILE}"
  exit "${rc}"
}

mkdir -p "${GO_CACHE_DIR}"
mkdir -p "${OUTPUT_DIR}"

if [[ -z "${SUMMARY_JSON_FILE}" ]]; then
  SUMMARY_JSON_FILE="${OUTPUT_DIR}/summary.json"
fi

if [[ -z "${OFFLINE_OUTPUT}" ]]; then
  OFFLINE_OUTPUT="${OUTPUT_DIR}/offline-validate"
fi
if [[ -z "${OFFLINE_BIN}" ]]; then
  OFFLINE_BIN="${OUTPUT_DIR}/opskit-release-check"
fi
if [[ -z "${OFFLINE_JSON_STATUS_FILE}" ]]; then
  OFFLINE_JSON_STATUS_FILE="${OFFLINE_OUTPUT}/status.json"
fi
if [[ -z "${OFFLINE_SUMMARY_JSON_FILE}" ]]; then
  OFFLINE_SUMMARY_JSON_FILE="${OFFLINE_OUTPUT}/summary.json"
fi
mkdir -p "$(dirname "${SUMMARY_JSON_FILE}")"

cd "${ROOT_DIR}"

if [[ "${SKIP_TESTS}" == "0" ]]; then
  run_step "go test ./..." "step_failed_go_test" env GOCACHE="${GO_CACHE_DIR}" go test ./...
fi

run_step "template validate demo-server-audit" "step_failed_template_validate_demo_server_audit" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit template validate assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json
run_step "template validate demo-hello-service" "step_failed_template_validate_demo_hello_service" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit template validate assets/templates/demo-hello-service.json --vars-file ./examples/vars/demo-hello-service.env

if [[ "${SKIP_TEMPLATE_JSON_CONTRACT}" == "0" ]]; then
  run_step "template validate json contract" "step_failed_template_validate_json_contract" env GO_CACHE_DIR="${GO_CACHE_DIR}" ./scripts/template-validate-check.sh --output "${OUTPUT_DIR}/template-validate-check" --clean
fi

if [[ "${SKIP_RUN}" == "0" ]]; then
  DEMO_OUT="${OUTPUT_DIR}/demo-audit"
  run_step "run A dry-run" "step_failed_run_a_dry_run" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit run A --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run_step "run D dry-run" "step_failed_run_d_dry_run" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run_step "accept dry-run" "step_failed_accept_dry_run" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json --output "${DEMO_OUT}" --dry-run
  run_step "status refresh" "step_failed_status_refresh" env GOCACHE="${GO_CACHE_DIR}" go run ./cmd/opskit status --output "${DEMO_OUT}"
fi

if [[ "${WITH_OFFLINE_VALIDATE}" == "1" ]]; then
  offline_args=(
    --bin "${OFFLINE_BIN}"
    --output "${OFFLINE_OUTPUT}"
    --json-status-file "${OFFLINE_JSON_STATUS_FILE}"
    --summary-json-file "${OFFLINE_SUMMARY_JSON_FILE}"
    --clean
  )
  if [[ "${OFFLINE_STRICT_EXIT}" == "1" ]]; then
    offline_args+=(--strict-exit)
  fi
  run_step "build offline validation binary" "step_failed_build_offline_binary" env GOCACHE="${GO_CACHE_DIR}" go build -o "${OFFLINE_BIN}" ./cmd/opskit
  run_step "offline validation gate" "step_failed_offline_validation_gate" ./scripts/kylin-offline-validate.sh "${offline_args[@]}"
fi

write_summary_json

echo ""
echo "release-check summary"
echo "- steps: ${STEP_COUNT}"
for line in "${STEP_LINES[@]}"; do
  IFS='|' read -r label elapsed rc reason <<< "${line}"
  echo "  - ${label}: ${elapsed}s (rc=${rc}, reason=${reason})"
done
echo "- total duration: ${TOTAL_SECONDS}s"
echo "- output: ${OUTPUT_DIR}"
echo "- reason: ${PRIMARY_REASON_CODE}"
echo "- recommended action: ${RECOMMENDED_ACTION}"
echo "- summary json: ${SUMMARY_JSON_FILE}"
if [[ "${WITH_OFFLINE_VALIDATE}" == "1" ]]; then
  echo "- offline output: ${OFFLINE_OUTPUT}"
  echo "- offline status json: ${OFFLINE_JSON_STATUS_FILE}"
  echo "- offline summary json: ${OFFLINE_SUMMARY_JSON_FILE}"
  echo "- offline strict exit: ${OFFLINE_STRICT_EXIT}"
fi
echo "release-check passed"

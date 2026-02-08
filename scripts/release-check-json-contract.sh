#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Validate machine-readable JSON contract of scripts/release-check.sh.

Usage:
  scripts/release-check-json-contract.sh [options]

Options:
  -o, --output <dir>      Output directory (default: ./.tmp/release-check-json-contract)
      --clean             Remove output directory before running
  -h, --help              Show help

Environment:
  GO_CACHE_DIR            Go cache dir (default: ./.tmp/gocache-release-check-json-contract)

Examples:
  scripts/release-check-json-contract.sh --clean
USAGE
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/.tmp/release-check-json-contract"
CLEAN=0
GO_CACHE_DIR="${GO_CACHE_DIR:-${ROOT_DIR}/.tmp/gocache-release-check-json-contract}"

FAIL_COUNT=0
FAIL_REASONS=()
RESULT="pass"
REASON_CODE="ok"
RECOMMENDED_ACTION="continue_ci"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --clean)
      CLEAN=1
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

if [[ "${CLEAN}" == "1" ]]; then
  rm -rf "${OUTPUT_DIR}"
fi
mkdir -p "${OUTPUT_DIR}" "${GO_CACHE_DIR}"

add_failure() {
  local reason="$1"
  FAIL_REASONS+=("${reason}")
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

require_pattern() {
  local file="$1"
  local pattern="$2"
  local reason="$3"
  if grep -Eq -- "${pattern}" "${file}"; then
    return 0
  fi
  add_failure "${reason}"
  return 0
}

PASS_SUMMARY="${OUTPUT_DIR}/release-check-pass-summary.json"
FAIL_SUMMARY="${OUTPUT_DIR}/release-check-fail-summary.json"
SUMMARY_JSON="${OUTPUT_DIR}/summary.json"

echo "==> pass scenario"
env GO_CACHE_DIR="${GO_CACHE_DIR}" "${ROOT_DIR}/scripts/release-check.sh" \
  --skip-tests \
  --skip-run \
  --summary-json-file "${PASS_SUMMARY}"

require_pattern "${PASS_SUMMARY}" '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "pass:missing_schema_version"
require_pattern "${PASS_SUMMARY}" '"result"[[:space:]]*:[[:space:]]*"pass"' "pass:missing_result_pass"
require_pattern "${PASS_SUMMARY}" '"reasonCode"[[:space:]]*:[[:space:]]*"ok"' "pass:missing_reason_ok"
require_pattern "${PASS_SUMMARY}" '"recommendedAction"[[:space:]]*:[[:space:]]*"continue_release"' "pass:missing_continue_release"
require_pattern "${PASS_SUMMARY}" '"stepResults"[[:space:]]*:[[:space:]]*\[' "pass:missing_step_results"
require_pattern "${PASS_SUMMARY}" '"reasonCode"[[:space:]]*:[[:space:]]*"ok"' "pass:missing_step_reason_ok"

echo "==> fail scenario (negative)"
set +e
env GO_CACHE_DIR="${GO_CACHE_DIR}" "${ROOT_DIR}/scripts/release-check.sh" \
  --skip-tests \
  --skip-run \
  --with-offline-validate \
  --offline-bin /dev/null/opskit \
  --summary-json-file "${FAIL_SUMMARY}"
FAIL_RC=$?
set -e

if [[ "${FAIL_RC}" == "0" ]]; then
  add_failure "fail:expected_non_zero_exit"
fi

require_pattern "${FAIL_SUMMARY}" '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "fail:missing_schema_version"
require_pattern "${FAIL_SUMMARY}" '"result"[[:space:]]*:[[:space:]]*"fail"' "fail:missing_result_fail"
require_pattern "${FAIL_SUMMARY}" '"reasonCode"[[:space:]]*:[[:space:]]*"step_failed_build_offline_binary"' "fail:missing_reason_code"
require_pattern "${FAIL_SUMMARY}" '"recommendedAction"[[:space:]]*:[[:space:]]*"block_release"' "fail:missing_block_release"
require_pattern "${FAIL_SUMMARY}" '"exitCode"[[:space:]]*:[[:space:]]*[1-9][0-9]*' "fail:missing_non_zero_step_exit_code"
require_pattern "${FAIL_SUMMARY}" '"reasonCode"[[:space:]]*:[[:space:]]*"step_failed_build_offline_binary"' "fail:missing_step_reason_code"

if [[ "${FAIL_COUNT}" -gt 0 ]]; then
  RESULT="fail"
  REASON_CODE="release_check_json_contract_failed"
  RECOMMENDED_ACTION="block_ci"
fi

{
  echo "{"
  echo "  \"schemaVersion\": \"v1\","
  echo "  \"result\": \"${RESULT}\","
  echo "  \"reasonCode\": \"${REASON_CODE}\","
  echo "  \"recommendedAction\": \"${RECOMMENDED_ACTION}\","
  echo "  \"failCount\": ${FAIL_COUNT},"
  echo "  \"outputs\": ["
  echo "    \"${PASS_SUMMARY}\","
  echo "    \"${FAIL_SUMMARY}\""
  echo "  ],"
  echo "  \"failures\": ["
  for i in "${!FAIL_REASONS[@]}"; do
    comma=","
    if [[ "${i}" -eq $((${#FAIL_REASONS[@]} - 1)) ]]; then
      comma=""
    fi
    echo "    \"${FAIL_REASONS[$i]}\"${comma}"
  done
  echo "  ]"
  echo "}"
} > "${SUMMARY_JSON}.tmp"
mv "${SUMMARY_JSON}.tmp" "${SUMMARY_JSON}"

echo ""
echo "release-check-json-contract summary"
echo "- result: ${RESULT}"
echo "- reason: ${REASON_CODE}"
echo "- recommended action: ${RECOMMENDED_ACTION}"
echo "- summary: ${SUMMARY_JSON}"

if [[ "${RESULT}" != "pass" ]]; then
  exit 1
fi

#!/usr/bin/env bash
set -uo pipefail

usage() {
  cat <<'USAGE'
Validate OpsKit basic data generation on an offline Kylin V10 host.

Usage:
  scripts/kylin-offline-validate.sh [options]

Options:
  -b, --bin <path>        OpsKit binary path or command name (default: opskit)
  -o, --output <dir>      Output directory (default: /data/opskit-regression-v034)
  -t, --template <id>     Template id/path (default: generic-manage-v1)
      --json-status-file <path>
                           Save `opskit status --json` output to this file
                           (default: <output>/status.json)
      --summary-json-file <path>
                           Save offline validation summary JSON to this file
                           (default: <output>/summary.json)
      --strict-exit       Require run A/D/accept/status exit code to be 0
      --clean             Remove output directory before running
  -h, --help              Show help

Notes:
  - This script runs real commands (not dry-run): run A, run D, accept, status.
  - Default expected stage/status exit codes are 0/1/3 (pass/fail/warn).
  - Use --strict-exit to enforce exit code 0 for all stage/status commands.
USAGE
}

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

BIN="${BIN:-opskit}"
OUTPUT="${OUTPUT:-/data/opskit-regression-v034}"
TEMPLATE="${TEMPLATE:-generic-manage-v1}"
JSON_STATUS_FILE=""
SUMMARY_JSON_FILE=""
STRICT_EXIT=0
CLEAN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    -b|--bin)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      BIN="$2"
      shift 2
      ;;
    -o|--output)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OUTPUT="$2"
      shift 2
      ;;
    -t|--template)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      TEMPLATE="$2"
      shift 2
      ;;
    --json-status-file)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      JSON_STATUS_FILE="$2"
      shift 2
      ;;
    --summary-json-file)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      SUMMARY_JSON_FILE="$2"
      shift 2
      ;;
    --strict-exit)
      STRICT_EXIT=1
      shift
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

if ! command -v "${BIN}" >/dev/null 2>&1; then
  echo "opskit binary not found: ${BIN}" >&2
  exit 2
fi

if [[ "${CLEAN}" == "1" ]]; then
  rm -rf "${OUTPUT}"
fi
mkdir -p "${OUTPUT}"

if [[ -z "${JSON_STATUS_FILE}" ]]; then
  JSON_STATUS_FILE="${OUTPUT}/status.json"
fi
if [[ -z "${SUMMARY_JSON_FILE}" ]]; then
  SUMMARY_JSON_FILE="${OUTPUT}/summary.json"
fi
mkdir -p "$(dirname "${JSON_STATUS_FILE}")"
mkdir -p "$(dirname "${SUMMARY_JSON_FILE}")"

RESULT_LINES=()
FAIL_MESSAGES=()
FAIL_COUNT=0
HARD_FAIL=0
PRIMARY_REASON_CODE="ok"
START_TS="$(now_s)"
LATEST_ACCEPT_REPORT=""

add_failure() {
  local reason="$1"
  local message="$2"
  FAIL_MESSAGES+=("${reason}|${message}")
  if [[ "${PRIMARY_REASON_CODE}" == "ok" ]]; then
    PRIMARY_REASON_CODE="${reason}"
  fi
}

mark_result() {
  local kind="$1"
  local name="$2"
  local rc="$3"
  local elapsed="$4"
  RESULT_LINES+=("${kind}|${name}|${rc}|${elapsed}")
}

is_allowed_stage_rc() {
  local rc="$1"
  if [[ "${STRICT_EXIT}" == "1" ]]; then
    [[ "${rc}" == "0" ]]
    return
  fi
  [[ "${rc}" == "0" || "${rc}" == "1" || "${rc}" == "3" ]]
}

run_expect_zero() {
  local name="$1"
  shift
  local started ended elapsed rc
  started="$(now_s)"
  echo "==> ${name}"
  set +e
  "${BIN}" "$@"
  rc=$?
  set -e
  ended="$(now_s)"
  elapsed=$((ended - started))
  mark_result "zero" "${name}" "${rc}" "${elapsed}"
  if [[ "${rc}" != "0" ]]; then
    echo "    failed: expected rc=0, got ${rc}" >&2
    add_failure "unexpected_exit_code" "${name} expected rc=0 actual=${rc}"
    HARD_FAIL=1
  else
    echo "    ok: rc=${rc}"
  fi
}

run_expect_stage_rc() {
  local name="$1"
  shift
  local started ended elapsed rc
  started="$(now_s)"
  echo "==> ${name}"
  set +e
  "${BIN}" "$@"
  rc=$?
  set -e
  ended="$(now_s)"
  elapsed=$((ended - started))
  mark_result "stage" "${name}" "${rc}" "${elapsed}"
  if is_allowed_stage_rc "${rc}"; then
    echo "    ok: rc=${rc}"
  else
    if [[ "${STRICT_EXIT}" == "1" ]]; then
      echo "    failed: unexpected rc=${rc} (expect 0 in strict mode)" >&2
    else
      echo "    failed: unexpected rc=${rc} (expect 0/1/3)" >&2
    fi
    add_failure "unexpected_exit_code" "${name} unexpected rc=${rc}"
    HARD_FAIL=1
  fi
}

run_status_json_expect_stage_rc() {
  local name="$1"
  shift
  local json_file="$1"
  shift
  local started ended elapsed rc
  started="$(now_s)"
  echo "==> ${name}"
  set +e
  "${BIN}" "$@" --json > "${json_file}"
  rc=$?
  set -e
  ended="$(now_s)"
  elapsed=$((ended - started))
  mark_result "status" "${name}" "${rc}" "${elapsed}"

  if is_allowed_stage_rc "${rc}"; then
    echo "    ok: rc=${rc}"
  else
    if [[ "${STRICT_EXIT}" == "1" ]]; then
      echo "    failed: unexpected rc=${rc} (expect 0 in strict mode)" >&2
    else
      echo "    failed: unexpected rc=${rc} (expect 0/1/3)" >&2
    fi
    add_failure "unexpected_exit_code" "${name} unexpected rc=${rc}"
    HARD_FAIL=1
  fi

  if [[ ! -s "${json_file}" ]]; then
    echo "    failed: status json is empty (${json_file})" >&2
    add_failure "status_json_empty" "status json is empty: ${json_file}"
    HARD_FAIL=1
    return
  fi
  if ! grep -q '"command"[[:space:]]*:[[:space:]]*"opskit status"' "${json_file}"; then
    echo "    failed: status json missing command field (${json_file})" >&2
    add_failure "status_json_invalid" "status json missing command field"
    HARD_FAIL=1
  fi
  if ! grep -q '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "${json_file}"; then
    echo "    failed: status json missing schemaVersion=v1 (${json_file})" >&2
    add_failure "status_json_invalid" "status json missing schemaVersion=v1"
    HARD_FAIL=1
  fi
  if ! grep -q "\"exitCode\"[[:space:]]*:[[:space:]]*${rc}" "${json_file}"; then
    echo "    failed: status json exitCode mismatch (${json_file})" >&2
    add_failure "status_json_invalid" "status json exitCode mismatch"
    HARD_FAIL=1
  fi
  echo "    json: ${json_file}"
}

require_file() {
  local path="$1"
  if [[ -f "${path}" ]]; then
    echo "    ok: ${path}"
    return
  fi
  echo "    missing: ${path}" >&2
  add_failure "required_file_missing" "missing file: ${path}"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

require_grep() {
  local pattern="$1"
  local path="$2"
  if grep -q -- "${pattern}" "${path}"; then
    echo "    ok: ${path} contains '${pattern}'"
    return
  fi
  echo "    check failed: ${path} missing '${pattern}'" >&2
  add_failure "required_content_missing" "${path} missing pattern: ${pattern}"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

write_summary_json() {
  local end_ts duration result reason tmp_file
  end_ts="$(now_s)"
  duration=$((end_ts - START_TS))
  result="pass"
  reason="ok"

  if [[ "${HARD_FAIL}" == "1" || "${FAIL_COUNT}" -gt 0 ]]; then
    result="fail"
    reason="${PRIMARY_REASON_CODE}"
    if [[ -z "${reason}" || "${reason}" == "ok" ]]; then
      reason="validation_failed"
    fi
  fi

  tmp_file="${SUMMARY_JSON_FILE}.tmp"

  {
    echo "{"
    echo "  \"schemaVersion\": \"v1\"," 
    echo "  \"command\": \"$(json_escape "scripts/kylin-offline-validate.sh")\"," 
    echo "  \"generatedAt\": \"$(now_iso)\"," 
    echo "  \"durationSeconds\": ${duration},"
    echo "  \"result\": \"${result}\"," 
    echo "  \"reasonCode\": \"$(json_escape "${reason}")\"," 
    echo "  \"strictExit\": $(bool_json "${STRICT_EXIT}"),"
    echo "  \"hardFail\": $(bool_json "${HARD_FAIL}"),"
    echo "  \"verifyFailCount\": ${FAIL_COUNT},"
    echo "  \"template\": \"$(json_escape "${TEMPLATE}")\"," 
    echo "  \"output\": \"$(json_escape "${OUTPUT}")\"," 
    echo "  \"statusJsonFile\": \"$(json_escape "${JSON_STATUS_FILE}")\"," 
    if [[ -n "${LATEST_ACCEPT_REPORT}" ]]; then
      echo "  \"latestAcceptReport\": \"$(json_escape "${LATEST_ACCEPT_REPORT}")\"," 
    else
      echo "  \"latestAcceptReport\": null," 
    fi
    echo "  \"stageResults\": ["

    local idx=0
    local total="${#RESULT_LINES[@]}"
    local line kind name rc elapsed allowed expected
    for line in "${RESULT_LINES[@]-}"; do
      IFS='|' read -r kind name rc elapsed <<< "${line}"
      allowed="false"
      expected="0"
      case "${kind}" in
        zero)
          if [[ "${rc}" == "0" ]]; then
            allowed="true"
          fi
          expected="0"
          ;;
        stage|status)
          if is_allowed_stage_rc "${rc}"; then
            allowed="true"
          fi
          if [[ "${STRICT_EXIT}" == "1" ]]; then
            expected="0"
          else
            expected="0/1/3"
          fi
          ;;
      esac

      idx=$((idx + 1))
      echo "    {"
      echo "      \"kind\": \"$(json_escape "${kind}")\"," 
      echo "      \"name\": \"$(json_escape "${name}")\"," 
      echo "      \"exitCode\": ${rc},"
      echo "      \"elapsedSeconds\": ${elapsed},"
      echo "      \"expectedExitCodes\": \"${expected}\"," 
      echo "      \"allowed\": ${allowed}"
      if [[ "${idx}" -lt "${total}" ]]; then
        echo "    },"
      else
        echo "    }"
      fi
    done

    echo "  ],"
    echo "  \"failures\": ["

    idx=0
    total="${#FAIL_MESSAGES[@]}"
    local reason_code message
    for line in "${FAIL_MESSAGES[@]-}"; do
      if [[ -z "${line}" ]]; then
        continue
      fi
      IFS='|' read -r reason_code message <<< "${line}"
      idx=$((idx + 1))
      echo "    {"
      echo "      \"reasonCode\": \"$(json_escape "${reason_code}")\"," 
      echo "      \"message\": \"$(json_escape "${message}")\""
      if [[ "${idx}" -lt "${total}" ]]; then
        echo "    },"
      else
        echo "    }"
      fi
    done

    echo "  ]"
    echo "}"
  } > "${tmp_file}"

  mv "${tmp_file}" "${SUMMARY_JSON_FILE}"
}

set -e
run_expect_zero "template validate (${TEMPLATE})" template validate "${TEMPLATE}" --output "${OUTPUT}"
run_expect_stage_rc "run A" run A --template "${TEMPLATE}" --output "${OUTPUT}"
run_expect_stage_rc "run D" run D --template "${TEMPLATE}" --output "${OUTPUT}"
run_expect_stage_rc "accept" accept --template "${TEMPLATE}" --output "${OUTPUT}"
run_status_json_expect_stage_rc "status" "${JSON_STATUS_FILE}" status --output "${OUTPUT}"

echo "==> verify outputs"
require_file "${OUTPUT}/state/overall.json"
require_file "${OUTPUT}/state/lifecycle.json"
require_file "${OUTPUT}/state/services.json"
require_file "${OUTPUT}/state/artifacts.json"
require_file "${JSON_STATUS_FILE}"

require_grep '"summary"' "${OUTPUT}/state/lifecycle.json"
require_grep 'acceptance-consistency-' "${OUTPUT}/state/artifacts.json"
require_grep '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "${JSON_STATUS_FILE}"
require_grep '"command"[[:space:]]*:[[:space:]]*"opskit status"' "${JSON_STATUS_FILE}"

latest_accept="$(ls -1t "${OUTPUT}"/reports/accept-*.html 2>/dev/null | head -n1 || true)"
if [[ -z "${latest_accept}" ]]; then
  echo "    missing: ${OUTPUT}/reports/accept-*.html" >&2
  add_failure "accept_report_missing" "missing file: ${OUTPUT}/reports/accept-*.html"
  FAIL_COUNT=$((FAIL_COUNT + 1))
else
  LATEST_ACCEPT_REPORT="${latest_accept}"
  echo "    ok: ${latest_accept}"
  require_grep '"consistency"' "${latest_accept}"
fi

echo ""
echo "offline validation summary"
for line in "${RESULT_LINES[@]}"; do
  IFS='|' read -r _kind name rc elapsed <<< "${line}"
  echo "- ${name}: rc=${rc}, elapsed=${elapsed}s"
done
echo "- output: ${OUTPUT}"
echo "- status json: ${JSON_STATUS_FILE}"
echo "- summary json: ${SUMMARY_JSON_FILE}"
echo "- strict exit: ${STRICT_EXIT}"

write_summary_json

if [[ "${HARD_FAIL}" == "1" || "${FAIL_COUNT}" -gt 0 ]]; then
  echo "offline validation failed (reason=${PRIMARY_REASON_CODE}, hard_fail=${HARD_FAIL}, verify_fail=${FAIL_COUNT})" >&2
  exit 1
fi

echo "offline validation passed"

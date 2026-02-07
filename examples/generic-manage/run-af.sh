#!/usr/bin/env bash
set -euo pipefail

BIN="${BIN:-opskit}"
OUTPUT="${OUTPUT:-/tmp/opskit-generic}"
DRY_RUN="${DRY_RUN:-0}"
STRICT_EXIT="${STRICT_EXIT:-0}"
FAIL_ON_UNEXPECTED="${FAIL_ON_UNEXPECTED:-1}"
JSON_STATUS_FILE="${JSON_STATUS_FILE:-${OUTPUT}/status.json}"
read -r -a BIN_CMD <<< "${BIN}"
STEP_NAMES=()
STEP_CODES=()
STEP_EXPECTED=()
STEP_ALLOWED=()
VERIFY_FAILURES=()
VERIFY_FAIL_COUNT=0

if [[ "${1:-}" == "--help" ]]; then
  cat <<'EOF'
Run generic-manage end-to-end checks (A~F) and print verification hints.

Environment variables:
  BIN                 opskit binary path or command name (default: opskit)
  OUTPUT              output root directory (default: /tmp/opskit-generic)
  DRY_RUN             set to 1 to execute dry-run flow
  STRICT_EXIT         set to 1 to require stage/status exit code = 0
  FAIL_ON_UNEXPECTED  set to 1 to return non-zero on unexpected failure (default: 1)
  JSON_STATUS_FILE    status --json output path (default: <output>/status.json)

Examples:
  BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
  DRY_RUN=1 STRICT_EXIT=1 ./examples/generic-manage/run-af.sh
EOF
  exit 0
fi

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

expected_stage_codes() {
  if [[ "${STRICT_EXIT}" == "1" ]]; then
    echo "0"
  else
    echo "0|1|3"
  fi
}

is_exit_allowed() {
  local code="$1"
  local expected="$2"
  IFS='|' read -r -a allowed_codes <<< "${expected}"
  for c in "${allowed_codes[@]}"; do
    if [[ "${code}" == "${c}" ]]; then
      return 0
    fi
  done
  return 1
}

record_step() {
  local name="$1"
  local code="$2"
  local expected="$3"
  local allowed="false"
  if is_exit_allowed "${code}" "${expected}"; then
    allowed="true"
  fi
  STEP_NAMES+=("${name}")
  STEP_CODES+=("${code}")
  STEP_EXPECTED+=("${expected}")
  STEP_ALLOWED+=("${allowed}")
}

run_and_capture() {
  local name="$1"
  local expected="$2"
  shift 2
  echo ""
  echo "[${name}] $* (expected=${expected})"
  set +e
  "$@"
  local code=$?
  set -e
  record_step "${name}" "${code}" "${expected}"
  if [[ "${STEP_ALLOWED[${#STEP_ALLOWED[@]}-1]}" == "true" ]]; then
    echo "[${name}] exit=${code} (allowed)"
  else
    echo "[${name}] exit=${code} (unexpected)" >&2
  fi
  return 0
}

run_status_json_capture() {
  local name="$1"
  local expected="$2"
  shift 2
  mkdir -p "$(dirname "${JSON_STATUS_FILE}")"
  echo ""
  echo "[${name}] $* --json > ${JSON_STATUS_FILE} (expected=${expected})"
  set +e
  "$@" --json > "${JSON_STATUS_FILE}"
  local code=$?
  set -e
  record_step "${name}" "${code}" "${expected}"
  if [[ "${STEP_ALLOWED[${#STEP_ALLOWED[@]}-1]}" == "true" ]]; then
    echo "[${name}] exit=${code} (allowed)"
  else
    echo "[${name}] exit=${code} (unexpected)" >&2
  fi
}

add_verify_failure() {
  local message="$1"
  VERIFY_FAILURES+=("${message}")
  VERIFY_FAIL_COUNT=$((VERIFY_FAIL_COUNT + 1))
}

require_file() {
  local path="$1"
  if [[ -f "${path}" ]]; then
    echo "verify ok: ${path}"
    return 0
  fi
  echo "verify missing: ${path}" >&2
  add_verify_failure "missing_file:${path}"
  return 0
}

require_grep() {
  local pattern="$1"
  local path="$2"
  if grep -Eq -- "${pattern}" "${path}"; then
    echo "verify ok: ${path} matches ${pattern}"
    return 0
  fi
  echo "verify failed: ${path} missing pattern ${pattern}" >&2
  add_verify_failure "missing_pattern:${path}:${pattern}"
  return 0
}

OVERALL_RESULT="pass"
REASON_CODE="ok"
RECOMMENDED_ACTION="continue_validation"
UNEXPECTED_STEPS=0
WARN_STEPS=0

compute_outcome() {
  local total="${#STEP_NAMES[@]}"
  local i
  UNEXPECTED_STEPS=0
  WARN_STEPS=0
  for ((i = 0; i < total; i++)); do
    if [[ "${STEP_ALLOWED[$i]}" != "true" ]]; then
      UNEXPECTED_STEPS=$((UNEXPECTED_STEPS + 1))
      continue
    fi
    if [[ "${STEP_CODES[$i]}" != "0" ]]; then
      WARN_STEPS=$((WARN_STEPS + 1))
    fi
  done

  OVERALL_RESULT="pass"
  REASON_CODE="ok"
  RECOMMENDED_ACTION="continue_validation"
  if [[ "${UNEXPECTED_STEPS}" -gt 0 ]]; then
    OVERALL_RESULT="fail"
    REASON_CODE="unexpected_exit_code"
    RECOMMENDED_ACTION="block_validation"
    return
  fi
  if [[ "${VERIFY_FAIL_COUNT}" -gt 0 ]]; then
    OVERALL_RESULT="fail"
    REASON_CODE="verification_failed"
    RECOMMENDED_ACTION="block_validation"
    return
  fi
  if [[ "${WARN_STEPS}" -gt 0 ]]; then
    OVERALL_RESULT="warn"
    REASON_CODE="allowed_non_zero_exit"
  fi
}

write_summary_json() {
  local summary="${OUTPUT}/summary.json"
  local total="${#STEP_NAMES[@]}"
  local i
  compute_outcome

  local legacy_overall="passed"
  local legacy_failed=0
  if [[ "${OVERALL_RESULT}" == "warn" ]]; then
    legacy_overall="warn"
    legacy_failed="${WARN_STEPS}"
  fi
  if [[ "${OVERALL_RESULT}" == "fail" ]]; then
    legacy_overall="failed"
    legacy_failed=$((UNEXPECTED_STEPS + VERIFY_FAIL_COUNT))
  fi

  {
    echo "{"
    echo "  \"schemaVersion\": \"v1\","
    echo "  \"result\": \"${OVERALL_RESULT}\","
    echo "  \"reasonCode\": \"${REASON_CODE}\","
    echo "  \"recommendedAction\": \"${RECOMMENDED_ACTION}\","
    echo "  \"overall\": \"${legacy_overall}\","
    echo "  \"generatedAt\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\","
    echo "  \"dryRun\": $(bool_json "${DRY_RUN}"),"
    echo "  \"strictExit\": $(bool_json "${STRICT_EXIT}"),"
    echo "  \"failOnUnexpected\": $(bool_json "${FAIL_ON_UNEXPECTED}"),"
    echo "  \"output\": \"$(json_escape "${OUTPUT}")\","
    echo "  \"statusJsonFile\": \"$(json_escape "${JSON_STATUS_FILE}")\","
    echo "  \"totalSteps\": ${total},"
    echo "  \"failedSteps\": ${legacy_failed},"
    echo "  \"unexpectedSteps\": ${UNEXPECTED_STEPS},"
    echo "  \"warningSteps\": ${WARN_STEPS},"
    echo "  \"verifyFailCount\": ${VERIFY_FAIL_COUNT},"
    echo "  \"steps\": ["
    for ((i = 0; i < total; i++)); do
      local name="${STEP_NAMES[$i]}"
      local code="${STEP_CODES[$i]}"
      local expected="${STEP_EXPECTED[$i]}"
      local allowed="${STEP_ALLOWED[$i]}"
      local status="passed"
      if [[ "${allowed}" != "true" ]]; then
        status="failed"
      elif [[ "${code}" != "0" ]]; then
        status="warn"
      fi
      local comma=","
      if [[ $i -eq $((total - 1)) ]]; then
        comma=""
      fi
      echo "    {\"name\":\"$(json_escape "${name}")\",\"exitCode\":${code},\"expectedExitCodes\":\"${expected}\",\"allowed\":${allowed},\"status\":\"${status}\"}${comma}"
    done
    echo "  ],"
    echo "  \"verifyFailures\": ["
    local vf_total="${#VERIFY_FAILURES[@]}"
    for ((i = 0; i < vf_total; i++)); do
      local v="${VERIFY_FAILURES[$i]}"
      local comma=","
      if [[ $i -eq $((vf_total - 1)) ]]; then
        comma=""
      fi
      echo "    \"$(json_escape "${v}")\"${comma}"
    done
    echo "  ]"
    echo "}"
  } > "${summary}"

  echo "summary:   ${summary}"
}

echo "== OpsKit generic-manage e2e =="
echo "BIN=${BIN}"
echo "OUTPUT=${OUTPUT}"
echo "DRY_RUN=${DRY_RUN}"
echo "STRICT_EXIT=${STRICT_EXIT}"
echo "FAIL_ON_UNEXPECTED=${FAIL_ON_UNEXPECTED}"
echo "JSON_STATUS_FILE=${JSON_STATUS_FILE}"

mkdir -p "${OUTPUT}" "$(dirname "${JSON_STATUS_FILE}")"

stage_expected="$(expected_stage_codes)"

run_and_capture validate "0" "${BIN_CMD[@]}" template validate templates/builtin/default-manage.json

if [[ "${DRY_RUN}" == "1" ]]; then
  run_and_capture install-dry "0" "${BIN_CMD[@]}" install --template generic-manage-v1 --dry-run --no-systemd --output "${OUTPUT}"
  run_and_capture run-af-dry "0" "${BIN_CMD[@]}" run AF --template generic-manage-v1 --dry-run --output "${OUTPUT}"
  run_status_json_capture status-dry-json "0" "${BIN_CMD[@]}" status --output "${OUTPUT}"
else
  run_and_capture install "${stage_expected}" "${BIN_CMD[@]}" install --template generic-manage-v1 --no-systemd --output "${OUTPUT}"
  run_and_capture run-af "${stage_expected}" "${BIN_CMD[@]}" run AF --template generic-manage-v1 --output "${OUTPUT}"
  run_and_capture status "${stage_expected}" "${BIN_CMD[@]}" status --output "${OUTPUT}"
  run_status_json_capture status-json "${stage_expected}" "${BIN_CMD[@]}" status --output "${OUTPUT}"
  run_and_capture accept "${stage_expected}" "${BIN_CMD[@]}" accept --template generic-manage-v1 --output "${OUTPUT}"
  run_and_capture handover "0" "${BIN_CMD[@]}" handover --output "${OUTPUT}"

  require_file "${OUTPUT}/state/overall.json"
  require_file "${OUTPUT}/state/lifecycle.json"
  require_file "${OUTPUT}/state/services.json"
  require_file "${OUTPUT}/state/artifacts.json"
  require_file "${JSON_STATUS_FILE}"
  require_grep "\"schemaVersion\"[[:space:]]*:[[:space:]]*\"v1\"" "${JSON_STATUS_FILE}"
  require_grep "\"command\"[[:space:]]*:[[:space:]]*\"opskit status\"" "${JSON_STATUS_FILE}"
  require_grep "\"health\"[[:space:]]*:[[:space:]]*\"(ok|warn|fail)\"" "${JSON_STATUS_FILE}"
fi

echo ""
echo "== Verification hints =="
echo "state:     ${OUTPUT}/state/overall.json"
echo "lifecycle: ${OUTPUT}/state/lifecycle.json"
echo "artifacts: ${OUTPUT}/state/artifacts.json"
echo "status:    ${JSON_STATUS_FILE}"
echo "reports:   ${OUTPUT}/reports"
echo "bundles:   ${OUTPUT}/bundles"
echo "ui:        ${OUTPUT}/ui/index.html"
write_summary_json

if [[ -f "${OUTPUT}/state/overall.json" && "${DRY_RUN}" != "1" ]]; then
  echo ""
  echo "overall.json:"
  sed -n '1,120p' "${OUTPUT}/state/overall.json"
fi

if [[ "${FAIL_ON_UNEXPECTED}" == "1" ]]; then
  if [[ "${OVERALL_RESULT}" == "fail" ]]; then
    exit 1
  fi
fi

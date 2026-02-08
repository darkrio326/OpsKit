#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Run machine-readable template validation checks.

Usage:
  scripts/template-validate-check.sh [options]

Options:
  -o, --output <dir>      Output directory (default: ./.tmp/template-validate-check)
      --bin <path>        OpsKit binary path (default: <output>/opskit-template-check)
      --clean             Remove output directory before running
      --skip-build        Skip binary build step
  -h, --help              Show help

Environment:
  GO_CACHE_DIR            Go build cache dir (default: ./.tmp/gocache-template-validate-check)

Examples:
  scripts/template-validate-check.sh --clean
  scripts/template-validate-check.sh --bin ./opskit --skip-build
USAGE
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/.tmp/template-validate-check"
BIN_PATH=""
SKIP_BUILD=0
CLEAN=0
GO_CACHE_DIR="${GO_CACHE_DIR:-${ROOT_DIR}/.tmp/gocache-template-validate-check}"

FAIL_COUNT=0
FAIL_REASONS=()
RESULT="pass"
RECOMMENDED_ACTION="continue_ci"
REASON_CODE="ok"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --bin)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      BIN_PATH="$2"
      shift 2
      ;;
    --clean)
      CLEAN=1
      shift
      ;;
    --skip-build)
      SKIP_BUILD=1
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

if [[ -z "${BIN_PATH}" ]]; then
  BIN_PATH="${OUTPUT_DIR}/opskit-template-check"
fi

if [[ "${SKIP_BUILD}" == "0" ]]; then
  echo "==> build binary"
  (
    cd "${ROOT_DIR}"
    env GOCACHE="${GO_CACHE_DIR}" go build -o "${BIN_PATH}" ./cmd/opskit
  )
fi

if [[ ! -x "${BIN_PATH}" ]]; then
  echo "binary not executable: ${BIN_PATH}" >&2
  exit 2
fi

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

run_json_validate_ok() {
  local name="$1"
  local template="$2"
  local vars_file="$3"
  local out_file="$4"
  echo "==> validate ${name}"
  "${BIN_PATH}" template validate --json --vars-file "${vars_file}" "${template}" > "${out_file}"
  require_pattern "${out_file}" '"command"[[:space:]]*:[[:space:]]*"opskit template validate"' "${name}:missing_command"
  require_pattern "${out_file}" '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "${name}:missing_schema_version"
  require_pattern "${out_file}" '"template"[[:space:]]*:[[:space:]]*".*"' "${name}:missing_template"
  require_pattern "${out_file}" '"valid"[[:space:]]*:[[:space:]]*true' "${name}:expected_valid_true"
  require_pattern "${out_file}" '"errorCount"[[:space:]]*:[[:space:]]*0' "${name}:expected_error_count_zero"
}

run_json_validate_ok_no_vars() {
  local name="$1"
  local template="$2"
  local out_file="$3"
  echo "==> validate ${name}"
  "${BIN_PATH}" template validate --json "${template}" > "${out_file}"
  require_pattern "${out_file}" '"command"[[:space:]]*:[[:space:]]*"opskit template validate"' "${name}:missing_command"
  require_pattern "${out_file}" '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "${name}:missing_schema_version"
  require_pattern "${out_file}" '"template"[[:space:]]*:[[:space:]]*".*"' "${name}:missing_template"
  require_pattern "${out_file}" '"valid"[[:space:]]*:[[:space:]]*true' "${name}:expected_valid_true"
  require_pattern "${out_file}" '"errorCount"[[:space:]]*:[[:space:]]*0' "${name}:expected_error_count_zero"
}

run_json_validate_fail() {
  local name="$1"
  local template="$2"
  local out_file="$3"
  echo "==> validate ${name} (negative)"
  set +e
  "${BIN_PATH}" template validate --json "${template}" > "${out_file}"
  local rc=$?
  set -e
  if [[ "${rc}" == "0" ]]; then
    add_failure "${name}:expected_non_zero_exit"
    return 0
  fi
  require_pattern "${out_file}" '"valid"[[:space:]]*:[[:space:]]*false' "${name}:expected_valid_false"
  require_pattern "${out_file}" '"errorCount"[[:space:]]*:[[:space:]]*[1-9][0-9]*' "${name}:expected_error_count_positive"
  require_pattern "${out_file}" '"code"[[:space:]]*:[[:space:]]*"template_file_not_found"' "${name}:expected_file_not_found_code"
}

run_json_validate_fail_with_vars() {
  local name="$1"
  local template="$2"
  local vars_file="$3"
  local out_file="$4"
  local expected_code="$5"
  local expected_path="$6"
  echo "==> validate ${name} (negative vars)"
  set +e
  "${BIN_PATH}" template validate --json --vars-file "${vars_file}" "${template}" > "${out_file}"
  local rc=$?
  set -e
  if [[ "${rc}" == "0" ]]; then
    add_failure "${name}:expected_non_zero_exit"
    return 0
  fi
  require_pattern "${out_file}" '"valid"[[:space:]]*:[[:space:]]*false' "${name}:expected_valid_false"
  require_pattern "${out_file}" '"errorCount"[[:space:]]*:[[:space:]]*[1-9][0-9]*' "${name}:expected_error_count_positive"
  require_pattern "${out_file}" "\"code\"[[:space:]]*:[[:space:]]*\"${expected_code}\"" "${name}:expected_code_${expected_code}"
  require_pattern "${out_file}" "\"path\"[[:space:]]*:[[:space:]]*\"${expected_path}\"" "${name}:expected_path_${expected_path}"
}

AUDIT_JSON="${OUTPUT_DIR}/demo-server-audit.json.out"
HELLO_JSON="${OUTPUT_DIR}/demo-hello-service.json.out"
RUNTIME_JSON="${OUTPUT_DIR}/demo-runtime-baseline.json.out"
BLACKBOX_DEFAULT_JSON="${OUTPUT_DIR}/demo-blackbox-middleware-manage.json.out"
BLACKBOX_FCS_JSON="${OUTPUT_DIR}/demo-blackbox-middleware-manage-fcs.json.out"
BLACKBOX_KINGDEE_JSON="${OUTPUT_DIR}/demo-blackbox-middleware-manage-kingdee.json.out"
MANAGE_JSON="${OUTPUT_DIR}/generic-manage-v1.json.out"
NEG_JSON="${OUTPUT_DIR}/missing-template.json.out"
NEG_VARS_JSON="${OUTPUT_DIR}/invalid-vars.json.out"
SUMMARY_JSON="${OUTPUT_DIR}/summary.json"

run_json_validate_ok \
  "demo-server-audit" \
  "${ROOT_DIR}/assets/templates/demo-server-audit.json" \
  "${ROOT_DIR}/examples/vars/demo-server-audit.json" \
  "${AUDIT_JSON}"

run_json_validate_ok \
  "demo-hello-service" \
  "${ROOT_DIR}/assets/templates/demo-hello-service.json" \
  "${ROOT_DIR}/examples/vars/demo-hello-service.json" \
  "${HELLO_JSON}"

run_json_validate_ok \
  "demo-runtime-baseline" \
  "${ROOT_DIR}/assets/templates/demo-runtime-baseline.json" \
  "${ROOT_DIR}/examples/vars/demo-runtime-baseline.json" \
  "${RUNTIME_JSON}"

run_json_validate_ok \
  "demo-blackbox-middleware-manage-default" \
  "${ROOT_DIR}/assets/templates/demo-blackbox-middleware-manage.json" \
  "${ROOT_DIR}/examples/vars/demo-blackbox-middleware-manage.json" \
  "${BLACKBOX_DEFAULT_JSON}"

run_json_validate_ok \
  "demo-blackbox-middleware-manage-fcs" \
  "${ROOT_DIR}/assets/templates/demo-blackbox-middleware-manage.json" \
  "${ROOT_DIR}/examples/vars/demo-blackbox-middleware-manage-fcs.json" \
  "${BLACKBOX_FCS_JSON}"

run_json_validate_ok \
  "demo-blackbox-middleware-manage-kingdee" \
  "${ROOT_DIR}/assets/templates/demo-blackbox-middleware-manage.json" \
  "${ROOT_DIR}/examples/vars/demo-blackbox-middleware-manage-kingdee.json" \
  "${BLACKBOX_KINGDEE_JSON}"

run_json_validate_ok_no_vars \
  "generic-manage-v1" \
  "generic-manage-v1" \
  "${MANAGE_JSON}"

run_json_validate_fail \
  "missing-template" \
  "${OUTPUT_DIR}/missing-template.json" \
  "${NEG_JSON}"

INVALID_VARS_FILE="${OUTPUT_DIR}/invalid-vars.json"
cat > "${INVALID_VARS_FILE}" <<'EOF'
{
  "SERVICE_NAME": "hello-service",
  "SERVICE_PORT": "oops",
  "SERVICE_UNIT": "hello-service.service",
  "PROCESS_MATCH": "hello-service",
  "INSTALL_ROOT": "/data/opskit-demo",
  "CONF_DIR": "/etc/opskit"
}
EOF

run_json_validate_fail_with_vars \
  "invalid-vars-type" \
  "${ROOT_DIR}/assets/templates/demo-hello-service.json" \
  "${INVALID_VARS_FILE}" \
  "${NEG_VARS_JSON}" \
  "template_var_type_mismatch" \
  "template.vars.SERVICE_PORT"

if [[ "${FAIL_COUNT}" -gt 0 ]]; then
  RESULT="fail"
  RECOMMENDED_ACTION="block_ci"
  REASON_CODE="template_validate_contract_failed"
fi

{
  echo "{"
  echo "  \"schemaVersion\": \"v1\","
  echo "  \"result\": \"${RESULT}\","
  echo "  \"reasonCode\": \"${REASON_CODE}\","
  echo "  \"recommendedAction\": \"${RECOMMENDED_ACTION}\","
  echo "  \"failCount\": ${FAIL_COUNT},"
  echo "  \"outputs\": ["
  echo "    \"${AUDIT_JSON}\","
  echo "    \"${HELLO_JSON}\","
  echo "    \"${RUNTIME_JSON}\","
  echo "    \"${BLACKBOX_DEFAULT_JSON}\","
  echo "    \"${BLACKBOX_FCS_JSON}\","
  echo "    \"${BLACKBOX_KINGDEE_JSON}\","
  echo "    \"${MANAGE_JSON}\","
  echo "    \"${NEG_JSON}\","
  echo "    \"${NEG_VARS_JSON}\""
  echo "  ],"
  echo "  \"failures\": ["
  for i in "${!FAIL_REASONS[@]}"; do
    sep=","
    if [[ "${i}" -eq $((${#FAIL_REASONS[@]} - 1)) ]]; then
      sep=""
    fi
    echo "    \"${FAIL_REASONS[$i]}\"${sep}"
  done
  echo "  ]"
  echo "}"
} > "${SUMMARY_JSON}.tmp"
mv "${SUMMARY_JSON}.tmp" "${SUMMARY_JSON}"

echo ""
echo "template-validate-check summary"
echo "- result: ${RESULT}"
echo "- reason: ${REASON_CODE}"
echo "- recommended action: ${RECOMMENDED_ACTION}"
echo "- summary: ${SUMMARY_JSON}"

if [[ "${RESULT}" != "pass" ]]; then
  exit 1
fi

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
      --clean             Remove output directory before running
  -h, --help              Show help

Notes:
  - This script runs real commands (not dry-run): run A, run D, accept, status.
  - Expected stage/status exit codes are 0/1/3 (pass/fail/warn).
USAGE
}

BIN="${BIN:-opskit}"
OUTPUT="${OUTPUT:-/data/opskit-regression-v034}"
TEMPLATE="${TEMPLATE:-generic-manage-v1}"
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

RESULT_LINES=()
FAIL_COUNT=0
HARD_FAIL=0

mark_result() {
  local name="$1"
  local rc="$2"
  RESULT_LINES+=("${name}|${rc}")
}

is_allowed_stage_rc() {
  local rc="$1"
  [[ "${rc}" == "0" || "${rc}" == "1" || "${rc}" == "3" ]]
}

run_expect_zero() {
  local name="$1"
  shift
  echo "==> ${name}"
  "${BIN}" "$@"
  local rc=$?
  mark_result "${name}" "${rc}"
  if [[ "${rc}" != "0" ]]; then
    echo "    failed: expected rc=0, got ${rc}" >&2
    HARD_FAIL=1
  else
    echo "    ok: rc=${rc}"
  fi
}

run_expect_stage_rc() {
  local name="$1"
  shift
  echo "==> ${name}"
  set +e
  "${BIN}" "$@"
  local rc=$?
  set -e
  mark_result "${name}" "${rc}"
  if is_allowed_stage_rc "${rc}"; then
    echo "    ok: rc=${rc}"
  else
    echo "    failed: unexpected rc=${rc} (expect 0/1/3)" >&2
    HARD_FAIL=1
  fi
}

require_file() {
  local path="$1"
  if [[ -f "${path}" ]]; then
    echo "    ok: ${path}"
    return
  fi
  echo "    missing: ${path}" >&2
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
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

set -e
run_expect_zero "template validate (${TEMPLATE})" template validate "${TEMPLATE}" --output "${OUTPUT}"
run_expect_stage_rc "run A" run A --template "${TEMPLATE}" --output "${OUTPUT}"
run_expect_stage_rc "run D" run D --template "${TEMPLATE}" --output "${OUTPUT}"
run_expect_stage_rc "accept" accept --template "${TEMPLATE}" --output "${OUTPUT}"
run_expect_stage_rc "status" status --output "${OUTPUT}"

echo "==> verify outputs"
require_file "${OUTPUT}/state/overall.json"
require_file "${OUTPUT}/state/lifecycle.json"
require_file "${OUTPUT}/state/services.json"
require_file "${OUTPUT}/state/artifacts.json"

require_grep '"summary"' "${OUTPUT}/state/lifecycle.json"
require_grep 'acceptance-consistency-' "${OUTPUT}/state/artifacts.json"

latest_accept="$(ls -1t "${OUTPUT}"/reports/accept-*.html 2>/dev/null | head -n1 || true)"
if [[ -z "${latest_accept}" ]]; then
  echo "    missing: ${OUTPUT}/reports/accept-*.html" >&2
  FAIL_COUNT=$((FAIL_COUNT + 1))
else
  echo "    ok: ${latest_accept}"
  require_grep '"consistency"' "${latest_accept}"
fi

echo ""
echo "offline validation summary"
for line in "${RESULT_LINES[@]}"; do
  name="${line%%|*}"
  rc="${line##*|}"
  echo "- ${name}: rc=${rc}"
done
echo "- output: ${OUTPUT}"

if [[ "${HARD_FAIL}" == "1" || "${FAIL_COUNT}" -gt 0 ]]; then
  echo "offline validation failed (hard_fail=${HARD_FAIL}, verify_fail=${FAIL_COUNT})" >&2
  exit 1
fi

echo "offline validation passed"

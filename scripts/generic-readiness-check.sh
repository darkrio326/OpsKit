#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Run a unified generic readiness gate before real server validation.

Usage:
  scripts/generic-readiness-check.sh [options]

Options:
  -o, --output <dir>      Output root (default: ./.tmp/generic-readiness-check)
      --bin <path>        OpsKit binary path (default: auto-build to <output>/opskit-gate)
      --clean             Remove output directory before running
      --skip-tests        Skip go test ./...
      --skip-release      Skip scripts/release-check.sh gate
      --skip-generic      Skip examples/generic-manage/run-af.sh gate
      --offline-strict    Pass --offline-strict-exit to release-check
      --generic-strict    Require generic-manage stage/status exit code = 0
  -h, --help              Show help

Environment:
  GO_CACHE_DIR            Go cache dir (default: ./.tmp/gocache-generic-readiness)

Examples:
  scripts/generic-readiness-check.sh --clean
  scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean
USAGE
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/.tmp/generic-readiness-check"
BIN_PATH=""
CLEAN=0
SKIP_TESTS=0
SKIP_RELEASE=0
SKIP_GENERIC=0
OFFLINE_STRICT=0
GENERIC_STRICT=0
GO_CACHE_DIR="${GO_CACHE_DIR:-${ROOT_DIR}/.tmp/gocache-generic-readiness}"

STEP_COUNT=0
TOTAL_SECONDS=0
STEP_LINES=()
VERIFY_FAIL_COUNT=0
VERIFY_FAILURES=()
RECOMMENDED_ACTION="continue_real_server_validation"
FINAL_RESULT="pass"
FINAL_REASON="ok"

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
    --skip-tests)
      SKIP_TESTS=1
      shift
      ;;
    --skip-release)
      SKIP_RELEASE=1
      shift
      ;;
    --skip-generic)
      SKIP_GENERIC=1
      shift
      ;;
    --offline-strict)
      OFFLINE_STRICT=1
      shift
      ;;
    --generic-strict)
      GENERIC_STRICT=1
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

add_verify_failure() {
  local msg="$1"
  VERIFY_FAILURES+=("${msg}")
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

run_step() {
  local label="$1"
  shift
  local started ended elapsed rc
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
  STEP_LINES+=("${label}|${elapsed}|${rc}")
  if [[ "${rc}" == "0" ]]; then
    echo "    done: ${elapsed}s"
    return 0
  fi
  echo "    failed: ${elapsed}s rc=${rc}" >&2
  FINAL_RESULT="fail"
  FINAL_REASON="step_failed"
  RECOMMENDED_ACTION="block_real_server_validation"
  return "${rc}"
}

write_summary_json() {
  local summary_file="$1"
  local tmp_file="${summary_file}.tmp"
  local i

  {
    echo "{"
    echo "  \"schemaVersion\": \"v1\"," 
    echo "  \"generatedAt\": \"$(now_iso)\"," 
    echo "  \"result\": \"${FINAL_RESULT}\"," 
    echo "  \"reasonCode\": \"${FINAL_REASON}\"," 
    echo "  \"recommendedAction\": \"${RECOMMENDED_ACTION}\"," 
    echo "  \"output\": \"$(json_escape "${OUTPUT_DIR}")\"," 
    echo "  \"binary\": \"$(json_escape "${BIN_PATH}")\"," 
    echo "  \"goCacheDir\": \"$(json_escape "${GO_CACHE_DIR}")\"," 
    echo "  \"strictGeneric\": $(bool_json "${GENERIC_STRICT}"),"
    echo "  \"strictOffline\": $(bool_json "${OFFLINE_STRICT}"),"
    echo "  \"steps\": ${STEP_COUNT},"
    echo "  \"totalDurationSeconds\": ${TOTAL_SECONDS},"
    echo "  \"verifyFailCount\": ${VERIFY_FAIL_COUNT},"
    echo "  \"stepResults\": ["

    for i in "${!STEP_LINES[@]}"; do
      local line label elapsed rc comma
      IFS='|' read -r label elapsed rc <<< "${STEP_LINES[$i]}"
      comma=","
      if [[ "${i}" -eq $((${#STEP_LINES[@]} - 1)) ]]; then
        comma=""
      fi
      echo "    {\"name\":\"$(json_escape "${label}")\",\"elapsedSeconds\":${elapsed},\"exitCode\":${rc}}${comma}"
    done

    echo "  ],"
    echo "  \"verifyFailures\": ["

    for i in "${!VERIFY_FAILURES[@]}"; do
      local comma=","
      if [[ "${i}" -eq $((${#VERIFY_FAILURES[@]} - 1)) ]]; then
        comma=""
      fi
      echo "    \"$(json_escape "${VERIFY_FAILURES[$i]}")\"${comma}"
    done

    echo "  ]"
    echo "}"
  } > "${tmp_file}"

  mv "${tmp_file}" "${summary_file}"
}

if [[ "${CLEAN}" == "1" ]]; then
  rm -rf "${OUTPUT_DIR}"
fi
mkdir -p "${OUTPUT_DIR}" "${GO_CACHE_DIR}"

if [[ -z "${BIN_PATH}" ]]; then
  BIN_PATH="${OUTPUT_DIR}/opskit-gate"
fi

RELEASE_OUT="${OUTPUT_DIR}/release-check"
GENERIC_OUT="${OUTPUT_DIR}/generic-manage"
GENERIC_STATUS_JSON="${GENERIC_OUT}/status.json"
GENERIC_SUMMARY_JSON="${GENERIC_OUT}/summary.json"
SUMMARY_JSON="${OUTPUT_DIR}/summary.json"

if [[ ! -x "${BIN_PATH}" ]]; then
  run_step "build gate binary" env GOCACHE="${GO_CACHE_DIR}" go build -o "${BIN_PATH}" ./cmd/opskit
fi

if [[ "${SKIP_TESTS}" == "0" ]]; then
  run_step "go test ./..." env GOCACHE="${GO_CACHE_DIR}" go test ./...
fi

if [[ "${SKIP_RELEASE}" == "0" ]]; then
  release_args=(
    --output "${RELEASE_OUT}"
    --skip-tests
    --with-offline-validate
    --offline-bin "${BIN_PATH}"
    --offline-output "${RELEASE_OUT}/offline-validate"
    --offline-json-status-file "${RELEASE_OUT}/offline-validate/status.json"
    --offline-summary-json-file "${RELEASE_OUT}/offline-validate/summary.json"
  )
  if [[ "${OFFLINE_STRICT}" == "1" ]]; then
    release_args+=(--offline-strict-exit)
  fi
  run_step "release-check gate" ./scripts/release-check.sh "${release_args[@]}"
  require_file "${RELEASE_OUT}/summary.json"
  require_grep '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "${RELEASE_OUT}/summary.json"
  require_grep '"result"[[:space:]]*:[[:space:]]*"pass"' "${RELEASE_OUT}/summary.json"
  require_grep '"reasonCode"[[:space:]]*:[[:space:]]*"ok"' "${RELEASE_OUT}/summary.json"
  require_grep '"recommendedAction"[[:space:]]*:[[:space:]]*"continue_release"' "${RELEASE_OUT}/summary.json"
fi

if [[ "${SKIP_GENERIC}" == "0" ]]; then
  run_step "generic-manage gate" env \
    BIN="${BIN_PATH}" \
    OUTPUT="${GENERIC_OUT}" \
    DRY_RUN=0 \
    STRICT_EXIT="${GENERIC_STRICT}" \
    FAIL_ON_UNEXPECTED=1 \
    JSON_STATUS_FILE="${GENERIC_STATUS_JSON}" \
    ./examples/generic-manage/run-af.sh

  require_file "${GENERIC_STATUS_JSON}"
  require_file "${GENERIC_SUMMARY_JSON}"
  require_grep '"schemaVersion"[[:space:]]*:[[:space:]]*"v1"' "${GENERIC_STATUS_JSON}"
  require_grep '"health"[[:space:]]*:[[:space:]]*"(ok|warn|fail)"' "${GENERIC_STATUS_JSON}"

  if [[ "${GENERIC_STRICT}" == "1" ]]; then
    require_grep '"result"[[:space:]]*:[[:space:]]*"pass"' "${GENERIC_SUMMARY_JSON}"
  else
    require_grep '"recommendedAction"[[:space:]]*:[[:space:]]*"continue_validation"' "${GENERIC_SUMMARY_JSON}"
  fi
fi

if [[ "${VERIFY_FAIL_COUNT}" -gt 0 ]]; then
  FINAL_RESULT="fail"
  FINAL_REASON="verification_failed"
  RECOMMENDED_ACTION="block_real_server_validation"
elif [[ "${FINAL_RESULT}" == "pass" ]]; then
  FINAL_REASON="ok"
  RECOMMENDED_ACTION="continue_real_server_validation"
fi

write_summary_json "${SUMMARY_JSON}"

echo ""
echo "generic-readiness summary"
echo "- result: ${FINAL_RESULT}"
echo "- reason: ${FINAL_REASON}"
echo "- recommended action: ${RECOMMENDED_ACTION}"
echo "- summary json: ${SUMMARY_JSON}"
echo "- release-check output: ${RELEASE_OUT}"
echo "- generic output: ${GENERIC_OUT}"

if [[ "${FINAL_RESULT}" == "fail" ]]; then
  exit 1
fi

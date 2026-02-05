#!/usr/bin/env bash
set -euo pipefail

BIN="${BIN:-opskit}"
OUTPUT="${OUTPUT:-/tmp/opskit-generic}"
DRY_RUN="${DRY_RUN:-0}"
read -r -a BIN_CMD <<< "${BIN}"
STEP_NAMES=()
STEP_CODES=()

if [[ "${1:-}" == "--help" ]]; then
  cat <<'EOF'
Run generic-manage end-to-end checks (A~F) and print verification hints.

Environment variables:
  BIN      opskit binary path or command name (default: opskit)
  OUTPUT   output root directory (default: /tmp/opskit-generic)
  DRY_RUN  set to 1 to execute dry-run flow

Examples:
  BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
  DRY_RUN=1 ./examples/generic-manage/run-af.sh
EOF
  exit 0
fi

run_and_capture() {
  local name="$1"
  shift
  STEP_NAMES+=("${name}")
  echo ""
  echo "[${name}] $*"
  set +e
  "$@"
  local code=$?
  set -e
  STEP_CODES+=("${code}")
  echo "[${name}] exit=${code}"
  return 0
}

write_summary_json() {
  local summary="${OUTPUT}/summary.json"
  local overall="passed"
  local total="${#STEP_NAMES[@]}"
  local failed=0

  for i in "${!STEP_NAMES[@]}"; do
    if [[ "${STEP_CODES[$i]}" != "0" ]]; then
      overall="failed"
      failed=$((failed + 1))
    fi
  done

  {
    echo "{"
    echo "  \"overall\": \"${overall}\","
    echo "  \"generatedAt\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\","
    echo "  \"dryRun\": $( [[ "${DRY_RUN}" == "1" ]] && echo "true" || echo "false" ),"
    echo "  \"output\": \"${OUTPUT}\","
    echo "  \"totalSteps\": ${total},"
    echo "  \"failedSteps\": ${failed},"
    echo "  \"steps\": ["
    for i in "${!STEP_NAMES[@]}"; do
      local name="${STEP_NAMES[$i]}"
      local code="${STEP_CODES[$i]}"
      local status="passed"
      if [[ "${code}" != "0" ]]; then
        status="failed"
      fi
      local comma=","
      if [[ $i -eq $((total - 1)) ]]; then
        comma=""
      fi
      echo "    {\"name\":\"${name}\",\"exitCode\":${code},\"status\":\"${status}\"}${comma}"
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

mkdir -p "${OUTPUT}"

run_and_capture validate "${BIN_CMD[@]}" template validate templates/builtin/default-manage.json

if [[ "${DRY_RUN}" == "1" ]]; then
  run_and_capture install-dry "${BIN_CMD[@]}" install --template generic-manage-v1 --dry-run --no-systemd --output "${OUTPUT}"
  run_and_capture run-af-dry "${BIN_CMD[@]}" run AF --template generic-manage-v1 --dry-run --output "${OUTPUT}"
else
  run_and_capture install "${BIN_CMD[@]}" install --template generic-manage-v1 --no-systemd --output "${OUTPUT}"
  run_and_capture run-af "${BIN_CMD[@]}" run AF --template generic-manage-v1 --output "${OUTPUT}"
  run_and_capture status "${BIN_CMD[@]}" status --output "${OUTPUT}"
  run_and_capture accept "${BIN_CMD[@]}" accept --template generic-manage-v1 --output "${OUTPUT}"
  run_and_capture handover "${BIN_CMD[@]}" handover --output "${OUTPUT}"
fi

echo ""
echo "== Verification hints =="
echo "state:     ${OUTPUT}/state/overall.json"
echo "lifecycle: ${OUTPUT}/state/lifecycle.json"
echo "artifacts: ${OUTPUT}/state/artifacts.json"
echo "reports:   ${OUTPUT}/reports"
echo "bundles:   ${OUTPUT}/bundles"
echo "ui:        ${OUTPUT}/ui/index.html"
write_summary_json

if [[ -f "${OUTPUT}/state/overall.json" ]]; then
  echo ""
  echo "overall.json:"
  sed -n '1,120p' "${OUTPUT}/state/overall.json"
fi

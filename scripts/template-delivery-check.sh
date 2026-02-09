#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Run delivery gate chain for all standard templates:
  run A -> run D -> accept -> status --json

This script intentionally allows non-zero stage exits (0/1/3) and validates
that artifacts can still be delivered (state/report/bundle/status json).

Usage:
  scripts/template-delivery-check.sh [--bin <path>] [--output <dir>] [--clean]

Options:
  --bin <path>     OpsKit binary path (default: auto build to <output>/opskit)
  --output <dir>   Output root (default: ./.tmp/template-delivery-check)
  --clean          Remove output directory before run
EOF
}

BIN=""
OUT_DIR="${PWD}/.tmp/template-delivery-check"
CLEAN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bin)
      BIN="${2:-}"
      shift 2
      ;;
    --output)
      OUT_DIR="${2:-}"
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
      echo "unknown arg: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ $CLEAN -eq 1 ]]; then
  rm -rf "$OUT_DIR"
fi
mkdir -p "$OUT_DIR"

if [[ -z "$BIN" ]]; then
  BIN="$OUT_DIR/opskit"
  mkdir -p "${PWD}/.tmp/gocache"
  GOCACHE="${PWD}/.tmp/gocache" go build -o "$BIN" ./cmd/opskit
fi

if [[ ! -x "$BIN" ]]; then
  echo "binary not executable: $BIN" >&2
  exit 2
fi

declare -a templates=(
  "generic-manage-v1|"
  "single-service-deploy-v1|"
  "assets/templates/demo-server-audit.json|examples/vars/demo-server-audit.json"
  "assets/templates/demo-runtime-baseline.json|examples/vars/demo-runtime-baseline.json"
  "assets/templates/demo-blackbox-middleware-manage.json|examples/vars/demo-blackbox-middleware-manage.json"
  "assets/templates/demo-generic-selfhost-deploy.json|examples/vars/demo-generic-selfhost-deploy.json"
  "assets/templates/demo-hello-service.json|examples/vars/demo-hello-service.json"
  "assets/templates/demo-minio-deploy.json|examples/vars/demo-minio-deploy.json"
  "assets/templates/demo-elk-deploy.json|examples/vars/demo-elk-deploy.json"
  "assets/templates/demo-powerjob-deploy.json|examples/vars/demo-powerjob-deploy.json"
)

SUMMARY="$OUT_DIR/summary.tsv"
echo -e "template\ta_rc\td_rc\taccept_rc\tbundle\treport\tstatus_json" > "$SUMMARY"

set +e
for item in "${templates[@]}"; do
  IFS='|' read -r tpl vars <<< "$item"
  name="$(basename "$tpl" .json)"
  run_out="$OUT_DIR/$name"
  rm -rf "$run_out"
  mkdir -p "$run_out"
  local_root="$run_out/runtime"

  args=(--template "$tpl" --output "$run_out")
  if [[ -n "$vars" ]]; then
    args+=(--vars-file "$vars")
  fi
  # Enforce writable local paths so check can run on non-root/dev hosts.
  args+=(--vars "INSTALL_ROOT=$local_root,CONF_DIR=$local_root/conf,EVIDENCE_DIR=$local_root/evidence,DATA_DIR=$local_root/data,LOG_DIR=$local_root/log,APP_DIR=$local_root/app")

  "$BIN" run A "${args[@]}" >/dev/null 2>&1
  a_rc=$?
  "$BIN" run D "${args[@]}" >/dev/null 2>&1
  d_rc=$?
  "$BIN" accept "${args[@]}" >/dev/null 2>&1
  f_rc=$?
  "$BIN" status --output "$run_out" --json > "$run_out/status.json" 2>/dev/null

  bundle=0
  report=0
  status_json=0
  ls "$run_out"/bundles/acceptance-*.tar.gz >/dev/null 2>&1 && bundle=1
  ls "$run_out"/reports/accept-*.html >/dev/null 2>&1 && report=1
  [[ -s "$run_out/status.json" ]] && status_json=1

  echo -e "$tpl\t$a_rc\t$d_rc\t$f_rc\t$bundle\t$report\t$status_json" >> "$SUMMARY"
done
set -e

echo "delivery gate summary:"
column -t -s $'\t' "$SUMMARY"
echo
echo "summary file: $SUMMARY"

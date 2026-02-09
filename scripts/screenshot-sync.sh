#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Sync screenshot slots from latest/ to a release version directory.

Usage:
  scripts/screenshot-sync.sh --version <version>

Example:
  scripts/screenshot-sync.sh --version v0.4.2-preview.1
EOF
}

VERSION=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      [[ $# -ge 2 ]] || { echo "missing value for --version" >&2; exit 2; }
      VERSION="$2"
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

if [[ -z "${VERSION}" ]]; then
  echo "--version is required" >&2
  usage >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="${ROOT_DIR}/docs/assets/screenshots/latest"
DST_DIR="${ROOT_DIR}/docs/assets/screenshots/releases/${VERSION}"

required=(
  "ui-template-stage.png"
  "ui-dashboard-evidence.png"
)

for f in "${required[@]}"; do
  if [[ ! -f "${SRC_DIR}/${f}" ]]; then
    echo "missing latest screenshot slot: ${SRC_DIR}/${f}" >&2
    exit 1
  fi
done

mkdir -p "${DST_DIR}"
for f in "${required[@]}"; do
  cp "${SRC_DIR}/${f}" "${DST_DIR}/${f}"
done

echo "synced screenshots to: ${DST_DIR}"

#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Check required screenshot slots for latest and a target release version.

Usage:
  scripts/screenshot-check.sh --version <version>

Example:
  scripts/screenshot-check.sh --version v0.4.2-preview.1
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
LATEST_DIR="${ROOT_DIR}/docs/assets/screenshots/latest"
RELEASE_DIR="${ROOT_DIR}/docs/assets/screenshots/releases/${VERSION}"

required=(
  "ui-template-stage.png"
  "ui-dashboard-evidence.png"
)

missing=0
for base in "${LATEST_DIR}" "${RELEASE_DIR}"; do
  for f in "${required[@]}"; do
    path="${base}/${f}"
    if [[ ! -f "${path}" ]]; then
      echo "missing screenshot: ${path}" >&2
      missing=1
    fi
  done
done

if [[ "${missing}" -ne 0 ]]; then
  exit 1
fi

echo "screenshot check passed for ${VERSION}"

#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Build OpsKit Linux release binaries and checksums.

Usage:
  scripts/release.sh [options]

Options:
  -v, --version <ver>   Optional version suffix in file names
  -o, --output <dir>    Output directory (default: ./dist)
      --name <name>     Binary base name (default: opskit)
      --clean           Remove output directory before building
  -h, --help            Show help

Environment:
  GO_CACHE_DIR          Go build cache dir (default: ./.tmp/gocache-release)

Examples:
  scripts/release.sh
  scripts/release.sh --version v0.3.0-preview.1
  scripts/release.sh --output ./.tmp/release-dist --clean
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/dist"
BINARY_NAME="opskit"
VERSION=""
CLEAN=0
GO_CACHE_DIR="${GO_CACHE_DIR:-${ROOT_DIR}/.tmp/gocache-release}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -v|--version)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      VERSION="$2"
      shift 2
      ;;
    -o|--output)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --name)
      [[ $# -ge 2 ]] || { echo "missing value for $1" >&2; exit 2; }
      BINARY_NAME="$2"
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

if ! command -v go >/dev/null 2>&1; then
  echo "go is required but not found in PATH" >&2
  exit 2
fi

if [[ "${CLEAN}" == "1" ]]; then
  rm -rf "${OUTPUT_DIR}"
fi
mkdir -p "${OUTPUT_DIR}"
mkdir -p "${GO_CACHE_DIR}"

build_target() {
  local arch="$1"
  local file="${BINARY_NAME}-linux-${arch}"
  if [[ -n "${VERSION}" ]]; then
    file="${BINARY_NAME}-${VERSION}-linux-${arch}"
  fi
  local out="${OUTPUT_DIR}/${file}"
  echo "==> building ${out}"
  (
    cd "${ROOT_DIR}"
    GOCACHE="${GO_CACHE_DIR}" CGO_ENABLED=0 GOOS=linux GOARCH="${arch}" go build -trimpath -o "${out}" ./cmd/opskit
  )
}

build_target "arm64"
build_target "amd64"

if command -v sha256sum >/dev/null 2>&1; then
  echo "==> generating checksums.txt with sha256sum"
  (
    cd "${OUTPUT_DIR}"
    sha256sum ./*linux-arm64 ./*linux-amd64 > checksums.txt
  )
elif command -v shasum >/dev/null 2>&1; then
  echo "==> generating checksums.txt with shasum -a 256"
  (
    cd "${OUTPUT_DIR}"
    shasum -a 256 ./*linux-arm64 ./*linux-amd64 > checksums.txt
  )
else
  echo "warning: no sha256sum/shasum found, skipping checksums.txt" >&2
fi

echo ""
echo "Release artifacts are ready in: ${OUTPUT_DIR}"
ls -1 "${OUTPUT_DIR}"

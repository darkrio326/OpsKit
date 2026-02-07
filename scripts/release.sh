#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Build OpsKit Linux release binaries, metadata, and checksums.

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
CHECKSUM_TOOL=""
CHECKSUM_CMD=""
CHECKSUM_VERIFY_CMD=""
BUILD_TIME_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
GIT_COMMIT="unknown"
GO_VERSION="unknown"
BUILT_FILES=()

json_escape() {
  local s="$1"
  s=${s//\\/\\\\}
  s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}
  s=${s//$'\r'/\\r}
  s=${s//$'\t'/\\t}
  printf '%s' "${s}"
}

detect_checksum_tool() {
  if command -v sha256sum >/dev/null 2>&1; then
    CHECKSUM_TOOL="sha256sum"
    CHECKSUM_CMD="sha256sum"
    CHECKSUM_VERIFY_CMD="sha256sum -c checksums.txt"
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    CHECKSUM_TOOL="shasum"
    CHECKSUM_CMD="shasum -a 256"
    CHECKSUM_VERIFY_CMD="shasum -a 256 -c checksums.txt"
    return
  fi
  echo "sha256sum or shasum is required to generate checksums.txt" >&2
  exit 2
}

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
  BUILT_FILES+=("${file}")
}

detect_checksum_tool

if git -C "${ROOT_DIR}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  GIT_COMMIT="$(git -C "${ROOT_DIR}" rev-parse --short HEAD)"
fi
GO_VERSION="$(go env GOVERSION 2>/dev/null || true)"
if [[ -z "${GO_VERSION}" ]]; then
  GO_VERSION="unknown"
fi

build_target "arm64"
build_target "amd64"

echo "==> generating checksums.txt with ${CHECKSUM_TOOL}"
(
  cd "${OUTPUT_DIR}"
  ${CHECKSUM_CMD} "${BUILT_FILES[@]}" > checksums.txt
)

echo "==> verifying checksums.txt"
(
  cd "${OUTPUT_DIR}"
  ${CHECKSUM_VERIFY_CMD}
)

echo "==> writing release-metadata.json"
{
  echo "{"
  echo "  \"schemaVersion\": \"v1\","
  echo "  \"generatedAt\": \"$(json_escape "${BUILD_TIME_UTC}")\","
  if [[ -n "${VERSION}" ]]; then
    echo "  \"version\": \"$(json_escape "${VERSION}")\","
  else
    echo "  \"version\": null,"
  fi
  echo "  \"binaryBaseName\": \"$(json_escape "${BINARY_NAME}")\","
  echo "  \"gitCommit\": \"$(json_escape "${GIT_COMMIT}")\","
  echo "  \"goVersion\": \"$(json_escape "${GO_VERSION}")\","
  echo "  \"checksumTool\": \"$(json_escape "${CHECKSUM_TOOL}")\","
  echo "  \"checksumsFile\": \"checksums.txt\","
  echo "  \"artifacts\": ["

  for i in "${!BUILT_FILES[@]}"; do
    f="${BUILT_FILES[$i]}"
    sep=","
    if [[ "${i}" -eq $((${#BUILT_FILES[@]} - 1)) ]]; then
      sep=""
    fi
    echo "    {\"name\": \"$(json_escape "${f}")\"}${sep}"
  done

  echo "  ]"
  echo "}"
} > "${OUTPUT_DIR}/release-metadata.json"

echo ""
echo "Release artifacts are ready in: ${OUTPUT_DIR}"
ls -1 "${OUTPUT_DIR}"

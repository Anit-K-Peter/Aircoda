#!/usr/bin/env bash
set -euo pipefail

GO_CMD="${GO:-go}"
if ! command -v "${GO_CMD}" &>/dev/null; then
    if [ -x "/home/anitkp/.local/go/bin/go" ]; then
        GO_CMD="/home/anitkp/.local/go/bin/go"
    fi
fi

VERSION="${VERSION:-v0.1.0}"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "dev")"
BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

LDFLAGS="-s -w -X aircoda/internal/system.Version=${VERSION} -X aircoda/internal/system.Commit=${COMMIT} -X aircoda/internal/system.BuildDate=${BUILD_DATE}"

echo "Building Aircoda ${VERSION} (commit: ${COMMIT})..."

TARGETS=(
    "linux/amd64/radio"
    "linux/arm64/radio"
    "windows/amd64/radio.exe"
    "windows/arm64/radio.exe"
    "darwin/amd64/radio"
    "darwin/arm64/radio"
)

rm -rf dist
mkdir -p dist

for target in "${TARGETS[@]}"; do
    IFS="/" read -r os arch binary <<< "$target"
    output_dir="dist/aircoda_${VERSION}_${os}_${arch}"
    mkdir -p "${output_dir}"
    
    echo "  -> Compiling ${os}/${arch}..."
    GOOS="${os}" GOARCH="${arch}" CGO_ENABLED=0 "${GO_CMD}" build -ldflags "${LDFLAGS}" -o "${output_dir}/${binary}" ./cmd/radio
done

echo "Build complete. Artifacts written to dist/"

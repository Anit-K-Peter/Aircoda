#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${ROOT_DIR}"

./scripts/build.sh

echo "Packaging release archives and generating checksums..."

cd dist

rm -f *.tar.gz *.zip SHA256SUMS

for dir in aircoda_*; do
    if [ -d "${dir}" ]; then
        if [[ "${dir}" == *"windows"* ]]; then
            echo "  -> Zipping ${dir}.zip"
            zip -rq "${dir}.zip" "${dir}"
        else
            echo "  -> Archiving ${dir}.tar.gz"
            tar -czf "${dir}.tar.gz" "${dir}"
        fi
    fi
done

echo "Generating SHA256SUMS..."
sha256sum *.tar.gz *.zip > SHA256SUMS

echo "Release packaging complete:"
cat SHA256SUMS

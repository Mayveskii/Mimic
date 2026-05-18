#!/usr/bin/env bash
#
# package-data.sh — Package mesh data for GitHub Release
#
# Usage: ./scripts/package-data.sh [version]
# Generates: dist/mimic-data-VERSION.tar.gz + .sha256

set -euo pipefail

VERSION="${1:-$(git describe --tags --always 2>/dev/null || echo 'dev')}"
DIST_DIR="dist"
DATA_DIR="data"

mkdir -p "$DIST_DIR"

echo "📦 Packaging Mimic data bundle v${VERSION}..."

# Verify data exists
if [[ ! -d "${DATA_DIR}/seeds" ]]; then
    echo "❌ Error: ${DATA_DIR}/seeds/ not found. Run 'python3 data/extraction/distill_pipeline.py' first."
    exit 1
fi

# Create tarball with only necessary data
echo "   Adding seeds/..."
echo "   Adding distilled/..."
echo "   Adding registry.json..."

tar -czf "${DIST_DIR}/mimic-data-${VERSION}.tar.gz" \
    -C "$DATA_DIR" \
    seeds/ \
    distilled/ \
    registry.json

# Compute SHA256
cd "$DIST_DIR"
sha256sum "mimic-data-${VERSION}.tar.gz" > "mimic-data-${VERSION}.tar.gz.sha256"
cd ..

SIZE=$(du -h "${DIST_DIR}/mimic-data-${VERSION}.tar.gz" | cut -f1)

echo ""
echo "✅ Data bundle ready:"
echo "   File:  ${DIST_DIR}/mimic-data-${VERSION}.tar.gz"
echo "   Size:  $SIZE"
echo "   SHA256: $(cat ${DIST_DIR}/mimic-data-${VERSION}.tar.gz.sha256 | cut -d' ' -f1)"
echo ""
echo "Upload to GitHub Release:"
echo "   gh release upload v${VERSION} ${DIST_DIR}/mimic-data-${VERSION}.tar.gz ${DIST_DIR}/mimic-data-${VERSION}.tar.gz.sha256"

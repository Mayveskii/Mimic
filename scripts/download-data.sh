#!/usr/bin/env bash
#
# download-data.sh — Download mesh data from GitHub Release on first run
#
# Called by: mimic serve (if data not present), make bootstrap, or install.sh
#
# Cache location:
#   Linux:   ~/.cache/mimic/data/
#   macOS:   ~/Library/Caches/mimic/data/
#   Windows: %LOCALAPPDATA%\mimic\data\

set -euo pipefail

REPO="Mayveskii/Mimic"
DEFAULT_VERSION="latest"
FORCE="${MIMIC_FORCE_DOWNLOAD:-0}"

# Detect OS and set cache dir
detect_cache_dir() {
    case "$(uname -s)" in
        Linux*)   echo "${HOME}/.cache/mimic/data" ;;
        Darwin*)  echo "${HOME}/Library/Caches/mimic/data" ;;
        MINGW*|MSYS*|CYGWIN*)
            local appdata="${LOCALAPPDATA:-${HOME}/AppData/Local}"
            echo "$appdata/mimic/data"
            ;;
        *)        echo "${HOME}/.cache/mimic/data" ;;
    esac
}

CACHE_DIR="$(detect_cache_dir)"
REGISTRY_FILE="$CACHE_DIR/registry.json"

download_registry() {
    local version="$1"
    local url
    
    if [[ "$version" == "latest" ]]; then
        # Get actual latest version from GitHub API
        version="$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | head -n1 | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')"
        if [[ -z "$version" || "$version" == "null" ]]; then
            echo "❌ Could not determine latest version. Check network or set MIMIC_DATA_VERSION."
            exit 1
        fi
    fi
    
    echo "📥 Downloading data registry for $version..."
    
    # Download registry.json from release
    url="https://github.com/${REPO}/releases/download/${version}/mimic-data-${version}.tar.gz"
    local registry_url="https://raw.githubusercontent.com/${REPO}/${version}/data/registry.json"
    
    mkdir -p "$CACHE_DIR"
    curl -sSL "$registry_url" -o "$REGISTRY_FILE" || {
        echo "❌ Failed to download registry"
        exit 1
    }
    
    echo "$version"
}

download_bundle() {
    local bundle_name="$1"
    local version="$2"
    
    local bundle_file="${CACHE_DIR}/mimic-data-${version}.tar.gz"
    local sha256_file="${bundle_file}.sha256"
    
    if [[ -f "$bundle_file" && "$FORCE" != "1" ]]; then
        echo "✅ Data bundle already exists: $bundle_file"
        return 0
    fi
    
    local url="https://github.com/${REPO}/releases/download/${version}/mimic-data-${version}.tar.gz"
    local sha256_url="https://github.com/${REPO}/releases/download/${version}/mimic-data-${version}.tar.gz.sha256"
    
    echo "📥 Downloading data bundle ($version)..."
    curl -sSL -o "$bundle_file" "$url" || {
        echo "❌ Failed to download data bundle"
        exit 1
    }
    
    echo "📥 Downloading checksum..."
    if curl -sSL -o "$sha256_file" "$sha256_url" 2>/dev/null; then
        echo "🔐 Verifying checksum..."
        if command -v sha256sum >/dev/null 2>&1; then
            if ! sha256sum --check "$sha256_file" >/dev/null 2>&1; then
                echo "❌ Checksum mismatch! Deleting corrupted bundle."
                rm -f "$bundle_file"
                exit 1
            fi
        elif command -v shasum >/dev/null 2>&1; then
            if ! shasum -a 256 --check "$sha256_file" >/dev/null 2>&1; then
                echo "❌ Checksum mismatch! Deleting corrupted bundle."
                rm -f "$bundle_file"
                exit 1
            fi
        fi
        echo "✅ Checksum verified"
    else
        echo "⚠️  Checksum file not available, skipping verification"
    fi
    
    echo "📦 Extracting..."
    tar -xzf "$bundle_file" -C "$CACHE_DIR" --strip-components=0
    
    echo "✅ Data ready: $CACHE_DIR"
}

main() {
    local version="${1:-${MIMIC_DATA_VERSION:-latest}}"
    
    echo "═══════════════════════════════════════════════════════════════"
    echo "  Mimic Mesh Data Download"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    # Check if data already exists
    if [[ -d "$CACHE_DIR/seeds" && "$FORCE" != "1" ]]; then
        echo "✅ Mesh data already present: $CACHE_DIR"
        echo "   Slots: $(ls "$CACHE_DIR/seeds"/*-artifacts.json 2>/dev/null | wc -l) files"
        echo ""
        echo "   To force re-download: MIMIC_FORCE_DOWNLOAD=1 $0"
        exit 0
    fi
    
    echo "Cache directory: $CACHE_DIR"
    echo "Target version:  $version"
    echo ""
    
    # Download registry to get version info
    local actual_version
    actual_version="$(download_registry "$version")"
    
    # Download bundle
    download_bundle "mesh-core" "$actual_version"
    
    # Verify
    local slot_count
    slot_count=$(find "$CACHE_DIR/seeds" -name '*-artifacts.json' 2>/dev/null | wc -l)
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  ✅ Mesh data ready!"
    echo "  Version:    $actual_version"
    echo "  Cache dir:  $CACHE_DIR"
    echo "  Seed files: $slot_count"
    echo "═══════════════════════════════════════════════════════════════"
}

main "$@"

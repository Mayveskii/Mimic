#!/usr/bin/env bash
#
# Mimic MCP Server — One-liner installer
#
# Install:
#   curl -sSL https://raw.githubusercontent.com/Mayveskii/Mimic/main/install.sh | bash
#
# Or with options:
#   curl -sSL ... | bash -s -- --tag v0.1.0 --install-dir /usr/local/bin
#
# Supports: Linux (amd64, arm64), macOS (amd64, arm64)

set -euo pipefail

REPO="Mayveskii/Mimic"
DEFAULT_TAG="latest"
DEFAULT_INSTALL_DIR="/usr/local/bin"
TMP_DIR="$(mktemp -d)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging
log_info() { printf "${BLUE}[INFO]${NC} %s\n" "$1"; }
log_ok()   { printf "${GREEN}[OK]${NC}   %s\n" "$1"; }
log_warn() { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
log_err()  { printf "${RED}[ERR]${NC}  %s\n" "$1" >&2; }

# Detect architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64)  echo "amd64" ;;
        aarch64) echo "arm64" ;;
        arm64)   echo "arm64" ;;
        *)
            log_err "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# Detect OS
detect_os() {
    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux)  echo "linux" ;;
        darwin) echo "darwin" ;;
        *)
            log_err "Unsupported OS: $os"
            exit 1
            ;;
    esac
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksum_file="$2"

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum --check "$checksum_file" >/dev/null 2>&1
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 --check "$checksum_file" >/dev/null 2>&1
    else
        log_warn "sha256sum not found, skipping checksum verification"
        return 0
    fi
}

# Download from GitHub releases
download_binary() {
    local tag="$1"
    local os="$2"
    local arch="$3"

    local binary_name="mimic_${tag}_${os}_${arch}"
    local tar_name="${binary_name}.tar.gz"
    local checksum_name="${tar_name}.sha256"

    local base_url="https://github.com/${REPO}/releases/download/${tag}"

    log_info "Downloading ${tar_name}..."
    curl -sSL -o "${TMP_DIR}/${tar_name}" "${base_url}/${tar_name}" || {
        log_err "Failed to download binary"
        exit 1
    }

    log_info "Downloading checksum..."
    if curl -sSL -o "${TMP_DIR}/${checksum_name}" "${base_url}/${checksum_name}" 2>/dev/null; then
        log_info "Verifying checksum..."
        if verify_checksum "${TMP_DIR}/${tar_name}" "${TMP_DIR}/${checksum_name}"; then
            log_ok "Checksum verified"
        else
            log_err "Checksum mismatch!"
            exit 1
        fi
    else
        log_warn "Checksum file not found, skipping verification"
    fi

    echo "${TMP_DIR}/${tar_name}"
}

# Build from source if binary not available
build_from_source() {
    log_info "Binary not available for your platform, building from source..."

    if ! command -v go >/dev/null 2>&1; then
        log_err "Go not found. Please install Go 1.22+ first:"
        log_err "  https://go.dev/dl/"
        exit 1
    fi

    local go_version
    go_version="$(go version | awk '{print $3}' | sed 's/go//')"
    log_info "Go version: $go_version"

    if ! command -v gcc >/dev/null 2>&1; then
        log_err "gcc not found. Please install build tools:"
        log_err "  Debian/Ubuntu: sudo apt-get install gcc make libssl-dev"
        log_err "  macOS: xcode-select --install"
        exit 1
    fi

    log_info "Cloning repository..."
    git clone --depth 1 "https://github.com/${REPO}.git" "${TMP_DIR}/mimic-src"

    cd "${TMP_DIR}/mimic-src"

    log_info "Building (this may take a minute)..."
    make clean && CGO_ENABLED=1 make

    if [[ ! -f bin/mimic ]]; then
        log_err "Build failed: bin/mimic not found"
        exit 1
    fi

    echo "${TMP_DIR}/mimic-src/bin/mimic"
}

# Install Docker version
install_docker() {
    log_info "Installing via Docker..."
    if ! command -v docker >/dev/null 2>&1; then
        log_err "Docker not found. Please install Docker first:"
        log_err "  https://docs.docker.com/get-docker/"
        exit 1
    fi

    docker pull "ghcr.io/${REPO}:latest"
    log_ok "Docker image pulled"

    cat <<'EOF'

# Run Mimic via Docker: 
docker run -it --rm \
  -p 1337:1337 \
  -e MIMIC_PORT=1337 \
  ghcr.io/Mayveskii/Mimic:latest serve

EOF
}

# Main install
main() {
    local tag="${1:-$DEFAULT_TAG}"
    local install_dir="${2:-$DEFAULT_INSTALL_DIR}"
    local method="${3:-auto}"  # auto, binary, source, docker

    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║               Mimic MCP Server Installer                     ║"
    echo "║      Deterministic AI-agent tool orchestration               ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""

    local os arch
    os="$(detect_os)"
    arch="$(detect_arch)"
    log_info "Detected: $os/$arch"
    log_info "Install directory: $install_dir"

    # Check if already installed
    if command -v mimic >/dev/null 2>&1; then
        local current_version
        current_version="$(mimic version 2>/dev/null || echo 'unknown')"
        log_warn "Mimic already installed: $current_version"
        read -rp "Reinstall? [y/N] " answer
        if [[ ! "$answer" =~ ^[Yy]$ ]]; then
            log_info "Cancelled."
            exit 0
        fi
    fi

    # Method dispatch
    local binary_path

    case "$method" in
        docker)
            install_docker
            exit 0
            ;;
        source)
            binary_path="$(build_from_source)"
            ;;
        auto|binary)
            if [[ "$tag" == "latest" ]]; then
                # Try to get latest tag from GitHub API
                tag="$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | head -n1 | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')"
                tag="${tag:-v0.1.0}"
            fi

            log_info "Installing version: $tag"

            # Try binary download first
            local tar_path
            tar_path="$(download_binary "$tag" "$os" "$arch" 2>/dev/null || true)"

            if [[ -n "$tar_path" && -f "$tar_path" ]]; then
                log_info "Extracting..."
                tar -xzf "$tar_path" -C "$TMP_DIR"
                binary_path="${TMP_DIR}/mimic"
            else
                log_warn "Binary not available, falling back to source build"
                binary_path="$(build_from_source)"
            fi
            ;;
        *)
            log_err "Unknown method: $method"
            exit 1
            ;;
    esac

    # Install binary
    log_info "Installing mimic to $install_dir..."
    if [[ ! -d "$install_dir" ]]; then
        log_info "Creating directory $install_dir..."
        mkdir -p "$install_dir" 2>/dev/null || {
            log_err "Cannot create $install_dir. Try with sudo:"
            log_err "  sudo bash -c '$(curl -sSL ...)'"
            exit 1
        }
    fi

    # Check permissions
    if [[ ! -w "$install_dir" ]]; then
        log_warn "Need write permission to $install_dir"
        if command -v sudo >/dev/null 2>&1; then
            sudo cp "$binary_path" "${install_dir}/mimic"
            sudo chmod +x "${install_dir}/mimic"
        else
            log_err "Cannot write to $install_dir and sudo not available"
            exit 1
        fi
    else
        cp "$binary_path" "${install_dir}/mimic"
        chmod +x "${install_dir}/mimic"
    fi

    # Verify installation
    if command -v mimic >/dev/null 2>&1; then
        local version
        version="$(mimic version 2>/dev/null || echo 'installed')"
        log_ok "Mimic installed successfully: $version"
    else
        # Add to PATH if not found
        if [[ ":$PATH:" != *":$install_dir:"* ]]; then
            log_warn "$install_dir not in PATH"
            log_info "Add to your shell profile:"
            log_info "  export PATH=\"\$PATH:$install_dir\""
        fi
        log_ok "Mimic installed to ${install_dir}/mimic"
    fi

    # Download mesh data
    log_info "Downloading mesh data (first run)..."
    if [[ -f scripts/download-data.sh ]]; then
        bash scripts/download-data.sh "$TAG"
    elif command -v mimic >/dev/null 2>&1; then
        # If installed binary, use embedded download
        mimic data bootstrap 2>/dev/null || \
            log_warn "Could not auto-download data. Run: mimic data bootstrap"
    fi

    # Quick start
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  QUICK START"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    echo "  # Start MCP server (stdio transport)"
    echo "  mimic serve"
    echo ""
    echo "  # Start HTTP server on port 1337"
    echo "  mimic serve --port 1337"
    echo ""
    echo "  # Check version"
    echo "  mimic version"
    echo ""
    echo "  # Run integration tests"
    echo "  mimic test"
    echo ""
    echo "  # Full documentation:"
    echo "  https://github.com/${REPO}"
    echo ""
    echo "═══════════════════════════════════════════════════════════════"

    # Cleanup
    rm -rf "$TMP_DIR"
}

# CLI args
TAG="${MIMIC_TAG:-$DEFAULT_TAG}"
INSTALL_DIR="${MIMIC_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
METHOD="auto"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --tag|-t)
            TAG="$2"; shift 2 ;;
        --install-dir|-d)
            INSTALL_DIR="$2"; shift 2 ;;
        --source)
            METHOD="source"; shift ;;
        --docker)
            METHOD="docker"; shift ;;
        --help|-h)
            cat <<'EOF'
Usage: install.sh [OPTIONS]

Options:
  -t, --tag TAG          Version tag (default: latest)
  -d, --install-dir DIR  Installation directory (default: /usr/local/bin)
      --source           Build from source instead of downloading binary
      --docker           Install Docker image instead of binary
  -h, --help             Show this help

Environment:
  MIMIC_TAG              Default tag
  MIMIC_INSTALL_DIR      Default install directory

Examples:
  # Default install
  curl -sSL ... | bash

  # Specific version
  curl -sSL ... | bash -s -- --tag v0.1.0

  # Build from source
  curl -sSL ... | bash -s -- --source

  # Docker install
  curl -sSL ... | bash -s -- --docker

EOF
            exit 0
            ;;
        *)
            log_err "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run
cd "$(dirname "$0")" || true
main "$TAG" "$INSTALL_DIR" "$METHOD"

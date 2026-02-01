#!/bin/bash
set -e

# Trabuco Installer
# Usage: curl -sSL https://github.com/your-org/trabuco/releases/latest/download/install.sh | bash

REPO="trabuco/trabuco"
BINARY_NAME="trabuco"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info() { echo -e "${CYAN}$1${NC}"; }
success() { echo -e "${GREEN}$1${NC}"; }
warn() { echo -e "${YELLOW}$1${NC}"; }
error() { echo -e "${RED}$1${NC}"; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest version from GitHub
get_latest_version() {
    curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/'
}

# Main installation
main() {
    echo ""
    info "╔════════════════════════════════════════╗"
    info "║      Trabuco Installer                 ║"
    info "║      Java Project Generator CLI        ║"
    info "╚════════════════════════════════════════╝"
    echo ""

    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected platform: ${OS}-${ARCH}"

    # Get version
    VERSION=${TRABUCO_VERSION:-$(get_latest_version)}
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Set TRABUCO_VERSION manually."
    fi
    info "Installing version: ${VERSION}"

    # Build download URL
    if [ "$OS" = "windows" ]; then
        FILENAME="${BINARY_NAME}-${OS}-${ARCH}.exe"
    else
        FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
    fi
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf ${TMP_DIR}" EXIT

    # Download binary
    info "Downloading ${DOWNLOAD_URL}..."
    if ! curl -sSL -o "${TMP_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"; then
        error "Failed to download binary. Check if release exists."
    fi

    # Make executable
    chmod +x "${TMP_DIR}/${BINARY_NAME}"

    # Determine install location
    INSTALL_DIR=""
    NEEDS_PATH_UPDATE=false

    if [ -w /usr/local/bin ]; then
        INSTALL_DIR="/usr/local/bin"
    elif [ -d /usr/local/bin ] && command -v sudo &> /dev/null; then
        info "Requesting sudo access to install to /usr/local/bin..."
        if sudo -v 2>/dev/null; then
            INSTALL_DIR="/usr/local/bin"
            USE_SUDO=true
        fi
    fi

    # Fall back to user directory
    if [ -z "$INSTALL_DIR" ]; then
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "$INSTALL_DIR"

        # Check if directory is in PATH
        if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
            NEEDS_PATH_UPDATE=true
        fi
    fi

    info "Installing to ${INSTALL_DIR}..."

    # Install binary
    if [ "$USE_SUDO" = true ]; then
        sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # Handle PATH update if needed
    if [ "$NEEDS_PATH_UPDATE" = true ]; then
        warn ""
        warn "~/.local/bin is not in your PATH."
        warn ""

        # Detect shell and config file
        SHELL_CONFIG=""
        case "$SHELL" in
            */zsh)  SHELL_CONFIG="$HOME/.zshrc" ;;
            */bash)
                if [ -f "$HOME/.bash_profile" ]; then
                    SHELL_CONFIG="$HOME/.bash_profile"
                else
                    SHELL_CONFIG="$HOME/.bashrc"
                fi
                ;;
            */fish) SHELL_CONFIG="$HOME/.config/fish/config.fish" ;;
        esac

        if [ -n "$SHELL_CONFIG" ]; then
            read -p "Add to PATH in ${SHELL_CONFIG}? [Y/n] " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Nn]$ ]]; then
                if [[ "$SHELL" == */fish ]]; then
                    echo "set -gx PATH \$HOME/.local/bin \$PATH" >> "$SHELL_CONFIG"
                else
                    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_CONFIG"
                fi
                success "Added to ${SHELL_CONFIG}"
                warn "Run 'source ${SHELL_CONFIG}' or restart your terminal."
            else
                warn "Skipped. Add manually: export PATH=\"\$HOME/.local/bin:\$PATH\""
            fi
        else
            warn "Add this to your shell config:"
            warn "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        fi
    fi

    echo ""
    success "╔════════════════════════════════════════╗"
    success "║  Trabuco installed successfully!       ║"
    success "╚════════════════════════════════════════╝"
    echo ""

    # Verify installation
    if command -v trabuco &> /dev/null; then
        info "Installed version:"
        trabuco version
    else
        info "Installed to: ${INSTALL_DIR}/${BINARY_NAME}"
        info "Run: ${INSTALL_DIR}/${BINARY_NAME} version"
    fi

    echo ""
    info "Get started:"
    info "  trabuco init"
    echo ""
}

main "$@"

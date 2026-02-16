#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

REPO="ipanardian/lu-hut"
LATEST_RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}
detect_os_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $OS in
        linux)
            OS_TITLE="Linux"
            ;;
        darwin)
            OS_TITLE="Darwin"
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    case $ARCH in
        x86_64)
            ARCH_NAME="x86_64"
            ;;
        arm64|aarch64)
            ARCH_NAME="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    print_info "Detected OS: $OS_TITLE, Architecture: $ARCH_NAME"
}

get_latest_version() {
    VERSION=$(curl -s "$LATEST_RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"v?([^"]*).*/\1/')
    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi
    print_info "Latest version: v$VERSION"
}

download_binary() {
    ARCHIVE_NAME="lu-hut_${VERSION}_${OS_TITLE}_${ARCH_NAME}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE_NAME}"

    print_info "Downloading lu from: $DOWNLOAD_URL"

    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    if ! curl -L -o "$TMP_DIR/${ARCHIVE_NAME}" "$DOWNLOAD_URL"; then
        print_error "Failed to download archive"
        exit 1
    fi

    print_info "Extracting archive..."
    if ! tar -xzf "$TMP_DIR/${ARCHIVE_NAME}" -C "$TMP_DIR"; then
        print_error "Failed to extract archive"
        exit 1
    fi

    if [ ! -f "$TMP_DIR/lu" ]; then
        print_error "Binary not found in archive"
        exit 1
    fi

    chmod +x "$TMP_DIR/lu"

    INSTALL_DIR="$HOME/.local/bin"
    if [ ! -d "$INSTALL_DIR" ]; then
        INSTALL_DIR="/usr/local/bin"
        if [ ! -w "$INSTALL_DIR" ]; then
            print_error "Cannot write to $INSTALL_DIR. Try running with sudo."
            exit 1
        fi
    fi

    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        sudo cp "$TMP_DIR/lu" "$INSTALL_DIR/"
    else
        mkdir -p "$INSTALL_DIR"
        cp "$TMP_DIR/lu" "$INSTALL_DIR/"
    fi

    print_info "lu installed successfully to $INSTALL_DIR/lu"
}

check_path() {
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        print_warning "$INSTALL_DIR is not in your PATH"
        echo "Add the following to your shell profile:"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
}

main() {
    echo "lu-hut Installation Script"
    echo "======================="
    echo

    detect_os_arch
    get_latest_version
    download_binary
    check_path

    echo
    print_info "Installation complete! Run 'lu --help' to get started."
}

main "$@"

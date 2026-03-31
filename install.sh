#!/bin/bash

set -e

REPO="vcvcvnvcvcvn/one-liner"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="ol"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print functions
print_info() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1" >&2
}

print_error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1" >&2
}

print_warning() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1" >&2
}

# Detect OS and architecture
detect_platform() {
    local os arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac

    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        i386|i686)
            ARCH="386"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Get latest release version
get_latest_version() {
    print_info "Fetching latest release..."

    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        print_error "Could not fetch latest version. Please check your internet connection."
        exit 1
    fi

    # Remove 'v' prefix if present for filename matching
    VERSION_NO_V="${VERSION#v}"

    print_info "Latest version: $VERSION"
}

# Download binary
download_binary() {
    local archive_name temp_dir archive_file temp_file
    local extracted_binary_name="ol-${PLATFORM}"
    archive_name="ol-${VERSION_NO_V}-${PLATFORM}.tar.gz"

    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${archive_name}"

    local temp_dir
    temp_dir=$(mktemp -d)
    local archive_file="${temp_dir}/${archive_name}"

    print_info "Downloading ${archive_name}..."
    print_info "URL: $download_url"

    if command -v curl >/dev/null 2>&1; then
        curl -sL "$download_url" -o "$archive_file" || {
            print_error "Download failed"
            rm -rf "$temp_dir"
            exit 1
        }
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "$archive_file" || {
            print_error "Download failed"
            rm -rf "$temp_dir"
            exit 1
        }
    fi

    # Extract archive
    print_info "Extracting archive..."
    (cd "$temp_dir" && tar -xzf "$archive_file")

    # Find the extracted binary (it's named ol-${PLATFORM} in the archive)
    local temp_file="${temp_dir}/${extracted_binary_name}"

    # Make executable
    chmod +x "$temp_file"

    echo "$temp_file"
}

# Install binary
install_binary() {
    local temp_file="$1"

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        print_warning "Root permission required to install to ${INSTALL_DIR}"
        sudo mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # Cleanup temp dir
    rm -rf "$(dirname "$temp_file")"

    print_info "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}"
}

# Verify installation
verify_installation() {
    if command -v ol >/dev/null 2>&1; then
        local version
        version=$(ol --version)
        print_info "Installation verified: $version"
        return 0
    else
        print_error "Installation verification failed. Please check your PATH."
        return 1
    fi
}

# Add to PATH reminder
path_reminder() {
    if ! command -v ol >/dev/null 2>&1; then
        print_warning "${INSTALL_DIR} is not in your PATH"
        echo ""
        echo "Add the following line to your shell configuration file:"
        echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
        echo ""
        echo "For bash: ~/.bashrc or ~/.bash_profile"
        echo "For zsh: ~/.zshrc"
        echo "For fish: ~/.config/fish/config.fish"
    fi
}

# Main installation flow
main() {
    echo "========================================"
    echo "   ol - One-liner Command Assistant"
    echo "========================================"
    echo ""

    detect_platform
    print_info "Detected platform: $PLATFORM"

    get_latest_version

    local temp_file
    temp_file=$(download_binary)
    install_binary "$temp_file"

    verify_installation
    path_reminder

    echo ""
    echo "========================================"
    print_info "Installation complete!"
    echo ""
    echo "Get started with:"
    echo "  ol --help       Show help"
    echo "  ol --init       Configure API"
    echo "  ol <request>    Generate command"
    echo "========================================"
}

# Run main function
main

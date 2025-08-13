#!/bin/bash
set -e

# Configuration
TOOL_NAME="maniplacer"
REPO_URL="https://github.com/dantedelordran/maniplacer-rebirth"
INSTALL_DIR="$HOME/.local/bin"
TEMP_DIR="/tmp/maniplacer-install"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Cleanup on exit
cleanup() {
    [ -d "$TEMP_DIR" ] && rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."
    
    if ! command_exists curl; then
        error "curl is required but not installed. Please install curl and try again."
    fi
    
    if ! command_exists grep; then
        error "grep is required but not installed."
    fi
    
    if ! command_exists sed; then
        error "sed is required but not installed."
    fi
    
    success "All prerequisites satisfied"
}

# Detect OS and architecture
detect_platform() {
    info "Detecting platform..."
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    # Map OS names
    case "$OS" in
        "linux") OS="linux" ;;
        "darwin") OS="darwin" ;;
        "mingw"*|"msys"*|"cygwin"*) OS="windows" ;;
        *) error "Unsupported OS: $OS" ;;
    esac
    
    # Map architecture naming
    case "$ARCH" in
        "x86_64"|"amd64") ARCH="amd64" ;;
        "arm64"|"aarch64") ARCH="arm64" ;;
        "i386"|"i686") ARCH="386" ;;
        *) warning "Unknown architecture: $ARCH, defaulting to amd64"; ARCH="amd64" ;;
    esac
    
    # Set binary name with extension for Windows
    BINARY_NAME="maniplacer-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    
    info "Platform detected: $OS/$ARCH"
}

# Get latest release version
get_latest_version() {
    info "Fetching latest version from GitHub..."
    
    # Try to get latest version with better error handling
    LATEST_VERSION=$(curl -s --fail "https://api.github.com/repos/dantedelordran/maniplacer-rebirth/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/' | \
        tr -d '\n\r')
    
    if [ -z "$LATEST_VERSION" ]; then
        error "Failed to fetch latest version from GitHub API"
    fi
    
    info "Latest version: $LATEST_VERSION"
}

# Check if already installed and get version
check_existing_installation() {
    if command_exists "$TOOL_NAME"; then
        CURRENT_VERSION=$($TOOL_NAME --version 2>/dev/null || echo "unknown")
        warning "$TOOL_NAME is already installed (version: $CURRENT_VERSION)"
        
        # Remove 'v' prefix for comparison if present
        CLEAN_LATEST=${LATEST_VERSION#v}
        CLEAN_CURRENT=${CURRENT_VERSION#v}
        
        if [ "$CLEAN_CURRENT" = "$CLEAN_LATEST" ]; then
            info "You already have the latest version installed."
            read -p "Do you want to reinstall? (y/N): " -r
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                info "Installation cancelled."
                exit 0
            fi
        fi
    fi
}

# Download binary
download_binary() {
    info "Creating temporary directory..."
    mkdir -p "$TEMP_DIR"
    
    # Construct download URL
    DOWNLOAD_URL="$REPO_URL/releases/download/$LATEST_VERSION/$BINARY_NAME"
    
    info "Downloading $TOOL_NAME $LATEST_VERSION ($OS/$ARCH)..."
    info "Download URL: $DOWNLOAD_URL"
    
    # Download with progress bar and better error handling
    if ! curl -L --fail --progress-bar "$DOWNLOAD_URL" -o "$TEMP_DIR/$TOOL_NAME"; then
        error "Download failed. Please check if the release exists for your platform."
    fi
    
    # Verify download
    if [ ! -f "$TEMP_DIR/$TOOL_NAME" ]; then
        error "Downloaded file not found"
    fi
    
    # Check if file is actually downloaded (not empty)
    if [ ! -s "$TEMP_DIR/$TOOL_NAME" ]; then
        error "Downloaded file is empty"
    fi
    
    success "Download completed"
}

# Install binary
install_binary() {
    info "Installing $TOOL_NAME..."
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Make executable
    chmod +x "$TEMP_DIR/$TOOL_NAME"
    
    # Move to install directory
    mv "$TEMP_DIR/$TOOL_NAME" "$INSTALL_DIR/"
    
    success "Binary installed to $INSTALL_DIR/$TOOL_NAME"
}

# Update PATH
update_path() {
    # Check if already in PATH
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        info "$INSTALL_DIR is already in PATH"
        return
    fi
    
    info "Adding $INSTALL_DIR to PATH..."
    
    # Detect shell and update appropriate RC file
    SHELL_NAME=$(basename "$SHELL")
    RC_FILE=""
    
    case "$SHELL_NAME" in
        "bash")
            if [ -f "$HOME/.bashrc" ]; then
                RC_FILE="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                RC_FILE="$HOME/.bash_profile"
            fi
            ;;
        "zsh")
            RC_FILE="$HOME/.zshrc"
            ;;
        "fish")
            # Fish uses a different syntax
            if [ -d "$HOME/.config/fish" ]; then
                mkdir -p "$HOME/.config/fish/conf.d"
                echo "set -gx PATH $INSTALL_DIR \$PATH" > "$HOME/.config/fish/conf.d/maniplacer.fish"
                info "Added to Fish configuration"
                return
            fi
            ;;
    esac
    
    if [ -n "$RC_FILE" ]; then
        # Check if the export line already exists
        if ! grep -q "export PATH.*$INSTALL_DIR" "$RC_FILE" 2>/dev/null; then
            echo "" >> "$RC_FILE"
            echo "# Added by maniplacer installer" >> "$RC_FILE"
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$RC_FILE"
            info "Added to $RC_FILE"
        else
            info "PATH already configured in $RC_FILE"
        fi
    else
        warning "Could not determine shell RC file. Please manually add $INSTALL_DIR to your PATH."
    fi
}

# Verify installation
verify_installation() {
    info "Verifying installation..."
    
    # Add to current PATH for verification
    export PATH="$INSTALL_DIR:$PATH"
    
    if ! command_exists "$TOOL_NAME"; then
        error "Installation verification failed. $TOOL_NAME not found in PATH."
    fi
    
    # Test version command
    VERSION_OUTPUT=$($TOOL_NAME version 2>/dev/null || echo "")
    if [ -z "$VERSION_OUTPUT" ]; then
        warning "Could not verify version, but binary is installed"
    else
        info "Installed version: $VERSION_OUTPUT"
    fi
    
    success "Installation verified successfully!"
}

# Main installation process
main() {
    echo "🚀 $TOOL_NAME Installer"
    echo "========================"
    
    check_prerequisites
    detect_platform
    get_latest_version
    check_existing_installation
    download_binary
    install_binary
    update_path
    verify_installation
    
    echo ""
    success "Successfully installed $TOOL_NAME $LATEST_VERSION!"
    echo ""
    info "Next steps:"
    echo "  1. Restart your terminal or run: source ~/.bashrc (or ~/.zshrc)"
    echo "  2. Verify installation: $TOOL_NAME --version"
    echo "  3. Get started with: $TOOL_NAME --help"
    echo ""
    info "If you encounter any issues, please visit: $REPO_URL"
}

# Run main function
main "$@"
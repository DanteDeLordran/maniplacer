#!/bin/bash
set -e

TOOL_NAME="maniplacer"
REPO_URL="https://github.com/dantedelordran/maniplacer"
INSTALL_DIR="$HOME/.local/bin"  # User-local installation

# Create install directory if needed
mkdir -p "$INSTALL_DIR"

# Detect OS/Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture naming
case "$ARCH" in
  "x86_64") ARCH="amd64" ;;
  "arm64") ARCH="arm64" ;;
  *) ARCH="amd64" ;;  # Default fallback
esac

# Get latest release version
echo "🔍 Checking for latest version..."
LATEST_VERSION=$(curl -s "https://api.github.com/repos/dantedelordran/maniplacer/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_VERSION" ]; then
  echo "❌ Failed to fetch latest version"
  exit 1
fi

# Construct download URL
BINARY_URL="$REPO_URL/releases/download/$LATEST_VERSION/maniplacer-${OS}-${ARCH}"

echo "⬇️  Downloading $TOOL_NAME $LATEST_VERSION ($OS/$ARCH)..."
if ! curl -L "$BINARY_URL" -o "$TOOL_NAME"; then
  echo "❌ Download failed"
  exit 1
fi

# Make executable
chmod +x "$TOOL_NAME"

# Install binary
mv "$TOOL_NAME" "$INSTALL_DIR/"

# Add to PATH if not already present
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  for rc_file in ~/.bashrc ~/.zshrc; do
    [ -f "$rc_file" ] && echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$rc_file"
  done
  echo "↪️ Added $INSTALL_DIR to PATH"
fi

echo "✅ Successfully installed $TOOL_NAME $LATEST_VERSION!"
echo "   Restart your terminal or run: source ~/.bashrc"
echo "   Get started with: $TOOL_NAME --help"
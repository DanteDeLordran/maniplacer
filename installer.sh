#!/bin/bash
set -e

TOOL_NAME="maniplacer"
REPO_URL="https://github.com/dantedelordran/maniplacer"
VERSION="1.0.0"

# Detect OS/Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map to GitHub Release naming
case "$ARCH" in
  "x86_64") ARCH="amd64" ;;
  "arm64") ARCH="arm64" ;;
esac

BINARY_URL="$REPO_URL/releases/download/v$VERSION/maniplacer-$VERSION-$OS-$ARCH"

# Download binary
echo "⬇️  Downloading $TOOL_NAME..."
curl -L "$BINARY_URL" -o "$TOOL_NAME"
chmod +x "$TOOL_NAME"

# Install to ~/.local/bin (no sudo needed)
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"
mv "$TOOL_NAME" "$INSTALL_DIR"

# Add to PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> ~/.bashrc
  echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> ~/.zshrc
fi

echo "✅ Installed! Run this to start using:"
echo "   source ~/.bashrc && $TOOL_NAME --help"
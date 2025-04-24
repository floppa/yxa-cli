#!/bin/bash

set -e

echo "🔍 Detecting your OS and architecture..."

OS="$(uname -s)"
ARCH="$(uname -m)"

# Normalize OS
case "$OS" in
  Linux*)   PLATFORM="linux";;
  Darwin*)  PLATFORM="darwin";;
  *)        echo "❌ Unsupported OS: $OS"; exit 1;;
esac

# Normalize ARCH
case "$ARCH" in
  x86_64)   ARCH="amd64";;
  arm64)    ARCH="arm64";;
  aarch64)  ARCH="arm64";;
  *)        echo "❌ Unsupported architecture: $ARCH"; exit 1;;
esac

# Build the URL
BINARY_URL="https://github.com/floppa/yxa-cli/releases/latest/download/yxa-${PLATFORM}-${ARCH}"

echo "📦 Downloading yxa from: $BINARY_URL"
curl -L "$BINARY_URL" -o yxa

chmod +x yxa
sudo mv yxa /usr/local/bin/

echo "✅ yxa installed successfully at /usr/local/bin/yxa"
echo "👉 Run 'yxa --help' to get started."
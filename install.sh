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
  x86_64)         ARCH="amd64";;
  arm64|aarch64)  ARCH="arm64";;
  *)              echo "❌ Unsupported architecture: $ARCH"; exit 1;;
esac

# Build download URL
BINARY_URL="https://github.com/floppa/yxa-cli/releases/latest/download/yxa-${PLATFORM}-${ARCH}"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/yxa"

echo "📦 Ready to download yxa from: $BINARY_URL"

# Prompt to create install directory
if [ ! -d "$INSTALL_DIR" ]; then
  read -p "📁 Install directory '$INSTALL_DIR' does not exist. Create it? [y/N]: " CREATE_DIR
  if [[ "$CREATE_DIR" =~ ^[Yy]$ ]]; then
    mkdir -p "$INSTALL_DIR"
    echo "✅ Created $INSTALL_DIR"
  else
    echo "❌ Cannot proceed without install directory. Aborting."
    exit 1
  fi
fi

# Download binary
curl -L "$BINARY_URL" -o "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Add to PATH if needed
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo "📍 $INSTALL_DIR is not in your PATH."

  if [ -n "$BASH_VERSION" ]; then
    SHELL_RC="$HOME/.bashrc"
  elif [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
  else
    SHELL_RC="$HOME/.profile"
  fi

  echo "📄 Will add export to: $SHELL_RC"
  read -p "➕ Do you want to add '$INSTALL_DIR' to your PATH in $SHELL_RC? [y/N]: " ADD_PATH
  if [[ "$ADD_PATH" =~ ^[Yy]$ ]]; then
    if [ ! -f "$SHELL_RC" ]; then
      read -p "⚠️ $SHELL_RC does not exist. Create it? [y/N]: " CREATE_RC
      if [[ "$CREATE_RC" =~ ^[Yy]$ ]]; then
        touch "$SHELL_RC"
        echo "📄 Created $SHELL_RC"
      else
        echo "❌ Cannot add to PATH without shell config file. Skipping."
        exit 1
      fi
    fi

    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
    echo "✅ Added to PATH in $SHELL_RC"
    echo "🔁 Please restart your shell or run: source $SHELL_RC"
  else
    echo "ℹ️ Skipping PATH update. You can add '$INSTALL_DIR' manually later."
  fi
fi

echo "✅ yxa installed successfully at $BINARY_PATH"
echo "👉 Run 'yxa --help' to get started."
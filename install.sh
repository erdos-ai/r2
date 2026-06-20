#!/bin/sh
set -e  # Exit on error

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize architecture names to match GoReleaser output
case "$ARCH" in
  x86_64|amd64) ARCH="x86_64" ;;
  i386|i686) ARCH="i386" ;;
  armv7l) ARCH="armv7" ;;
  aarch64) ARCH="arm64" ;;
esac

# Normalize OS names to match GoReleaser output (title case)
case "$OS" in
  linux) OS="Linux" ;;
  darwin) OS="Darwin" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

BINARY_NAME="r2"
TAG_NAME="v0.4.0"
REPO_USER="erdos-ai"
REPO_NAME="r2"

ARCHIVE="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO_USER}/${REPO_NAME}/releases/download/${TAG_NAME}/${ARCHIVE}"
INSTALL_DIR="$HOME/${BINARY_NAME}-${TAG_NAME}"

# Cleanup function
cleanup() {
  rm -f "$HOME/$ARCHIVE"
}
trap cleanup EXIT

echo "Downloading ${BINARY_NAME} ${TAG_NAME} for ${OS}-${ARCH}..."
curl -fsSL -o "$HOME/$ARCHIVE" "$DOWNLOAD_URL" || {
  echo "Error: Failed to download from $DOWNLOAD_URL" >&2
  exit 1
}

echo "Extracting to ${INSTALL_DIR}..."
mkdir -p "$INSTALL_DIR"
tar -xzf "$HOME/$ARCHIVE" -C "$INSTALL_DIR" || {
  echo "Error: Failed to extract archive" >&2
  exit 1
}

chmod +x "$INSTALL_DIR/$BINARY_NAME"

if [ "$(id -u)" -eq 0 ]; then
  echo "Installing to /usr/bin/${BINARY_NAME}..."
  mv "$INSTALL_DIR/$BINARY_NAME" "/usr/bin/$BINARY_NAME"
  rm -rf "$INSTALL_DIR"
  echo "Installation complete! Run '${BINARY_NAME} --version' to verify."
else
  echo "Installation successful!"
  echo ""
  echo "Since you're not running as root, manual steps required:"
  echo "  sudo mv \"$INSTALL_DIR/$BINARY_NAME\" \"/usr/bin/$BINARY_NAME\""
  echo "  rm -rf \"$INSTALL_DIR\""
  echo ""
  echo "Or add to your PATH:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi

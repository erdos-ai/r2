#!/bin/bash
set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Get repository information with validation
REPO_USER=$(git config --get remote.origin.url | cut -d'/' -f4)
REPO_NAME=$(basename "$(git rev-parse --show-toplevel)")
TAG_NAME=$(git describe --tags --abbrev=0)
BINARY_NAME="r2"

# Validate required variables
if [ -z "$REPO_USER" ] || [ -z "$REPO_NAME" ] || [ -z "$TAG_NAME" ]; then
  echo "Error: Failed to determine repository information" >&2
  echo "REPO_USER: ${REPO_USER:-empty}" >&2
  echo "REPO_NAME: ${REPO_NAME:-empty}" >&2
  echo "TAG_NAME: ${TAG_NAME:-empty}" >&2
  exit 1
fi

cat > install.sh << 'EOF'
#!/bin/sh
set -e  # Exit on error

OS=$(uname -s)
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

EOF

# Insert dynamic values (use double quotes for variable expansion)
cat >> install.sh << EOF
BINARY_NAME="${BINARY_NAME}"
TAG_NAME="${TAG_NAME}"
REPO_USER="${REPO_USER}"
REPO_NAME="${REPO_NAME}"

EOF

# Continue with static content (use single quotes to prevent expansion)
cat >> install.sh << 'EOF'
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
EOF

chmod +x install.sh
echo "Generated install.sh for ${TAG_NAME}"

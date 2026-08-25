#!/bin/sh
set -eu

# Agent Ledger Unix Installer (macOS and Linux)
# Repository: https://github.com/rahumanrahuu/agent-ledger

REPO="rahumanrahuu/agent-ledger"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Check required system utilities
for tool in curl tar uname; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "Error: Required system tool '$tool' is not installed or not in PATH." >&2
    exit 1
  fi
done

# Detect operating system
OS_TYPE="$(uname -s)"
case "$OS_TYPE" in
  Darwin)
    OS="darwin"
    ;;
  Linux)
    OS="linux"
    ;;
  *)
    echo "Error: Unsupported operating system: $OS_TYPE" >&2
    echo "Agent Ledger install.sh supports macOS and Linux. For Windows, please run install.ps1." >&2
    exit 1
    ;;
esac

# Detect CPU architecture
ARCH_TYPE="$(uname -m)"
case "$ARCH_TYPE" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH_TYPE" >&2
    echo "Agent Ledger supports amd64 (x86_64) and arm64 architectures." >&2
    exit 1
    ;;
esac

# Determine release version
if [ -n "${VERSION:-}" ]; then
  case "$VERSION" in
    v*) TAG="$VERSION" ;;
    *) TAG="v$VERSION" ;;
  esac
  echo "Installing requested version: $TAG"
else
  echo "Determining latest release for $REPO..."
  TAG=""
  
  # Try GitHub Releases API first
  API_RESPONSE="$(curl -fsSL -H "Accept: application/vnd.github.v3+json" "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null || true)"
  if [ -n "$API_RESPONSE" ]; then
    TAG="$(echo "$API_RESPONSE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | head -n 1)"
  fi
  
  # Fallback to redirect URL inspection
  if [ -z "$TAG" ]; then
    REDIRECT_HEADER="$(curl -fsSI "https://github.com/$REPO/releases/latest" 2>/dev/null | grep -i "^location:" || true)"
    if [ -n "$REDIRECT_HEADER" ]; then
      TAG="$(echo "$REDIRECT_HEADER" | sed -E 's/.*tag\/(.*)/\1/' | tr -d '\r\n')"
    fi
  fi
  
  # Fallback default if API rate-limited and no releases yet
  if [ -z "$TAG" ]; then
    TAG="v0.2.1"
    echo "Notice: Could not query latest GitHub release tag dynamically; defaulting to $TAG."
  else
    echo "Found latest version: $TAG"
  fi
fi

ARCHIVE_NAME="agent-ledger_${TAG}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$ARCHIVE_NAME"

# Create secure temporary directory
TMP_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'agent-ledger-install')"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

echo "Downloading $ARCHIVE_NAME..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"; then
  echo "Error: Failed to download release archive from $DOWNLOAD_URL." >&2
  echo "Please verify that version $TAG exists at https://github.com/$REPO/releases" >&2
  exit 1
fi

echo "Extracting binaries..."
tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/agent-ledger" ] || [ ! -f "$TMP_DIR/ledger-mcp" ]; then
  echo "Error: Release archive did not contain expected executables (agent-ledger, ledger-mcp)." >&2
  exit 1
fi

# Ensure user-local install directory exists
mkdir -p "$INSTALL_DIR"

echo "Installing binaries into $INSTALL_DIR..."
cp "$TMP_DIR/agent-ledger" "$INSTALL_DIR/agent-ledger"
cp "$TMP_DIR/ledger-mcp" "$INSTALL_DIR/ledger-mcp"
chmod +x "$INSTALL_DIR/agent-ledger" "$INSTALL_DIR/ledger-mcp"

# Verify installation
if [ ! -x "$INSTALL_DIR/agent-ledger" ] || [ ! -x "$INSTALL_DIR/ledger-mcp" ]; then
  echo "Error: Installation verification failed. Files are not executable in $INSTALL_DIR." >&2
  exit 1
fi

echo "Successfully installed Agent Ledger ($TAG)!"
echo "  - CLI: $INSTALL_DIR/agent-ledger"
echo "  - MCP: $INSTALL_DIR/ledger-mcp"
echo

# Check if INSTALL_DIR is in PATH
case ":$PATH:" in
  *:"$INSTALL_DIR":*)
    echo "Verification:"
    echo "  Run 'agent-ledger --help' to get started."
    echo "  Run 'ledger-mcp --help' to view MCP server details."
    ;;
  *)
    echo "Notice: $INSTALL_DIR is not currently in your PATH."
    echo "Add it to your environment by running:"
    echo
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo
    echo "To make this permanent, add the line above to your ~/.bashrc, ~/.zshrc, or ~/.profile."
    ;;
esac

#!/bin/sh
set -eu

# Agent Ledger Unix Installer (macOS and Linux)
# Repository: https://github.com/rahumanrahuu/agent-ledger

REPO="rahumanrahuu/agent-ledger"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
FALLBACK_TAGS="v0.2.2 v0.2.0"

asset_exists() {
  tag="$1"; archive="$2"
  curl -fsSI -H "User-Agent: agent-ledger-installer" \
    "https://github.com/$REPO/releases/download/$tag/$archive" >/dev/null 2>&1
}

latest_tag() {
  API_RESPONSE="$(curl -fsSL -H "Accept: application/vnd.github.v3+json" -H "User-Agent: agent-ledger-installer" "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null || true)"
  if [ -n "$API_RESPONSE" ]; then
    TAG_CANDIDATE="$(echo "$API_RESPONSE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | head -n 1)"
    if [ -n "$TAG_CANDIDATE" ] && asset_exists "$TAG_CANDIDATE" "agent-ledger_${TAG_CANDIDATE}_${OS}_${ARCH}.tar.gz"; then
      echo "$TAG_CANDIDATE"
      return 0
    fi
  fi
  REDIRECT_HEADER="$(curl -fsSI "https://github.com/$REPO/releases/latest" 2>/dev/null | grep -i "^location:" || true)"
  if [ -n "$REDIRECT_HEADER" ]; then
    TAG_CANDIDATE="$(echo "$REDIRECT_HEADER" | sed -E 's/.*tag\/(.*)/\1/' | tr -d '\r\n')"
    if [ -n "$TAG_CANDIDATE" ] && asset_exists "$TAG_CANDIDATE" "agent-ledger_${TAG_CANDIDATE}_${OS}_${ARCH}.tar.gz"; then
      echo "$TAG_CANDIDATE"
      return 0
    fi
  fi
  for FALLBACK in $FALLBACK_TAGS; do
    if asset_exists "$FALLBACK" "agent-ledger_${FALLBACK}_${OS}_${ARCH}.tar.gz"; then
      echo "$FALLBACK"
      return 0
    fi
  done
  return 1
}

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

  # Auto-fallback to the latest available release if the requested one has no assets
  if ! asset_exists "$TAG" "agent-ledger_${TAG}_${OS}_${ARCH}.tar.gz"; then
    echo "Notice: Version $TAG has no published binaries for ${OS}/${ARCH}." >&2
    echo "Falling back to the latest available release automatically..."
    TAG=""
  fi
fi

if [ -z "${TAG:-}" ]; then
  echo "Determining latest release for $REPO..."
  if TAG="$(latest_tag)"; then
    echo "Using version: $TAG"
  else
    echo "Error: Unable to determine an available release. Check https://github.com/$REPO/releases" >&2
    exit 1
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

if [ ! -f "$TMP_DIR/agent-ledger" ]; then
  echo "Error: Release archive did not contain expected executable (agent-ledger)." >&2
  exit 1
fi

# Ensure user-local install directory exists
mkdir -p "$INSTALL_DIR"

echo "Installing binary into $INSTALL_DIR..."
cp "$TMP_DIR/agent-ledger" "$INSTALL_DIR/agent-ledger"
chmod +x "$INSTALL_DIR/agent-ledger"

# Verify installation
if [ ! -x "$INSTALL_DIR/agent-ledger" ]; then
  echo "Error: Installation verification failed. File is not executable in $INSTALL_DIR." >&2
  exit 1
fi

echo "Successfully installed Agent Ledger ($TAG)!"
echo "  - Binary: $INSTALL_DIR/agent-ledger"
echo

# Persist INSTALL_DIR in the user's shell startup file
case "${SHELL:-}" in
  */zsh)
    PATH_CONFIG="${ZDOTDIR:-$HOME}/.zshrc"
    ;;
  */bash)
    if [ -f "$HOME/.bash_profile" ]; then
      PATH_CONFIG="$HOME/.bash_profile"
    else
      PATH_CONFIG="$HOME/.bashrc"
    fi
    ;;
  *)
    PATH_CONFIG="$HOME/.profile"
    ;;
esac

PATH_EXPORT="export PATH=\"$INSTALL_DIR:\$PATH\""
if [ ! -f "$PATH_CONFIG" ] || ! grep -Fq "$INSTALL_DIR" "$PATH_CONFIG"; then
  printf '\n# Agent Ledger\n%s\n' "$PATH_EXPORT" >> "$PATH_CONFIG"
  echo "Added $INSTALL_DIR to PATH in $PATH_CONFIG."
fi

case ":$PATH:" in
  *:"$INSTALL_DIR":*)
    ;;
  *)
    export PATH="$INSTALL_DIR:$PATH"
    ;;
esac

echo "Verification:"
echo "  Run 'agent-ledger --help' to get started."
echo "  Run 'agent-ledger mcp' to start the MCP server."
echo "  Restart your terminal or run: . \"$PATH_CONFIG\""

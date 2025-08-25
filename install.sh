#!/usr/bin/env bash
set -euo pipefail

# Usage: ./install-bind9-api.sh 0.1.0
VERSION="$1"
REPO="EpicLabs23/bind9-api"
BINARY="bind9-api"
INSTALL_DIR="/usr/local/bin"
SYSTEMD_DIR="/etc/systemd/system"

echo "🔍 Installing $BINARY version $VERSION..."

# Download tarball
URL="https://github.com/$REPO/releases/download/v$VERSION/${BINARY}_Linux_x86_64.tar.gz"
URL="https://github.com/$REPO/releases/download/v$VERSION/${BINARY}_${VERSION}_linux_amd64.tar.gz"
TMP_DIR=$(mktemp -d)

echo "⬇️  Downloading from $URL"
curl -sSL "$URL" -o "$TMP_DIR/$BINARY.tar.gz"

# Extract
echo "📦 Extracting..."
tar -xzf "$TMP_DIR/$BINARY.tar.gz" -C "$TMP_DIR"

# Install binary
echo "⚙️  Installing binary to $INSTALL_DIR"
sudo install -m 755 "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"

# Install systemd service
if [ -f "$TMP_DIR/packaging/systemd/$BINARY.service" ]; then
    echo "⚙️  Installing systemd service"
    sudo cp "$TMP_DIR/packaging/systemd/$BINARY.service" "$SYSTEMD_DIR/"
else
    echo "⚠️  No systemd service file found in packaging/systemd/"
fi

# Reload and restart systemd service
echo "🔄 Reloading systemd"
sudo systemctl daemon-reload
sudo systemctl enable "$BINARY.service"
sudo systemctl restart "$BINARY.service"

echo "✅ Installation complete. Service status:"
sudo systemctl status "$BINARY.service" --no-pager

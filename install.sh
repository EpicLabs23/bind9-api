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
TMP_DIR=$(mktemp -d)

echo "⬇️  Downloading from $URL"
curl -sSL "$URL" -o "$TMP_DIR/$BINARY.tar.gz"

# Extract
echo "📦 Extracting..."
tar -xzf "$TMP_DIR/$BINARY.tar.gz" -C "$TMP_DIR"

# Install config file
echo "⚙️  Installing config file"
if [ ! -f /etc/bind9-api/config.yaml ]; then
    cp "$TMP_DIR/config.yaml" /etc/bind9-api/config.yaml
else
    echo "⚠️  Config file already exists, skipping..."
fi

# Install binary
echo "⚙️  Installing binary to $INSTALL_DIR"
install -m 755 "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"

# Install systemd service
if [ -f "$TMP_DIR/packaging/systemd/$BINARY.service" ]; then
    echo "⚙️  Installing systemd service"
    cp "$TMP_DIR/packaging/systemd/$BINARY.service" "$SYSTEMD_DIR/"
else
    echo "⚠️  No systemd service file found in packaging/systemd/"
fi

# Reload and restart systemd service
echo "🔄 Reloading systemd"
systemctl daemon-reload
systemctl enable "$BINARY.service"
systemctl restart "$BINARY.service"

echo "✅ Installation complete. Service status:"
systemctl status "$BINARY.service" --no-pager

# Cleanup
rm -rf "$TMP_DIR"
#!/usr/bin/env bash
set -e

REPO="yourusername/bind9-api"   # <-- change to your GitHub repo
BIN_NAME="bind9-api"
INSTALL_DIR="/usr/local/bin"
SYSTEMD_PATH="/etc/systemd/system/${BIN_NAME}.service"

echo "🔍 Fetching latest release info..."
latest_release=$(curl -s "https://api.github.com/repos/${REPO}/releases" \
  | jq -r 'map(select(.prerelease == false)) | sort_by(.published_at) | last')

tag_name=$(echo "$latest_release" | jq -r .tag_name)
if [[ -z "$tag_name" || "$tag_name" == "null" ]]; then
  echo "❌ Could not find a release. Make sure you have published releases."
  exit 1
fi

echo "📦 Latest version: $tag_name"

# Detect OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  armv7l) ARCH="armv7" ;;
esac

echo "💻 Detected OS=$OS ARCH=$ARCH"

# Download binary
asset_url=$(echo "$latest_release" \
  | jq -r ".assets[] | select(.name | test(\"${OS}_${ARCH}\")) | .browser_download_url")

if [[ -z "$asset_url" ]]; then
  echo "❌ No asset found for ${OS}_${ARCH}"
  exit 1
fi

tmp_file=$(mktemp)
echo "⬇️  Downloading binary..."
curl -L -o "$tmp_file" "$asset_url"
chmod +x "$tmp_file"
sudo mv "$tmp_file" "${INSTALL_DIR}/${BIN_NAME}"

echo "✅ Installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}"

# Install or update systemd service
echo "⚙️  Installing systemd service..."
curl -sL "https://raw.githubusercontent.com/${REPO}/main/packaging/systemd/${BIN_NAME}.service" \
  | sudo tee "$SYSTEMD_PATH" >/dev/null

sudo systemctl daemon-reload
sudo systemctl enable "${BIN_NAME}.service"
sudo systemctl restart "${BIN_NAME}.service"

echo "🚀 ${BIN_NAME} ${tag_name} is installed and running."
echo "👉 To check status: sudo systemctl status ${BIN_NAME}.service"

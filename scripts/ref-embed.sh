#!/bin/bash
set -e

# Download and embed lazydocker (Linux amd64)

EMBED_DIR="cmd/embed"
mkdir -p "$EMBED_DIR"

echo "Fetching latest lazydocker version..."
GITHUB_LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/jesseduffield/lazydocker/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
GITHUB_FILE="lazydocker_${GITHUB_LATEST_VERSION//v/}_Linux_x86_64.tar.gz"
GITHUB_URL="https://github.com/jesseduffield/lazydocker/releases/download/${GITHUB_LATEST_VERSION}/${GITHUB_FILE}"

echo "Downloading $GITHUB_URL..."
curl -L -o /tmp/lazydocker.tar.gz "$GITHUB_URL"

echo "Extracting..."
tar -xzf /tmp/lazydocker.tar.gz -C /tmp lazydocker

echo "Compressing with upx..."
upx --best --lzma /tmp/lazydocker

echo "Moving to $EMBED_DIR/..."
mv /tmp/lazydocker "$EMBED_DIR/lazydocker"

rm -f /tmp/lazydocker.tar.gz

echo "Done! File size: $(du -h "$EMBED_DIR/lazydocker" | cut -f1)"


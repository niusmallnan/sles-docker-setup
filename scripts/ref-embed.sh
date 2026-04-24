#!/bin/bash
set -e

# Download and embed lazydocker and trivy (Linux amd64)

EMBED_DIR="cmd/embed"
mkdir -p "$EMBED_DIR"

# ===== lazydocker =====
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

echo "lazydocker done! File size: $(du -h "$EMBED_DIR/lazydocker" | cut -f1)"

# ===== trivy =====
echo ""
echo "Fetching latest trivy version..."
GITHUB_LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/aquasecurity/trivy/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
GITHUB_FILE="trivy_${GITHUB_LATEST_VERSION//v/}_Linux-64bit.tar.gz"
GITHUB_URL="https://github.com/aquasecurity/trivy/releases/download/${GITHUB_LATEST_VERSION}/${GITHUB_FILE}"

echo "Downloading $GITHUB_URL..."
curl -L -o /tmp/trivy.tar.gz "$GITHUB_URL"

echo "Extracting..."
tar -xzf /tmp/trivy.tar.gz -C /tmp trivy

echo "Compressing with upx..."
upx --best --lzma /tmp/trivy

echo "Moving to $EMBED_DIR/..."
mv /tmp/trivy "$EMBED_DIR/trivy"

rm -f /tmp/trivy.tar.gz

echo "trivy done! File size: $(du -h "$EMBED_DIR/trivy" | cut -f1)"

echo ""
echo "All binaries embedded successfully!"


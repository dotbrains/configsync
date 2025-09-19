#!/bin/bash

# Build script for ConfigSync
set -e

VERSION=${1:-"dev"}
OUTPUT_DIR="bin"

echo "Building ConfigSync v${VERSION}..."

# Clean previous builds
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# Build for macOS (Intel)
echo "Building for macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/dotbrains/configsync/cmd/configsync/cmd.version=${VERSION}" -o ${OUTPUT_DIR}/configsync-darwin-amd64 ./cmd/configsync

# Build for macOS (Apple Silicon)
echo "Building for macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/dotbrains/configsync/cmd/configsync/cmd.version=${VERSION}" -o ${OUTPUT_DIR}/configsync-darwin-arm64 ./cmd/configsync

# Create universal binary
echo "Creating universal binary..."
lipo -create -output ${OUTPUT_DIR}/configsync-darwin-universal ${OUTPUT_DIR}/configsync-darwin-amd64 ${OUTPUT_DIR}/configsync-darwin-arm64

# Create checksums
echo "Creating checksums..."
cd ${OUTPUT_DIR}
shasum -a 256 * > checksums.txt
cd ..

echo "Build complete! Files in ${OUTPUT_DIR}:"
ls -la ${OUTPUT_DIR}/

echo ""
echo "To install:"
echo "  cp ${OUTPUT_DIR}/configsync-darwin-universal /usr/local/bin/configsync"
echo "  chmod +x /usr/local/bin/configsync"

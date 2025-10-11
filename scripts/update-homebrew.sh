#!/bin/bash

# Script to manually update the Homebrew formula for ConfigSync
# Usage: ./scripts/update-homebrew.sh v1.0.3

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if version argument is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Please provide a version tag${NC}"
    echo "Usage: $0 v1.0.3"
    exit 1
fi

VERSION=$1
VERSION_NUMBER=${VERSION#v}  # Remove 'v' prefix

echo -e "${GREEN}🍺 Updating Homebrew formula for ConfigSync ${VERSION}${NC}"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
TAP_DIR="${TEMP_DIR}/homebrew-tap"

echo -e "${YELLOW}📁 Cloning Homebrew tap...${NC}"
git clone https://github.com/dotbrains/homebrew-tap.git "${TAP_DIR}"
cd "${TAP_DIR}"

# Download the universal binary to calculate SHA256
DOWNLOAD_URL="https://github.com/dotbrains/configsync/releases/download/${VERSION}/configsync-${VERSION}-darwin-universal.tar.gz"
TEMP_FILE="${TEMP_DIR}/configsync.tar.gz"

echo -e "${YELLOW}📦 Downloading release binary to calculate SHA256...${NC}"
curl -L "${DOWNLOAD_URL}" -o "${TEMP_FILE}"

# Calculate SHA256
SHA256=$(shasum -a 256 "${TEMP_FILE}" | cut -d' ' -f1)
echo -e "${GREEN}🔍 Calculated SHA256: ${SHA256}${NC}"

# Update the formula
FORMULA_FILE="Formula/configsync.rb"

if [ ! -f "${FORMULA_FILE}" ]; then
    echo -e "${RED}Error: Formula file not found at ${FORMULA_FILE}${NC}"
    exit 1
fi

echo -e "${YELLOW}✏️  Updating formula file...${NC}"

# Create backup
cp "${FORMULA_FILE}" "${FORMULA_FILE}.bak"

# Update version and SHA256 in the formula
sed -i.tmp "s/version \".*\"/version \"${VERSION_NUMBER}\"/" "${FORMULA_FILE}"
sed -i.tmp "s/sha256 \".*\"/sha256 \"${SHA256}\"/" "${FORMULA_FILE}"
sed -i.tmp "s|url \".*\"|url \"${DOWNLOAD_URL}\"|" "${FORMULA_FILE}"

# Remove temporary sed file
rm -f "${FORMULA_FILE}.tmp"

# Show the diff
echo -e "${YELLOW}📋 Changes made:${NC}"
git diff "${FORMULA_FILE}"

# Commit and push
echo -e "${YELLOW}💾 Committing changes...${NC}"
git add "${FORMULA_FILE}"
git commit -m "Update configsync to ${VERSION}"

echo -e "${YELLOW}🚀 Pushing to GitHub...${NC}"
git push origin main

echo -e "${GREEN}✅ Successfully updated Homebrew formula for ConfigSync ${VERSION}${NC}"
echo -e "${GREEN}🎉 Users can now install with: brew upgrade configsync${NC}"

# Change back to original directory before cleanup and testing
cd "$HOME"

# Clean up
rm -rf "${TEMP_DIR}"

# Test the formula (optional)
echo -e "${YELLOW}🧪 Testing the updated formula...${NC}"
if command -v brew >/dev/null 2>&1; then
    echo -e "${YELLOW}Running brew audit...${NC}"
    brew audit --strict dotbrains/tap/configsync || echo -e "${YELLOW}⚠️  Audit warnings (usually non-critical)${NC}"
else
    echo -e "${YELLOW}Homebrew not installed, skipping formula test${NC}"
fi

echo -e "${GREEN}🏁 Manual Homebrew update complete!${NC}"

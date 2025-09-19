# Homebrew Setup Guide

This guide explains how to set up Homebrew support for ConfigSync.

## 1. Create the Homebrew Tap Repository

1. Go to GitHub and create a new repository named: `homebrew-tap`
2. Make it public (required for Homebrew taps)
3. Add a description: "Homebrew tap for dotbrains tools"
4. Initialize with a README

## 2. Clone and Set Up the Tap Repository

```bash
# Clone the tap repository
git clone https://github.com/dotbrains/homebrew-tap.git
cd homebrew-tap

# Create the Formula directory
mkdir Formula

# Create the configsync formula
cat > Formula/configsync.rb << 'EOF'
class Configsync < Formula
  desc "Synchronize macOS application configurations across machines"
  homepage "https://github.com/dotbrains/configsync"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/dotbrains/configsync/releases/download/v1.0.0/configsync-v1.0.0-darwin-amd64.tar.gz"
      sha256 "ff361bd9156d8d922fe9963c9b4d8efc6d602042861eb218bc8193087afed86b"
    elsif Hardware::CPU.arm?
      url "https://github.com/dotbrains/configsync/releases/download/v1.0.0/configsync-v1.0.0-darwin-arm64.tar.gz"
      sha256 "69bc5d100c86001c8bc4e1856ea0a48a45ca84b48c601d7dc4f6d14996dc4ab5"
    end
  end

  def install
    bin.install "configsync-darwin-amd64" => "configsync" if Hardware::CPU.intel?
    bin.install "configsync-darwin-arm64" => "configsync" if Hardware::CPU.arm?
  end

  test do
    system "#{bin}/configsync", "--version"
    system "#{bin}/configsync", "--help"
  end
end
EOF

# Commit and push
git add .
git commit -m "feat: add configsync formula v1.0.0"
git push origin main
```

## 3. Set Up GitHub Token for Automated Updates

1. Go to GitHub Settings > Developer Settings > Personal Access Tokens
2. Create a new token with these permissions:
   - `public_repo` (for public repositories)
   - `workflow` (to update workflows)
3. Add the token as a repository secret in your main `configsync` repo:
   - Go to Settings > Secrets and variables > Actions
   - Add new secret named `HOMEBREW_TOKEN`
   - Paste your token value

## 4. Test the Homebrew Installation

```bash
# Add your tap
brew tap dotbrains/tap

# Install configsync
brew install configsync

# Test the installation
configsync --version
configsync --help

# Clean up (optional)
brew uninstall configsync
brew untap dotbrains/tap
```

## 5. Automated Updates

The release workflow is already configured to automatically update the Homebrew formula when you create new releases. The workflow will:

1. Detect new tags starting with 'v' (e.g., v1.1.0)
2. Download the new release assets
3. Calculate SHA256 checksums
4. Update the formula in your tap repository
5. Commit and push the changes

## Usage Instructions for Users

Once set up, users can install ConfigSync with:

```bash
# Install
brew install dotbrains/tap/configsync

# Upgrade (when new versions are available)
brew upgrade configsync

# Uninstall
brew uninstall configsync
```

## Troubleshooting

### Formula Validation
```bash
# Test the formula locally
brew install --build-from-source dotbrains/tap/configsync
brew test configsync
brew audit --strict configsync
```

### Manual Formula Updates
If automatic updates fail, you can manually update:

```bash
# Clone your tap repo
git clone https://github.com/dotbrains/homebrew-tap.git
cd homebrew-tap

# Edit Formula/configsync.rb with new version, URLs, and SHA256s
# Commit and push changes
git add Formula/configsync.rb
git commit -m "feat: update configsync to v1.x.x"
git push origin main
```

### Getting SHA256 Checksums
```bash
# Download and get checksums for new releases
curl -L https://github.com/dotbrains/configsync/releases/download/v1.0.0/checksums.txt
```

# Release Workflow Guide

This document outlines the release process for ConfigSync, including both automated and manual procedures.

## Current Release Process

### 1. Automated GitHub Release ✅
- **Trigger**: Push a git tag (e.g., `v1.0.3`)
- **What it does**:
  - Builds binaries for macOS (Intel, ARM64, Universal)
  - Creates GitHub release with binaries and changelog
  - Generates checksums and installation instructions

### 2. Homebrew Formula Update ⚠️
- **Automated**: Uses third-party action (has edge cases)
- **Manual fallback**: Script provided for reliable updates

## Creating a New Release

### Step 1: Prepare Release

```bash
# 1. Update CHANGELOG.md with new version
vim CHANGELOG.md

# 2. Commit changelog
git add CHANGELOG.md
git commit -m "chore: prepare v1.0.3 release"

# 3. Push changes
git push origin main
```

### Step 2: Create Release Tag

```bash
# Create and push tag
git tag v1.0.3
git push origin v1.0.3
```

This triggers the automated GitHub Actions workflow.

### Step 3: Monitor Release Workflow

```bash
# Check workflow status
gh run list --workflow=Release --limit=5

# View specific run if needed
gh run view <run-id>
```

### Step 4: Handle Homebrew Update

#### Option A: Automated Success ✅
If the GitHub Actions workflow completes successfully, the Homebrew formula is automatically updated.

#### Option B: Automated Failure → Manual Update 🛠️
If the Homebrew step fails (common with edge cases), use the manual script:

```bash
# Run manual update script
./scripts/update-homebrew.sh v1.0.3
```

The script will:
1. Clone the Homebrew tap repository
2. Download the release binary
3. Calculate SHA256 checksum
4. Update the formula with new version and checksum
5. Commit and push changes
6. Test the formula (if Homebrew is available)

#### Option C: Completely Manual Update 📝
If you prefer to update manually:

1. **Clone the tap repository:**
   ```bash
   git clone https://github.com/dotbrains/homebrew-tap.git
   cd homebrew-tap
   ```

2. **Get the release information:**
   ```bash
   VERSION="v1.0.3"
   URL="https://github.com/dotbrains/configsync/releases/download/${VERSION}/configsync-${VERSION}-darwin-universal.tar.gz"

   # Download and calculate SHA256
   curl -L "${URL}" | shasum -a 256
   ```

3. **Update `Formula/configsync.rb`:**
   ```ruby
   class Configsync < Formula
     desc "Synchronize macOS application configurations across machines"
     homepage "https://github.com/dotbrains/configsync"
     url "https://github.com/dotbrains/configsync/releases/download/v1.0.3/configsync-v1.0.3-darwin-universal.tar.gz"
     sha256 "NEW_SHA256_HERE"
     version "1.0.3"
     # ... rest of formula
   ```

4. **Commit and push:**
   ```bash
   git add Formula/configsync.rb
   git commit -m "Update configsync to v1.0.3"
   git push origin main
   ```

## Verifying Releases

### Test GitHub Release
```bash
# Test download and installation
curl -L https://github.com/dotbrains/configsync/releases/latest/download/configsync-darwin-universal.tar.gz | tar -xz
./configsync-darwin-universal --version
```

### Test Homebrew Formula
```bash
# Test the formula
brew audit --strict dotbrains/tap/configsync

# Test installation (if you have a test environment)
brew install dotbrains/tap/configsync
configsync --version
```

## Troubleshooting Common Issues

### Homebrew Automation Failures

**Issue**: "You need to bump this formula manually since the new version and old version are both X.X.X"
- **Cause**: The action thinks the version hasn't changed
- **Solution**: Use the manual update script or add `force: true` to the action

**Issue**: "Cannot open: File exists" errors
- **Cause**: Cache conflicts in GitHub Actions
- **Solution**: The workflow has been updated with `continue-on-error: true`

**Issue**: Checksum verification failures
- **Cause**: The action can't properly calculate checksums
- **Solution**: Manual update ensures correct checksums

### Version Issues

**Issue**: Wrong version reported in binary
- **Cause**: Version injection during build failed
- **Solution**: Check the `ldflags` in the build step

## Best Practices

1. **Always update CHANGELOG.md** before releasing
2. **Test locally** before creating tags when possible
3. **Monitor the release workflow** after pushing tags
4. **Have the manual script ready** as a backup
5. **Verify both GitHub and Homebrew** distributions work
6. **Tag commits on main branch** only

## Release Checklist

- [ ] Update CHANGELOG.md with new version
- [ ] Commit changelog changes
- [ ] Push to main branch
- [ ] Create and push version tag
- [ ] Monitor GitHub Actions workflow
- [ ] Verify GitHub release is created correctly
- [ ] Check Homebrew automation status
- [ ] Run manual Homebrew update if needed
- [ ] Test both installation methods
- [ ] Announce release (if applicable)

## Rollback Process

If a release needs to be rolled back:

1. **Delete the problematic tag:**
   ```bash
   git tag -d v1.0.3
   git push origin --delete v1.0.3
   ```

2. **Delete GitHub release:**
   ```bash
   gh release delete v1.0.3
   ```

3. **Revert Homebrew formula** (if updated):
   ```bash
   # Clone tap and revert the formula to previous version
   git clone https://github.com/dotbrains/homebrew-tap.git
   cd homebrew-tap
   git revert HEAD  # if the update was the last commit
   git push origin main
   ```

## Automation Improvements

The workflow has been updated to be more resilient:
- ✅ Force updates enabled to handle version conflicts
- ✅ Continue-on-error prevents Homebrew failures from breaking releases
- ✅ Manual instructions displayed when automation fails
- ✅ Manual update script provided as reliable fallback

This provides the best of both worlds: automation when it works, reliable manual process when it doesn't.

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.3] - 2024-09-19

### Added
- **Release Documentation**: Comprehensive release workflow guide (RELEASE.md)
- **Manual Update Script**: Automated script for manual Homebrew formula updates
- **Enhanced Automation**: Improved GitHub Actions workflow with better error handling
- **Fallback Process**: Reliable manual process when automation fails

### Changed
- **Release Workflow**: Enhanced with `force: true` and `continue-on-error` for better reliability
- **Documentation**: Added reference to release documentation in README
- **Error Handling**: Better feedback when Homebrew automation fails

### Fixed
- **Homebrew Edge Cases**: Improved handling of version conflicts and cache issues
- **Release Resilience**: Workflow no longer fails completely if Homebrew update fails

## [1.0.2] - 2024-09-19

### Fixed
- **Version Injection**: Fixed build-time version injection to show correct version in `--version` output
- **Homebrew Automation**: Resolved workflow issues with automated formula updates
- **Build Process**: Updated ldflags to use correct package path for version variable

### Changed
- Updated build scripts and workflows to properly inject version information
- Improved release automation reliability

## [1.0.1] - 2024-09-19

### Added
- **Homebrew Support**: Official Homebrew tap for easy installation
- **Automated Formula Updates**: Release workflow automatically updates Homebrew formula
- **Multi-Architecture Homebrew Formula**: Native support for Intel and Apple Silicon Macs
- **Enhanced Installation Documentation**: Homebrew as primary installation method
- **Homebrew Setup Guide**: Comprehensive guide for maintaining the tap

### Changed
- Updated README.md to prioritize Homebrew installation
- Enhanced release workflow with automated Homebrew formula updates
- Improved installation instructions across all documentation

### Infrastructure
- Created `dotbrains/homebrew-tap` repository
- Added automated formula update workflow
- Integrated SHA256 checksum validation for Homebrew releases

### Removed
- Homebrew setup guide (no longer needed after successful setup)

## [1.0.0] - 2024-09-19

### Added
- Initial release of ConfigSync
- CLI commands: `init`, `add`, `sync`, `status`, `remove`
- Support for common macOS applications (VS Code, Git, SSH, Terminal, etc.)
- Automatic application detection and configuration path discovery
- Safe symlink management with backup capabilities
- Dry-run mode for testing changes
- Verbose output for troubleshooting
- Comprehensive documentation and examples

### Features
- **Application Management**: Add, remove, and manage application configurations
- **Smart Detection**: Automatically detect configuration files for popular macOS apps
- **Safe Operations**: Dry-run mode and automatic backups before changes
- **Symlink Management**: Create and manage symlinks between original locations and central store
- **Status Monitoring**: Check sync status and detect configuration drift
- **Cross-Architecture Support**: Universal binaries for Intel and Apple Silicon Macs

### Supported Applications
- Visual Studio Code (settings.json, keybindings.json, snippets)
- Git (.gitconfig, .gitignore_global)
- SSH (.ssh/config)
- Terminal (com.apple.Terminal.plist)
- iTerm2 (com.googlecode.iterm2.plist)
- Google Chrome (preferences and settings)
- Firefox (profiles and preferences)
- Sublime Text (user packages and settings)

### Development
- Complete test suite with unit tests
- GitHub Actions for CI/CD (testing, building, releasing)
- Cross-platform build support
- Linting and code quality checks
- Comprehensive documentation

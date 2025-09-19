# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

## [1.0.0] - 2024-XX-XX

### Added
- Initial stable release
- All core features implemented and tested
- Production-ready CLI tool for macOS configuration management
---
layout: default
title: Contributing
permalink: /contributing/
---

# Contributing to ConfigSync

*Help make ConfigSync better for everyone*

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code and help create a welcoming environment for all contributors.

## Ways to Contribute

There are many ways you can contribute to ConfigSync:

### 🐛 Report Bugs
Help us improve by reporting bugs, issues, or unexpected behavior you encounter.

### 💡 Suggest Features
Share your ideas for new features or improvements to existing functionality.

### 💻 Submit Code
Contribute bug fixes, new features, or improvements to the codebase.

### 📝 Improve Documentation
Help make ConfigSync more accessible by improving documentation and examples.

## Reporting Bugs

Before creating bug reports, please check the existing issues to see if the problem has already been reported. When you create a bug report, please include:

### Essential Information

- **Clear, descriptive title** - Summarize the issue in one line
- **Steps to reproduce** - Exact steps that cause the problem
- **Expected behavior** - What you thought should happen
- **Actual behavior** - What actually happened instead
- **System information**:
  - macOS version (e.g., macOS 14.1 Sonoma)
  - ConfigSync version (`configsync --version`)
  - Architecture (Intel or Apple Silicon)

### Additional Context

```bash
# Include verbose output when relevant
configsync status --verbose
configsync sync --dry-run --verbose

# Include configuration if applicable
cat ~/.configsync/config.yaml

# Include recent logs
tail -20 ~/.configsync/logs/configsync.log
```

### Bug Report Template

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run `configsync init`
2. Execute `configsync add vscode`
3. Run `configsync sync`
4. See error

**Expected behavior**
ConfigSync should create symlinks without errors.

**System Information**
- macOS: [e.g., 14.1 Sonoma]
- ConfigSync: [e.g., v1.2.3]
- Architecture: [Intel/Apple Silicon]

**Additional context**
Add any other context, logs, or screenshots.
```

## Suggesting Enhancements

Enhancement suggestions help make ConfigSync better for everyone. When creating an enhancement suggestion:

### What to Include

- **Clear title** - Summarize the enhancement briefly
- **Problem statement** - What problem does this solve?
- **Proposed solution** - How would you like it to work?
- **Use cases** - When would this be useful?
- **Examples** - Show how it would work in practice

### Enhancement Template

```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Other solutions you've thought about.

**Additional context**
Any other context or screenshots about the feature request.
```

## Development Setup

Ready to contribute code? Here's how to set up your development environment:

### Prerequisites

- **Go 1.21 or later** - [Download Go](https://golang.org/dl/)
- **macOS** - Required for testing macOS-specific functionality
- **Git** - For version control
- **Make** - For build automation

### Getting Started

```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR-USERNAME/configsync.git
cd configsync

# 3. Add upstream remote
git remote add upstream https://github.com/dotbrains/configsync.git

# 4. Install dependencies
go mod download

# 5. Build the project
make build

# 6. Run tests
make test

# 7. Run linter
make lint
```

### Development Workflow

```bash
# 1. Create a feature branch
git checkout -b feature/your-feature-name

# 2. Make your changes
# ... edit files ...

# 3. Test your changes
make test
make lint

# 4. Commit your changes
git add .
git commit -m "feat: add your feature description"

# 5. Push to your fork
git push origin feature/your-feature-name

# 6. Create a Pull Request on GitHub
```

## Code Style Guidelines

### Go Code Style

- **Follow `go fmt`** - All code must be formatted with `go fmt`
- **Use `goimports`** - Organize imports automatically
- **Write clear comments** - Explain complex logic and public APIs
- **Keep functions focused** - One responsibility per function
- **Use meaningful names** - Variables and functions should be self-documenting

### Example Code Style

```go
// Good: Clear function with focused responsibility
func validateConfigPath(path string) error {
    if path == "" {
        return errors.New("configuration path cannot be empty")
    }

    if !filepath.IsAbs(path) {
        return fmt.Errorf("configuration path must be absolute: %s", path)
    }

    return nil
}

// Good: Table-driven test
func TestValidateConfigPath(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        wantErr bool
    }{
        {"valid absolute path", "/home/user/config", false},
        {"empty path", "", true},
        {"relative path", "config", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateConfigPath(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateConfigPath() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Testing Guidelines

ConfigSync maintains high test coverage (75%+). When contributing:

### Writing Tests

- **Unit tests** for all new functions
- **Integration tests** for CLI commands
- **Edge case testing** for error conditions
- **Cross-platform testing** on Intel and Apple Silicon

### Test Structure

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/config -v

# Run with coverage
make test-coverage

# Run with race detection
go test -race ./...
```

### Test Naming Convention

```go
func TestFunctionName(t *testing.T)           # Basic test
func TestFunctionName_ErrorCase(t *testing.T) # Error condition
func TestFunctionName_EdgeCase(t *testing.T)  # Edge case
```

## Adding Application Support

One of the most valuable contributions is adding support for new macOS applications.

### Research Phase

Before adding an application, research its configuration:

```bash
# Find the application bundle
find /Applications -name "*.app" -exec basename {} \; | grep -i "appname"

# Check the bundle ID
plutil -p /Applications/AppName.app/Contents/Info.plist | grep CFBundleIdentifier

# Find configuration files
find ~/Library -name "*appname*" -type f 2>/dev/null
find ~/Library -name "*bundleid*" -type f 2>/dev/null
```

### Implementation Steps

1. **Add to detector** (`pkg/apps/detector.go`):

```go
"appname": {
    Name:        "appname",
    DisplayName: "Application Name",
    BundleID:    "com.company.appname",
    Paths: []PathInfo{
        {
            Source:      "~/Library/Preferences/com.company.appname.plist",
            Destination: "Library/Preferences/com.company.appname.plist",
            Type:        config.PathTypeFile,
            Required:    false,
        },
        {
            Source:      "~/Library/Application Support/AppName/",
            Destination: "Library/Application Support/AppName/",
            Type:        config.PathTypeDirectory,
            Required:    true,
        },
    },
},
```

2. **Add tests** (`pkg/apps/detector_test.go`):

```go
func TestDetectAppName(t *testing.T) {
    detector := NewDetector()
    apps := detector.DetectApplications()

    // Test if app is detected when installed
    // Test configuration paths are correct
}
```

3. **Test manually**:

```bash
# Test discovery
configsync discover --filter="appname" --list --verbose

# Test addition
configsync add appname --dry-run

# Test sync
configsync sync appname --dry-run
```

4. **Update documentation**:
- Add to README.md supported applications list
- Add usage examples if needed
- Update this documentation site

## Pull Request Process

### Before Submitting

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated if needed
- [ ] CHANGELOG.md updated for notable changes
- [ ] Commit messages follow conventions

### PR Description Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing completed
- [ ] Tested on Intel Mac
- [ ] Tested on Apple Silicon Mac

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes or breaking changes documented
```

### Review Process

1. **Automated Checks** - GitHub Actions run tests and linting
2. **Code Review** - Maintainers review code and provide feedback
3. **Discussion** - Address any questions or requested changes
4. **Approval** - Once approved, maintainers will merge

## Commit Message Guidelines

We follow conventional commit format for clear history:

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

### Examples

```bash
feat(apps): add support for Sublime Text configuration sync

fix(sync): resolve symlink creation issue on Apple Silicon

docs(readme): update installation instructions for Homebrew

test(backup): add integration tests for backup validation
```

## Release Process

ConfigSync uses automated releases with manual fallbacks. Contributors don't typically handle releases, but here's the overview:

### For Maintainers

1. **Update CHANGELOG.md** with release notes
2. **Create and push tag** (e.g., `v1.2.3`)
3. **Monitor GitHub Actions** for automated release
4. **Handle Homebrew update** if automation fails

See [RELEASE.md](https://github.com/dotbrains/configsync/blob/main/RELEASE.md) for complete details.

## Getting Help

Need help contributing? Here are ways to get support:

### 💬 Discussions
[GitHub Discussions](https://github.com/dotbrains/configsync/discussions) for questions and ideas

### 🐛 Issues
[GitHub Issues](https://github.com/dotbrains/configsync/issues) for bugs and feature requests

### 🔀 Pull Requests
[GitHub PRs](https://github.com/dotbrains/configsync/pulls) for code contributions

### ❓ Questions
Open an issue with "question" label for contribution help

## Recognition

We appreciate all contributions to ConfigSync! Contributors are recognized in:

- **GitHub Contributors** page
- **CHANGELOG.md** for notable contributions
- **README.md** acknowledgments section

## Get Started

- **[View on GitHub](https://github.com/dotbrains/configsync)** - Explore the codebase and open issues
- **[Create Issue](https://github.com/dotbrains/configsync/issues/new)** - Report bugs or suggest features

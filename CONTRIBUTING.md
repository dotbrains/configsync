# Contributing to ConfigSync

Thank you for your interest in contributing to ConfigSync! We welcome contributions from everyone.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the existing issues to see if the problem has already been reported. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed and what behavior you expected**
- **Include your macOS version and ConfigSync version**
- **Add verbose output if applicable** (`--verbose` flag)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Provide specific examples to demonstrate how the enhancement would work**
- **Explain why this enhancement would be useful**

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`make test`)
6. Run the linter (`make lint`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Development Setup

1. **Prerequisites:**
   - Go 1.21 or later
   - macOS (for testing macOS-specific functionality)
   - Git

2. **Clone and setup:**
   ```bash
   git clone https://github.com/dotbrains/configsync.git
   cd configsync
   go mod download
   ```

3. **Build and test:**
   ```bash
   make build
   make test
   make lint
   ```

### Code Style

- Follow standard Go formatting (`go fmt`)
- Write clear, concise comments
- Add tests for new functionality
- Keep functions focused and small
- Use meaningful variable and function names

### Testing

- Write unit tests for new functions
- Test edge cases and error conditions
- Ensure tests pass on both Intel and Apple Silicon Macs
- Use table-driven tests where appropriate

Example test structure:
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "test", "expected", false},
        {"invalid input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NewFeature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewFeature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("NewFeature() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Adding Support for New Applications

To add support for a new macOS application:

1. **Research the application:**
   - Find where it stores configuration files
   - Identify the bundle ID (usually in `/Applications/App.app/Contents/Info.plist`)
   - Test the configuration paths on a real system

2. **Add to the detector:**
   ```go
   // In pkg/apps/detector.go
   "appname": {
       Name:        "appname",
       DisplayName: "App Display Name",
       BundleID:    "com.company.appname",
       Paths: []PathInfo{
           {
               Source:      "~/Library/Preferences/com.company.appname.plist",
               Destination: "Library/Preferences/com.company.appname.plist",
               Type:        config.PathTypeFile,
               Required:    false,
           },
       },
   },
   ```

3. **Add tests:**
   ```go
   func TestDetectNewApp(t *testing.T) {
       // Test the new application detection
   }
   ```

4. **Update documentation:**
   - Add to README.md supported applications list
   - Add example usage if needed

### Commit Messages

Use clear and meaningful commit messages:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Examples:
```
Add support for Sublime Text configuration sync

- Add Sublime Text to known applications list
- Include Packages/User directory detection
- Add tests for Sublime Text detection
- Update documentation

Fixes #123
```

### Documentation

- Update README.md if adding new features
- Add inline comments for complex logic
- Update EXAMPLES.md with new usage patterns
- Update CHANGELOG.md for notable changes

### Release Process

Releases are automated through GitHub Actions when a tag is pushed:

1. Update CHANGELOG.md with release notes
2. Create and push a new tag: `git tag v1.2.3 && git push origin v1.2.3`
3. GitHub Actions will build and create a release automatically

## Questions?

Feel free to open an issue with the "question" label if you have questions about contributing.

Thank you for contributing to ConfigSync! 🎉
# ConfigSync Examples

This document provides practical examples of how to use ConfigSync to manage your macOS application configurations.

## Quick Start

### 1. Initialize ConfigSync

```bash
configsync init
```

This creates the `~/.configsync/` directory with the following structure:
```
~/.configsync/
├── config.yaml              # Main configuration file
├── store/                   # Central storage for all app configs
│   ├── Library/
│   │   ├── Preferences/     # App preference files
│   │   └── Application Support/  # App support files
│   └── .config/            # XDG-style configs
├── backups/                # Backup of original configs
└── logs/                   # Operation logs
```

### 2. Add Applications

List supported applications:
```bash
configsync add --list-supported
```

Add specific applications:
```bash
# Add VS Code
configsync add vscode

# Add multiple apps at once
configsync add git ssh Terminal "Google Chrome"

# Add with verbose output to see what's being managed
configsync add firefox --verbose
```

### 3. Check Status

```bash
# Basic status
configsync status

# Detailed status with path information
configsync status --verbose
```

### 4. Sync Configurations

```bash
# Dry run to see what would happen
configsync sync --dry-run --verbose

# Sync all applications
configsync sync

# Sync specific applications
configsync sync git ssh
```

## Common Workflows

### Setting Up a New Mac

1. Clone your dotfiles repository (if using version control)
2. Install ConfigSync
3. Initialize: `configsync init`
4. Add applications: `configsync add git ssh vscode Terminal`
5. Sync: `configsync sync`

### Adding a New Application

1. Add the application: `configsync add myapp --verbose`
2. Check what was detected: `configsync status --verbose`
3. Test sync with dry-run: `configsync sync myapp --dry-run --verbose`
4. Sync for real: `configsync sync myapp`

### Removing an Application

1. Remove from management: `configsync remove myapp --verbose`
   - This removes symlinks and restores original files
2. Check status: `configsync status`

### Using with Version Control

ConfigSync works great with Git for maintaining your configurations across multiple machines:

```bash
# Initialize ConfigSync
configsync init

# Add your apps
configsync add git vscode Terminal

# Sync them
configsync sync

# Initialize git in the store directory
cd ~/.configsync/store
git init
git add .
git commit -m "Initial configuration sync"

# Push to your remote repository
git remote add origin https://github.com/yourusername/dotfiles.git
git push -u origin main
```

On a new machine:
```bash
# Clone your configurations
git clone https://github.com/yourusername/dotfiles.git ~/.configsync/store

# Initialize ConfigSync and add apps
configsync init
configsync add git vscode Terminal

# Sync (this will use the files from git)
configsync sync
```

## Supported Applications

ConfigSync automatically detects configuration files for these applications:

| Application | Configuration Files |
|------------|-------------------|
| VS Code | settings.json, keybindings.json, snippets/ |
| Git | .gitconfig, .gitignore_global |
| SSH | .ssh/config |
| Terminal | com.apple.Terminal.plist |
| iTerm2 | com.googlecode.iterm2.plist |
| Google Chrome | Preferences, com.google.Chrome.plist |
| Firefox | Profiles, org.mozilla.firefox.plist |
| Sublime Text | Packages/User/ |

## Advanced Usage

### Dry Run Mode

Always test changes first:
```bash
configsync sync --dry-run --verbose
configsync remove myapp --dry-run --verbose
```

### Verbose Output

Get detailed information about operations:
```bash
configsync add vscode --verbose
configsync sync --verbose
configsync status --verbose
```

### Custom Home Directory

Use a different home directory:
```bash
configsync --home /Users/other sync
```

## Troubleshooting

### Application Not Detected

If ConfigSync can't detect an application:

1. Check if it's in the supported list: `configsync add --list-supported`
2. Try with verbose output: `configsync add myapp --verbose`
3. The app might store configs in non-standard locations
4. Create a GitHub issue with details about the application

### Sync Issues

If syncing fails:

1. Run with verbose output: `configsync sync myapp --verbose`
2. Check file permissions
3. Ensure the application isn't currently running
4. Try a dry-run first: `configsync sync myapp --dry-run --verbose`

### Removing Symlinks

To completely remove ConfigSync and restore original files:

1. Remove all apps: `configsync remove app1 app2 app3`
2. Or restore manually by copying files from `~/.configsync/store/` back to their original locations
3. Delete `~/.configsync/` directory

## Tips and Best Practices

1. **Always use dry-run first** when trying new operations
2. **Use version control** for your `~/.configsync/store/` directory
3. **Back up important configs** before first sync
4. **Test on a non-critical machine first** when setting up
5. **Use verbose mode** when troubleshooting
6. **Check status regularly** to ensure everything stays in sync
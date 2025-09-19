---
layout: default
title: CLI Reference
---

<div class="content">
    <h1>CLI Reference</h1>
    <p class="section-subtitle">
        Complete command-line interface documentation for ConfigSync
    </p>

    ## Global Options

    These options are available for all ConfigSync commands:

    ```bash
    --config string    Path to config file (default: ~/.configsync/config.yaml)
    --verbose         Enable verbose output
    --quiet           Suppress non-essential output
    --help            Show help for any command
    --version         Show version information
    ```

    ## Core Commands

    ### `configsync init`

    Initialize ConfigSync in the current user directory.

    **Usage:**
    ```bash
    configsync init [flags]
    ```

    **Flags:**
    ```bash
    --force       Overwrite existing ConfigSync installation
    --dry-run     Show what would be created without making changes
    ```

    **Examples:**
    ```bash
    # Initialize ConfigSync
    configsync init

    # Initialize with force (overwrite existing)
    configsync init --force

    # Preview initialization without making changes
    configsync init --dry-run
    ```

    **What it does:**
    - Creates `~/.configsync/` directory structure
    - Initializes `config.yaml` with default settings
    - Creates subdirectories for store, backups, logs, and temp files
    - Sets up logging configuration

    ---

    ### `configsync add <app>`

    Add an application's configuration to ConfigSync management.

    **Usage:**
    ```bash
    configsync add <app> [app2] [app3] ... [flags]
    ```

    **Flags:**
    ```bash
    --config-path string   Custom configuration path for the application
    --force               Add application even if already managed
    --dry-run             Preview addition without making changes
    ```

    **Examples:**
    ```bash
    # Add single application
    configsync add vscode

    # Add multiple applications
    configsync add vscode chrome firefox

    # Add with custom configuration path
    configsync add myapp --config-path="~/Library/Preferences/com.myapp.plist"

    # Preview addition
    configsync add vscode --dry-run
    ```

    **Supported application names:**
    - `vscode` - Visual Studio Code
    - `chrome` - Google Chrome
    - `firefox` - Firefox
    - `iterm2` - iTerm2
    - `terminal` - macOS Terminal
    - `git` - Git configuration
    - `ssh` - SSH configuration
    - And many more...

    ---

    ### `configsync remove <app>`

    Remove an application from ConfigSync management and restore original configurations.

    **Usage:**
    ```bash
    configsync remove <app> [app2] [app3] ... [flags]
    ```

    **Flags:**
    ```bash
    --keep-backup     Keep backup files after removal
    --force           Remove even if restoration fails
    --dry-run         Preview removal without making changes
    ```

    **Examples:**
    ```bash
    # Remove single application
    configsync remove vscode

    # Remove multiple applications
    configsync remove vscode chrome

    # Remove and keep backups
    configsync remove vscode --keep-backup

    # Preview removal
    configsync remove vscode --dry-run
    ```

    ---

    ### `configsync sync`

    Sync all or specific configurations (create/update symlinks).

    **Usage:**
    ```bash
    configsync sync [app1] [app2] ... [flags]
    ```

    **Flags:**
    ```bash
    --dry-run            Preview sync operations without making changes
    --force              Override conflicts and force sync
    --rebuild-links      Rebuild all symlinks from scratch
    --include string     Include only files matching pattern (glob)
    --exclude string     Exclude files matching pattern (glob)
    --check-integrity    Verify symlink integrity after sync
    ```

    **Examples:**
    ```bash
    # Sync all applications
    configsync sync

    # Sync specific applications
    configsync sync vscode chrome

    # Preview sync operations
    configsync sync --dry-run

    # Force sync (override conflicts)
    configsync sync --force

    # Rebuild all symlinks
    configsync sync --rebuild-links

    # Sync only JSON and plist files
    configsync sync --include="*.json,*.plist"

    # Sync excluding cache files
    configsync sync --exclude="cache/*,logs/*"
    ```

    ---

    ### `configsync status`

    Show detailed status of all managed configurations.

    **Usage:**
    ```bash
    configsync status [flags]
    ```

    **Flags:**
    ```bash
    --verbose           Show detailed path information
    --check-integrity   Verify symlink integrity
    --format string     Output format: table, json, yaml (default: table)
    ```

    **Examples:**
    ```bash
    # Show basic status
    configsync status

    # Show detailed status with paths
    configsync status --verbose

    # Check and report symlink integrity
    configsync status --check-integrity

    # Output as JSON
    configsync status --format=json
    ```

    ## Discovery Commands

    ### `configsync discover`

    Automatically discover installed applications and their configurations.

    **Usage:**
    ```bash
    configsync discover [flags]
    ```

    **Flags:**
    ```bash
    --list              List discovered applications in table format
    --auto-add          Automatically add all discovered applications
    --filter string     Filter results to specific applications (comma-separated)
    --verbose           Show detailed configuration paths
    --dry-run           Preview operations without making changes
    ```

    **Examples:**
    ```bash
    # Discover all applications
    configsync discover

    # List discovered applications in table format
    configsync discover --list

    # Show detailed paths for discovered apps
    configsync discover --list --verbose

    # Auto-add all discovered applications
    configsync discover --auto-add

    # Preview auto-add operations
    configsync discover --auto-add --dry-run

    # Filter discovery to specific apps
    configsync discover --filter="vscode,chrome,firefox"

    # Filter and auto-add specific apps
    configsync discover --filter="vscode,chrome" --auto-add
    ```

    ## Backup & Restore Commands

    ### `configsync backup`

    Create backups of configurations with checksum validation.

    **Usage:**
    ```bash
    configsync backup [app1] [app2] ... [flags]
    ```

    **Flags:**
    ```bash
    --validate         Validate integrity of existing backups
    --keep-days int    Clean up backups older than specified days
    --compress         Compress backup files to save space
    ```

    **Examples:**
    ```bash
    # Backup all applications
    configsync backup

    # Backup specific applications
    configsync backup vscode chrome

    # Validate existing backups
    configsync backup --validate

    # Clean up backups older than 30 days
    configsync backup --keep-days 30

    # Create compressed backups
    configsync backup --compress
    ```

    ---

    ### `configsync restore`

    Restore original configurations from backups.

    **Usage:**
    ```bash
    configsync restore <app> [flags]
    ```

    **Flags:**
    ```bash
    --all                 Restore all applications with backups
    --list                List available backups
    --backup-date string  Restore from specific backup date (YYYY-MM-DD)
    --force              Restore even if current config would be overwritten
    ```

    **Examples:**
    ```bash
    # Restore specific application
    configsync restore vscode

    # Restore all applications
    configsync restore --all

    # List available backups
    configsync restore --list

    # Restore from specific backup date
    configsync restore vscode --backup-date=2024-01-15

    # Force restore (overwrite current config)
    configsync restore vscode --force
    ```

    ## Deployment Commands

    ### `configsync export`

    Export configuration bundle for deployment to other systems.

    **Usage:**
    ```bash
    configsync export [flags]
    ```

    **Flags:**
    ```bash
    --output string     Output file path (default: configsync-export-{timestamp}.tar.gz)
    --apps string       Export only specific applications (comma-separated)
    --compress-level    Compression level 1-9 (default: 6)
    ```

    **Examples:**
    ```bash
    # Export all configurations
    configsync export

    # Export to specific file
    configsync export --output my-config.tar.gz

    # Export only specific applications
    configsync export --apps vscode,git,ssh

    # Export with custom output path
    configsync export --output ~/Desktop/my-setup.tar.gz
    ```

    ---

    ### `configsync import`

    Import configuration bundle from another system.

    **Usage:**
    ```bash
    configsync import <bundle> [flags]
    ```

    **Flags:**
    ```bash
    --force             Force import even with conflicts
    --preview           Show what would be imported without making changes
    --validate-only     Only validate bundle integrity without importing
    ```

    **Examples:**
    ```bash
    # Import configuration bundle
    configsync import ~/Desktop/my-config.tar.gz

    # Force import (override conflicts)
    configsync import --force ~/Desktop/my-config.tar.gz

    # Preview import operations
    configsync import --preview ~/Desktop/my-config.tar.gz

    # Validate bundle without importing
    configsync import --validate-only ~/Desktop/my-config.tar.gz
    ```

    ---

    ### `configsync deploy`

    Deploy imported configurations to the current system.

    **Usage:**
    ```bash
    configsync deploy [flags]
    ```

    **Flags:**
    ```bash
    --force             Force deployment overriding conflicts
    --dry-run          Preview deployment without making changes
    --apps string      Deploy only specific applications (comma-separated)
    ```

    **Examples:**
    ```bash
    # Deploy imported configurations
    configsync deploy

    # Force deployment (override conflicts)
    configsync deploy --force

    # Preview deployment
    configsync deploy --dry-run

    # Deploy only specific applications
    configsync deploy --apps vscode,chrome
    ```

    ## Utility Commands

    ### `configsync completion`

    Generate shell completion scripts for bash, zsh, or fish.

    **Usage:**
    ```bash
    configsync completion <shell>
    ```

    **Supported shells:**
    - `bash`
    - `zsh`
    - `fish`

    **Examples:**
    ```bash
    # Generate bash completion
    configsync completion bash > /usr/local/etc/bash_completion.d/configsync

    # Generate zsh completion
    configsync completion zsh > ~/.oh-my-zsh/completions/_configsync

    # Generate fish completion
    configsync completion fish > ~/.config/fish/completions/configsync.fish
    ```

    ---

    ### `configsync help`

    Show help information for ConfigSync commands.

    **Usage:**
    ```bash
    configsync help [command]
    ```

    **Examples:**
    ```bash
    # Show general help
    configsync help

    # Show help for specific command
    configsync help sync
    configsync help discover
    configsync help backup
    ```

    ## Exit Codes

    ConfigSync uses standard exit codes to indicate command results:

    - `0` - Success
    - `1` - General error
    - `2` - Configuration error
    - `3` - Permission error
    - `4` - File/directory not found
    - `5` - Backup/restore error
    - `6` - Sync/symlink error
    - `7` - Network/download error

    ## Configuration File

    The main configuration file is located at `~/.configsync/config.yaml`:

    ```yaml
    # ConfigSync Configuration
    version: "1.0"
    store_path: ~/.configsync/store
    backup_enabled: true
    logging:
      level: info
      file: ~/.configsync/logs/configsync.log

    applications:
      vscode:
        name: "Visual Studio Code"
        enabled: true
        paths:
          - source: "~/Library/Application Support/Code/User/settings.json"
            destination: "Library/Application Support/Code/User/settings.json"
            type: file
          - source: "~/Library/Application Support/Code/User/keybindings.json"
            destination: "Library/Application Support/Code/User/keybindings.json"
            type: file
        last_sync: "2024-01-15T14:30:45Z"
    ```

    ## Environment Variables

    ConfigSync respects the following environment variables:

    - `CONFIGSYNC_HOME` - Override default ConfigSync directory
    - `CONFIGSYNC_CONFIG` - Override default config file path
    - `CONFIGSYNC_LOG_LEVEL` - Set log level (debug, info, warn, error)
    - `CONFIGSYNC_BACKUP_ENABLED` - Enable/disable automatic backups
    - `NO_COLOR` - Disable colored output

    **Examples:**
    ```bash
    # Use custom ConfigSync directory
    export CONFIGSYNC_HOME=~/my-configsync
    configsync init

    # Enable debug logging
    export CONFIGSYNC_LOG_LEVEL=debug
    configsync sync --verbose

    # Disable colored output
    export NO_COLOR=1
    configsync status
    ```

    ## Tips & Best Practices

    ### Command Chaining

    ```bash
    # Complete setup in one line
    configsync init && configsync discover --auto-add && configsync sync

    # Backup and sync workflow
    configsync backup --validate && configsync sync && configsync status
    ```

    ### Using Dry Run

    Always test operations with `--dry-run` first:

    ```bash
    # Preview before executing
    configsync sync --dry-run
    configsync discover --auto-add --dry-run
    configsync deploy --dry-run
    ```

    ### Filtering and Selection

    Use filters to work with specific applications:

    ```bash
    # Work with development tools only
    configsync discover --filter="*code*,*term*,git" --auto-add
    configsync sync --include="*.json,*.yaml"

    # Work with browsers only
    configsync discover --filter="chrome,firefox,safari" --list
    ```

    ### Automation and Scripting

    ConfigSync is designed to work well in scripts:

    ```bash
    #!/bin/bash
    # Daily sync script
    configsync backup --keep-days 30
    configsync sync --check-integrity
    configsync status --format=json > /tmp/configsync-status.json
    ```

    <div class="text-center mt-4">
        <a href="{{ '/getting-started/' | relative_url }}" class="btn btn-primary">
            <i class="fas fa-play"></i>
            Getting Started Guide
        </a>
        <a href="https://github.com/{{ site.repository }}/issues" class="btn btn-secondary" target="_blank">
            <i class="fas fa-bug"></i>
            Report Issues
        </a>
    </div>
</div>

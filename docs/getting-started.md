---
layout: default
title: Getting Started
permalink: /getting-started/
---

# Getting Started with ConfigSync

*A step-by-step guide to start managing your macOS application configurations*

## Prerequisites

    Before getting started, make sure you have:

    - macOS 10.15 (Catalina) or later
    - ConfigSync installed ([Installation Guide]({{ '/installation/' | relative_url }}))
    - Applications you want to manage (VS Code, Chrome, etc.)

    ## Step 1: Initialize ConfigSync

    Start by initializing ConfigSync in your home directory:

    ```bash
    # Initialize ConfigSync
    configsync init
    ```

    This creates the following directory structure:

    ```
    ~/.configsync/
    ├── config.yaml              # Main configuration registry
    ├── store/                   # Central storage for configurations
    ├── backups/                 # Automated backups
    ├── logs/                    # Operation logs
    └── temp/                    # Temporary files for operations
    ```

    ## Step 2: Discover Applications

    ConfigSync can automatically discover installed applications and their configuration files:

    ```bash
    # Discover all installed applications
    configsync discover

    # View discovered applications in a table format
    configsync discover --list

    # Discover with detailed path information
    configsync discover --list --verbose
    ```

    **Sample Output:**
    ```
    ┌─────────────────┬──────────────────────────────────────┬────────────────┐
    │ Application     │ Bundle ID                            │ Configurations │
    ├─────────────────┼──────────────────────────────────────┼────────────────┤
    │ Visual Studio   │ com.microsoft.VSCode                 │ 3 files        │
    │ Code            │                                      │                │
    ├─────────────────┼──────────────────────────────────────┼────────────────┤
    │ Google Chrome   │ com.google.Chrome                    │ 2 files        │
    ├─────────────────┼──────────────────────────────────────┼────────────────┤
    │ iTerm2          │ com.googlecode.iterm2                │ 1 file         │
    └─────────────────┴──────────────────────────────────────┴────────────────┘
    ```

    ## Step 3: Add Applications

    You can add applications to ConfigSync management in several ways:

    ### Automatic Addition (Recommended)

    ```bash
    # Auto-add all discovered applications
    configsync discover --auto-add

    # Auto-add with preview (dry-run)
    configsync discover --auto-add --dry-run

    # Auto-add specific applications only
    configsync discover --filter="vscode,chrome" --auto-add
    ```

    ### Manual Addition

    ```bash
    # Add specific applications
    configsync add vscode
    configsync add "Google Chrome"
    configsync add firefox terminal

    # Add multiple applications at once
    configsync add vscode chrome firefox
    ```

    ## Step 4: Check Status

    After adding applications, check their status:

    ```bash
    configsync status
    ```

    **Sample Output:**
    ```
    ConfigSync Status Report
    ========================

    Configuration: ~/.configsync/config.yaml
    Store Location: ~/.configsync/store/

    Applications (3 managed):

    ✅ Visual Studio Code (vscode)
       Settings: ~/.configsync/store/Library/Application Support/Code/User/
       Status: Synced (3 files)
       Last Sync: 2024-01-15 14:30:45

    ⚠️  Google Chrome (chrome)
       Settings: ~/.configsync/store/Library/Application Support/Google/Chrome/
       Status: Needs Sync (configuration changed)
       Last Sync: Never

    🔗 iTerm2 (iterm2)
       Settings: ~/.configsync/store/Library/Preferences/
       Status: Linked (1 symlink active)
       Last Sync: 2024-01-15 12:15:30
    ```

    ## Step 5: Sync Configurations

    Sync your application configurations to the central store:

    ```bash
    # Sync all applications
    configsync sync

    # Preview sync operations (dry-run)
    configsync sync --dry-run

    # Sync specific applications only
    configsync sync vscode chrome

    # Force sync (override any conflicts)
    configsync sync --force
    ```

    The sync process will:
    1. Create backups of existing configurations
    2. Copy configurations to the central store
    3. Create symlinks from app locations to the store
    4. Verify symlink integrity

    ## Step 6: Backup Management

    ConfigSync automatically creates backups, but you can also manage them manually:

    ```bash
    # Create manual backup
    configsync backup

    # Backup specific applications
    configsync backup vscode chrome

    # Validate existing backups
    configsync backup --validate

    # Clean up old backups (older than 30 days)
    configsync backup --keep-days 30

    # List available backups
    configsync restore --list
    ```

    ## Working with Multiple Macs

    ConfigSync makes it easy to deploy your configurations to multiple Mac systems.

    ### Exporting Configurations

    On your source Mac:

    ```bash
    # Export all configurations
    configsync export --output my-configs.tar.gz

    # Export specific applications only
    configsync export --output dev-tools.tar.gz --apps "vscode,git,ssh"

    # Export to a specific directory
    configsync export --output ~/Desktop/my-setup.tar.gz
    ```

    ### Importing on New Mac

    On your target Mac:

    ```bash
    # Initialize ConfigSync
    configsync init

    # Import the configuration bundle
    configsync import ~/Desktop/my-configs.tar.gz

    # Deploy the configurations
    configsync deploy

    # Or deploy with force (if conflicts exist)
    configsync deploy --force
    ```

    ## Advanced Usage

    ### Filter Discovery Results

    ```bash
    # Discover only development tools
    configsync discover --filter="vscode,sublime,iterm"

    # Discover browsers only
    configsync discover --filter="chrome,firefox,safari"

    # Use patterns in filters
    configsync discover --filter="*code*,*term*"
    ```

    ### Custom Configuration Paths

    You can manually specify configuration paths for applications not automatically detected:

    ```bash
    # Add application with custom paths
    configsync add myapp --config-path="~/Library/Preferences/com.myapp.plist"
    ```

    ### Selective Sync

    ```bash
    # Sync only specific file types
    configsync sync --include="*.json,*.plist"

    # Exclude certain files from sync
    configsync sync --exclude="cache/*,logs/*"
    ```

    ## Common Workflows

    ### Daily Development Setup

    ```bash
    # Morning routine: sync latest changes
    configsync sync
    configsync status

    # Check for new applications
    configsync discover --auto-add --dry-run
    ```

    ### Setting Up a New Mac

    ```bash
    # On new Mac
    configsync init
    configsync import ~/path/to/backup.tar.gz
    configsync deploy --force

    # Verify everything is working
    configsync status
    ```

    ### Before Major Changes

    ```bash
    # Create full backup before updates
    configsync backup --validate

    # Make changes to your applications
    # ...

    # Sync changes to central store
    configsync sync
    ```

    ## Troubleshooting

    ### Sync Issues

    If sync operations fail:

    ```bash
    # Check detailed status
    configsync status --verbose

    # View recent logs
    tail -f ~/.configsync/logs/configsync.log

    # Reset and re-sync specific app
    configsync remove appname
    configsync add appname
    configsync sync appname
    ```

    ### Backup Recovery

    If you need to restore from backup:

    ```bash
    # List available backups
    configsync restore --list

    # Restore specific application
    configsync restore vscode

    # Restore all applications
    configsync restore --all

    # Restore from specific backup date
    configsync restore --backup-date=2024-01-15
    ```

    ### Symlink Problems

    If symlinks become broken:

    ```bash
    # Check symlink integrity
    configsync status --check-integrity

    # Rebuild symlinks
    configsync sync --rebuild-links

    # Remove and re-add problematic app
    configsync remove appname
    configsync add appname
    configsync sync appname
    ```

    ## Best Practices

    ### Regular Maintenance

    - Run `configsync sync` regularly to keep configurations up-to-date
    - Use `configsync backup --validate` weekly to ensure backup integrity
    - Clean up old backups monthly with `--keep-days` option
    - Check `configsync status` before major system changes

    ### Version Control Integration

    Consider version controlling your ConfigSync store:

    ```bash
    cd ~/.configsync/store/
    git init
    git add .
    git commit -m "Initial configuration backup"

    # After changes
    configsync sync
    cd ~/.configsync/store/
    git add .
    git commit -m "Updated configurations"
    ```

    ### Security Considerations

    - Be careful with sensitive configuration files
    - Consider excluding files with passwords or tokens
    - Use `.gitignore` if version controlling the store
    - Regularly validate backup checksums

## Next Steps

Now that you have ConfigSync set up, here are some helpful resources:

- **[CLI Reference]({{ '/cli-reference/' | relative_url }})** - Complete command-line documentation
- **[Contributing Guide]({{ '/contributing/' | relative_url }})** - Help improve ConfigSync
- **[GitHub Discussions](https://github.com/{{ site.repository }}/discussions)** - Join the community

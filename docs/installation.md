---
layout: default
title: Installation
permalink: /installation/
---

# Installation Guide

*Choose your preferred installation method to get started with ConfigSync*

## Homebrew (Recommended)

    The easiest way to install ConfigSync is using Homebrew:

    ```bash
    # Add the tap and install
    brew install dotbrains/tap/configsync

    # Verify installation
    configsync --version
    ```

    **Benefits of Homebrew installation:**
    - Automatic dependency management
    - Easy updates with `brew upgrade configsync`
    - Automatic PATH configuration

    ## Pre-built Binaries

    Download pre-built binaries from our [GitHub Releases](https://github.com/dotbrains/configsync/releases) page.

    ### Universal Binary (Recommended)

    Works on both Intel and Apple Silicon Macs:

    ```bash
    # Download and install universal binary
    curl -L https://github.com/dotbrains/configsync/releases/latest/download/configsync-darwin-universal.tar.gz | tar -xz
    sudo mv configsync-darwin-universal /usr/local/bin/configsync
    chmod +x /usr/local/bin/configsync
    ```

    ### Architecture-Specific Binaries

    If you prefer architecture-specific binaries:

    #### Intel Macs (x86_64)
    ```bash
    curl -L https://github.com/dotbrains/configsync/releases/latest/download/configsync-darwin-amd64.tar.gz | tar -xz
    sudo mv configsync-darwin-amd64 /usr/local/bin/configsync
    chmod +x /usr/local/bin/configsync
    ```

    #### Apple Silicon Macs (ARM64)
    ```bash
    curl -L https://github.com/dotbrains/configsync/releases/latest/download/configsync-darwin-arm64.tar.gz | tar -xz
    sudo mv configsync-darwin-arm64 /usr/local/bin/configsync
    chmod +x /usr/local/bin/configsync
    ```

    ## From Source

    Build ConfigSync from source if you want the latest development version:

    ### Prerequisites
    - Go 1.21 or later
    - Git

    ### Installation Steps

    ```bash
    # Install from source (requires Go 1.21+)
    go install github.com/dotbrains/configsync@latest
    ```

    Or build locally:

    ```bash
    # Clone the repository
    git clone https://github.com/dotbrains/configsync.git
    cd configsync

    # Install dependencies
    go mod download

    # Build the project
    make build

    # Install to system PATH
    sudo cp configsync /usr/local/bin/
    ```

    ## Verification

    After installation, verify that ConfigSync is working correctly:

    ```bash
    # Check version
    configsync --version

    # View help
    configsync --help

    # Test basic functionality
    configsync init --dry-run
    ```

    ## Shell Completion (Optional)

    ConfigSync supports shell completion for bash, zsh, and fish. Set up completion for your shell:

    ### Bash
    ```bash
    # Install completion script
    configsync completion bash > /usr/local/etc/bash_completion.d/configsync

    # Reload your shell or source the completion
    source /usr/local/etc/bash_completion.d/configsync
    ```

    ### Zsh (oh-my-zsh)
    ```bash
    # Create completions directory if it doesn't exist
    mkdir -p ~/.oh-my-zsh/completions

    # Install completion script
    configsync completion zsh > ~/.oh-my-zsh/completions/_configsync

    # Reload your shell
    exec zsh
    ```

    ### Fish
    ```bash
    # Install completion script
    configsync completion fish > ~/.config/fish/completions/configsync.fish

    # Reload fish completions
    fish -c "source ~/.config/fish/completions/configsync.fish"
    ```

    ## System Requirements

    - **Operating System:** macOS 10.15 (Catalina) or later
    - **Architecture:** Intel (x86_64) or Apple Silicon (ARM64)
    - **Disk Space:** ~10MB for binary, additional space for configuration storage
    - **Permissions:** Read/write access to home directory and `~/Library/` folders

    ## Troubleshooting

    ### Permission Denied Errors

    If you get permission denied errors when installing:

    ```bash
    # Make sure the binary is executable
    chmod +x /usr/local/bin/configsync

    # If /usr/local/bin doesn't exist, create it
    sudo mkdir -p /usr/local/bin

    # Make sure /usr/local/bin is in your PATH
    echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
    source ~/.zshrc
    ```

    ### Homebrew Installation Issues

    If Homebrew installation fails:

    ```bash
    # Update Homebrew
    brew update

    # Try installing again
    brew install dotbrains/tap/configsync

    # If the tap doesn't exist, add it manually
    brew tap dotbrains/tap
    brew install configsync
    ```

    ### Binary Not Found

    If you get "command not found" errors:

    1. Verify the binary is in your PATH:
       ```bash
       which configsync
       echo $PATH
       ```

    2. Add the installation directory to your PATH:
       ```bash
       # For zsh (default on macOS)
       echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
       source ~/.zshrc

       # For bash
       echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bash_profile
       source ~/.bash_profile
       ```

    ## Uninstalling

    ### Homebrew
    ```bash
    brew uninstall configsync
    brew untap dotbrains/tap  # Optional: remove the tap
    ```

    ### Manual Installation
    ```bash
    # Remove the binary
    sudo rm /usr/local/bin/configsync

    # Remove shell completion (optional)
    rm /usr/local/etc/bash_completion.d/configsync
    rm ~/.oh-my-zsh/completions/_configsync
    rm ~/.config/fish/completions/configsync.fish

    # Remove ConfigSync data (optional)
    # WARNING: This will delete all your configurations and backups
    rm -rf ~/.configsync
    ```

## Next Steps

Once ConfigSync is installed, you're ready to start managing your configurations:

- **[Getting Started Guide]({{ '/getting-started/' | relative_url }})** - Learn how to set up and use ConfigSync
- **[CLI Reference]({{ '/cli-reference/' | relative_url }})** - Complete command documentation

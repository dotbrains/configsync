---
layout: default
title: Documentation
permalink: /docs/
---

<div class="content">
    <h1>ConfigSync Documentation</h1>
    <p class="section-subtitle">
        Complete documentation for ConfigSync - the macOS configuration management tool
    </p>

    ## Getting Started

    New to ConfigSync? Start here to get up and running quickly.

    <div class="features-grid" style="margin: 2rem 0;">
        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-download"></i>
            </div>
            <h3><a href="{{ '/installation/' | relative_url }}">Installation Guide</a></h3>
            <p>Step-by-step instructions for installing ConfigSync using Homebrew, pre-built binaries, or from source.</p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-play"></i>
            </div>
            <h3><a href="{{ '/getting-started/' | relative_url }}">Getting Started</a></h3>
            <p>Complete walkthrough of your first ConfigSync setup, from initialization to syncing configurations.</p>
        </div>
    </div>

    ## Core Documentation

    Deep dive into ConfigSync's features and capabilities.

    <div class="features-grid" style="margin: 2rem 0;">
        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-terminal"></i>
            </div>
            <h3><a href="{{ '/cli-reference/' | relative_url }}">CLI Reference</a></h3>
            <p>Complete command-line interface documentation with examples, flags, and usage patterns.</p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-cogs"></i>
            </div>
            <h3>Configuration</h3>
            <p>Learn how to customize ConfigSync behavior, manage application paths, and optimize your setup.</p>
            <p><em>Coming soon...</em></p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-apps"></i>
            </div>
            <h3>Supported Applications</h3>
            <p>Browse the complete list of supported macOS applications and their configuration details.</p>
            <p><em>Coming soon...</em></p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-shield-alt"></i>
            </div>
            <h3>Backup & Recovery</h3>
            <p>Learn about ConfigSync's safety features, backup validation, and disaster recovery procedures.</p>
            <p><em>Coming soon...</em></p>
        </div>
    </div>

    ## Advanced Topics

    Master ConfigSync's advanced features and workflows.

    <div class="features-grid" style="margin: 2rem 0;">
        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-rocket"></i>
            </div>
            <h3>Deployment Workflows</h3>
            <p>Best practices for deploying configurations across multiple Mac systems and team environments.</p>
            <p><em>Coming soon...</em></p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-code"></i>
            </div>
            <h3>Automation & Scripting</h3>
            <p>Integrate ConfigSync into your automation workflows, CI/CD pipelines, and system provisioning.</p>
            <p><em>Coming soon...</em></p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-search"></i>
            </div>
            <h3>Troubleshooting</h3>
            <p>Common issues, debugging techniques, and solutions for ConfigSync problems.</p>
            <p><em>Coming soon...</em></p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-magic"></i>
            </div>
            <h3>Tips & Tricks</h3>
            <p>Power user techniques, performance optimization, and workflow enhancements.</p>
            <p><em>Coming soon...</em></p>
        </div>
    </div>

    ## Community & Contributing

    Join the ConfigSync community and contribute to the project.

    <div class="features-grid" style="margin: 2rem 0;">
        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-code-branch"></i>
            </div>
            <h3><a href="{{ '/contributing/' | relative_url }}">Contributing Guide</a></h3>
            <p>Learn how to contribute code, documentation, or application support to ConfigSync.</p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-comments"></i>
            </div>
            <h3><a href="https://github.com/dotbrains/configsync/discussions" target="_blank">Community Discussions</a></h3>
            <p>Join discussions, ask questions, and share your ConfigSync experiences with other users.</p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-bug"></i>
            </div>
            <h3><a href="https://github.com/dotbrains/configsync/issues" target="_blank">Report Issues</a></h3>
            <p>Found a bug or have a feature request? Report it on GitHub to help improve ConfigSync.</p>
        </div>

        <div class="feature-card">
            <div class="feature-icon">
                <i class="fas fa-heart"></i>
            </div>
            <h3>Support the Project</h3>
            <p>Star the project on GitHub, share it with others, and help spread the word about ConfigSync.</p>
        </div>
    </div>

    ## Quick Reference

    ### Essential Commands

    ```bash
    # Initialize ConfigSync
    configsync init

    # Auto-discover and add applications
    configsync discover --auto-add

    # Sync all configurations
    configsync sync

    # Check status of managed apps
    configsync status

    # Create backup
    configsync backup

    # Export for deployment
    configsync export --output my-configs.tar.gz
    ```

    ### Useful Links

    - **GitHub Repository**: [dotbrains/configsync](https://github.com/dotbrains/configsync)
    - **Latest Release**: [Download](https://github.com/dotbrains/configsync/releases/latest)
    - **Homebrew Formula**: `brew install dotbrains/tap/configsync`
    - **Issues & Bugs**: [Report](https://github.com/dotbrains/configsync/issues)
    - **Feature Requests**: [Suggest](https://github.com/dotbrains/configsync/issues/new)

    ### System Requirements

    - **OS**: macOS 10.15 (Catalina) or later
    - **Architecture**: Intel (x86_64) or Apple Silicon (ARM64)
    - **Disk Space**: ~10MB for binary + configuration storage
    - **Permissions**: Read/write access to `~/Library/` directories

    ## Need Help?

    Can't find what you're looking for? Here are additional ways to get help:

    1. **Search the documentation** using your browser's find function (Cmd+F)
    2. **Check GitHub Issues** for similar problems and solutions
    3. **Ask in Discussions** for community support and advice
    4. **Read the source code** for implementation details

    <div class="text-center mt-4">
        <a href="{{ '/getting-started/' | relative_url }}" class="btn btn-primary">
            <i class="fas fa-play"></i>
            Get Started Now
        </a>
        <a href="https://github.com/dotbrains/configsync" class="btn btn-secondary" target="_blank">
            <i class="fab fa-github"></i>
            View on GitHub
        </a>
    </div>

    ---

    **Last Updated**: {{ site.time | date: '%B %d, %Y' }}

    Found an error in the documentation? [Edit this page on GitHub](https://github.com/dotbrains/configsync/edit/main/docs/{{ page.path }}) or [report an issue](https://github.com/dotbrains/configsync/issues/new).
</div>

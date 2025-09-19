---
layout: default
title: Home
---

<section class="hero">
    <div class="hero-container">
        <h1>ConfigSync</h1>
        <p>A powerful command-line tool for managing macOS application settings and configurations with centralized storage and syncing across multiple Mac systems.</p>
        <div class="hero-buttons">
            <a href="{{ '/installation/' | relative_url }}" class="btn btn-primary">
                <i class="fas fa-download"></i>
                Get Started
            </a>
            <a href="https://github.com/{{ site.repository }}" class="btn btn-secondary" target="_blank">
                <i class="fab fa-github"></i>
                View on GitHub
            </a>
        </div>
    </div>
</section>

<section class="features">
    <div class="features-container">
        <h2 class="section-title">Why ConfigSync?</h2>
        <p class="section-subtitle">
            Simplify your Mac setup and maintenance with powerful configuration management features
        </p>

        <div class="features-grid">
            <div class="feature-card">
                <div class="feature-icon">
                    <i class="fas fa-sync-alt"></i>
                </div>
                <h3>Centralized Management</h3>
                <p>Store all your app configurations in one central location with organized directory structure that mirrors macOS system paths.</p>
            </div>

            <div class="feature-card">
                <div class="feature-icon">
                    <i class="fas fa-link"></i>
                </div>
                <h3>Smart Symlinks</h3>
                <p>Uses intelligent symlink management with integrity checks to safely sync settings between central storage and application locations.</p>
            </div>

            <div class="feature-card">
                <div class="feature-icon">
                    <i class="fas fa-shield-alt"></i>
                </div>
                <h3>Safe Backups</h3>
                <p>Automatically creates and validates backups before making changes, with checksum verification and easy restoration capabilities.</p>
            </div>

            <div class="feature-card">
                <div class="feature-icon">
                    <i class="fas fa-search"></i>
                </div>
                <h3>Auto Discovery</h3>
                <p>Automatically detects installed applications and their configuration files using multiple scanning methods including Spotlight and system profiler.</p>
            </div>

            <div class="feature-card">
                <div class="feature-icon">
                    <i class="fas fa-rocket"></i>
                </div>
                <h3>Easy Deployment</h3>
                <p>Export configuration bundles and deploy them to new Mac systems with conflict detection and force deployment options.</p>
            </div>

            <div class="feature-card">
                <div class="feature-icon">
                    <i class="fas fa-code"></i>
                </div>
                <h3>Developer Friendly</h3>
                <p>Comprehensive CLI with shell completion for bash, zsh, and fish. Perfect for automation and scripting workflows.</p>
            </div>
        </div>
    </div>
</section>

<section class="content">
    <div style="max-width: 1200px; margin: 0 auto;">
        <h2 class="text-center">Supported Applications</h2>
        <p class="text-center section-subtitle" style="margin-bottom: 3rem;">
            ConfigSync works with 20+ popular macOS applications out of the box, plus automatic detection for any app
        </p>

        <div class="features-grid">
            <div class="feature-card">
                <h4><i class="fas fa-code" style="margin-right: 0.5rem; color: var(--primary-color);"></i>Development Tools</h4>
                <p>Visual Studio Code, Sublime Text, iTerm2, Terminal, Git, SSH, Homebrew</p>
            </div>

            <div class="feature-card">
                <h4><i class="fas fa-globe" style="margin-right: 0.5rem; color: var(--primary-color);"></i>Browsers</h4>
                <p>Google Chrome, Firefox with preferences and user data</p>
            </div>

            <div class="feature-card">
                <h4><i class="fas fa-tools" style="margin-right: 0.5rem; color: var(--primary-color);"></i>Utilities</h4>
                <p>Bartender 4, Rectangle, Magnet, Alfred, CleanMyMac X, 1Password</p>
            </div>

            <div class="feature-card">
                <h4><i class="fas fa-comments" style="margin-right: 0.5rem; color: var(--primary-color);"></i>Communication</h4>
                <p>Slack, Discord, Spotify with workspace and preferences</p>
            </div>
        </div>
    </div>
</section>

</section>

<div class="py-4" style="background: var(--bg-secondary);">
    <div class="content">
        <h2 class="text-center">Quick Start</h2>
        <p class="text-center section-subtitle">
            Get up and running with ConfigSync in minutes
        </p>
    </div>
</div>

```bash
# Install ConfigSync
brew install dotbrains/tap/configsync

# Initialize ConfigSync
configsync init

# Auto-discover and add applications
configsync discover --auto-add

# Sync all configurations
configsync sync

# Export for deployment to another Mac
configsync export --output my-configs.tar.gz
```

<div class="text-center mt-4">
    <a href="{{ '/getting-started/' | relative_url }}" class="btn btn-primary">
        <i class="fas fa-play"></i>
        View Complete Guide
    </a>
</div>

<section class="content">

<section class="content">
    <div style="max-width: 1200px; margin: 0 auto;">
        <h2 class="text-center">Quality & Safety First</h2>
        <p class="text-center section-subtitle">
            Built with reliability and safety in mind
        </p>

        <div class="features-grid">
            <div class="feature-card">
                <div class="feature-icon" style="background: linear-gradient(135deg, #10b981, #06d6a0);">
                    <i class="fas fa-check-circle"></i>
                </div>
                <h3>75%+ Test Coverage</h3>
                <p>Comprehensive test suites including unit tests, integration tests, and benchmarks across all core modules.</p>
            </div>

            <div class="feature-card">
                <div class="feature-icon" style="background: linear-gradient(135deg, #f59e0b, #f97316);">
                    <i class="fas fa-shield-alt"></i>
                </div>
                <h3>Safety Features</h3>
                <p>Automatic backups, conflict detection, dry-run mode, rollback support, and comprehensive operation logging.</p>
            </div>

            <div class="feature-card">
                <div class="feature-icon" style="background: linear-gradient(135deg, #8b5cf6, #a855f7);">
                    <i class="fas fa-memory"></i>
                </div>
                <h3>Performance Tested</h3>
                <p>Optimized with smart caching, validated on Intel and Apple Silicon Macs, with benchmark testing for critical operations.</p>
            </div>
        </div>
    </div>
</section>

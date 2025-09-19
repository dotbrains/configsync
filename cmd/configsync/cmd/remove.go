package cmd

import (
	"fmt"

	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/internal/symlink"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [app1] [app2] ...",
	Short: "Remove application(s) from configuration management",
	Long: `Remove one or more applications from ConfigSync management.
This will also remove any symlinks and restore original files.

Examples:
  configsync remove vscode
  configsync remove "Google Chrome" Firefox`,
	RunE: runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	// Require at least one app name
	if len(args) == 0 {
		return fmt.Errorf("at least one application name is required")
	}

	// Create configuration manager
	manager := config.NewManager(homeDir)

	// Check if ConfigSync is initialized
	if !manager.ConfigExists() {
		return fmt.Errorf("ConfigSync is not initialized. Run 'configsync init' first")
	}

	// Load configuration
	cfg, err := manager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create symlink manager for unsyncing
	symlinkManager := symlink.NewManager(homeDir, cfg.StorePath, cfg.BackupPath, dryRun, verbose)

	// Process each application
	var successful []string
	var failed []string

	for _, appName := range args {
		if verbose {
			fmt.Printf("Processing application: %s\n", appName)
		}

		// Check if app exists
		appConfig, exists := cfg.Apps[appName]
		if !exists {
			if verbose {
				fmt.Printf("  ✗ Application %s is not configured\n", appName)
			}
			failed = append(failed, appName)
			continue
		}

		// Unsync the application first
		if verbose {
			fmt.Printf("  Removing symlinks for %s\n", appConfig.DisplayName)
		}
		if err := symlinkManager.UnsyncApp(appConfig); err != nil {
			if verbose {
				fmt.Printf("  ✗ Failed to unsync %s: %v\n", appConfig.DisplayName, err)
			}
			failed = append(failed, appConfig.DisplayName)
			continue
		}

		// Remove from configuration
		if !dryRun {
			if err := manager.RemoveApp(appName); err != nil {
				if verbose {
					fmt.Printf("  ✗ Failed to remove %s from config: %v\n", appConfig.DisplayName, err)
				}
				failed = append(failed, appConfig.DisplayName)
				continue
			}
		}

		if verbose {
			fmt.Printf("  ✓ Successfully removed %s\n", appConfig.DisplayName)
		}
		successful = append(successful, appConfig.DisplayName)
	}

	// Show results
	if len(successful) > 0 {
		verb := "removed"
		if dryRun {
			verb = "would be removed"
		}
		fmt.Printf("✓ Successfully %s %d application(s):\n", verb, len(successful))
		for _, name := range successful {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(failed) > 0 {
		verb := "failed to remove"
		if dryRun {
			verb = "would fail to remove"
		}
		fmt.Printf("\n✗ %s %d application(s):\n", verb, len(failed))
		for _, name := range failed {
			fmt.Printf("  - %s\n", name)
		}

		if len(successful) == 0 {
			return fmt.Errorf("failed to remove any applications")
		}
	}

	if dryRun && len(successful) > 0 {
		fmt.Println("\nRun without --dry-run to apply these changes.")
	}

	return nil
}

func init() {
	// No additional flags needed for remove command
}

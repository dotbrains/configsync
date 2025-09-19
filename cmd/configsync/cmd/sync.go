package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/internal/symlink"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync [app1] [app2] ...",
	Short: "Sync configurations by creating symlinks",
	Long: `Sync application configurations by creating symlinks from their 
original locations to the central store.

If no app names are provided, all managed applications will be synced.

Examples:
  configsync sync              # Sync all apps
  configsync sync vscode       # Sync only VS Code
  configsync sync Terminal iTerm2  # Sync multiple specific apps`,
	RunE: runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
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

	// Get applications to sync
	var appsToSync map[string]*config.AppConfig
	if len(args) == 0 {
		// Sync all apps
		appsToSync = cfg.Apps
		if len(appsToSync) == 0 {
			fmt.Println("No applications configured. Use 'configsync add <app>' to add applications.")
			return nil
		}
		if verbose {
			fmt.Printf("Syncing all %d configured applications...\n", len(appsToSync))
		}
	} else {
		// Sync specific apps
		appsToSync = make(map[string]*config.AppConfig)
		for _, appName := range args {
			if app, exists := cfg.Apps[appName]; exists {
				appsToSync[appName] = app
			} else {
				return fmt.Errorf("application %s is not configured. Use 'configsync add %s' first", appName, appName)
			}
		}
		if verbose {
			fmt.Printf("Syncing %d specified applications...\n", len(appsToSync))
		}
	}

	// Create symlink manager
	symlinkManager := symlink.NewManager(homeDir, cfg.StorePath, cfg.BackupPath, dryRun, verbose)

	// Sync each application
	var successful []string
	var failed []string

	for _, appConfig := range appsToSync {
		if verbose || dryRun {
			fmt.Printf("\n=== %s ===\n", appConfig.DisplayName)
		}

		if err := symlinkManager.SyncApp(appConfig); err != nil {
			if verbose {
				fmt.Printf("✗ Failed to sync %s: %v\n", appConfig.DisplayName, err)
			}
			failed = append(failed, appConfig.DisplayName)
		} else {
			if verbose || dryRun {
				fmt.Printf("✓ Successfully synced %s\n", appConfig.DisplayName)
			}
			successful = append(successful, appConfig.DisplayName)
		}
	}

	// Update last sync time if not dry run
	if !dryRun && len(successful) > 0 {
		if err := manager.UpdateLastSync(); err != nil {
			fmt.Printf("Warning: failed to update last sync time: %v\n", err)
		}
	}

	// Show summary
	fmt.Println()
	if dryRun {
		fmt.Println("=== DRY RUN SUMMARY ===")
	} else {
		fmt.Println("=== SYNC SUMMARY ===")
	}

	if len(successful) > 0 {
		verb := "synced"
		if dryRun {
			verb = "would be synced"
		}
		fmt.Printf("✓ %d application(s) %s:\n", len(successful), verb)
		for _, name := range successful {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(failed) > 0 {
		verb := "failed to sync"
		if dryRun {
			verb = "would fail to sync"
		}
		fmt.Printf("\n✗ %d application(s) %s:\n", len(failed), verb)
		for _, name := range failed {
			fmt.Printf("  - %s\n", name)
		}
		
		if len(successful) == 0 {
			return fmt.Errorf("failed to sync any applications")
		}
	}

	if dryRun && len(successful) > 0 {
		fmt.Println("\nRun without --dry-run to apply these changes.")
	}

	return nil
}

func init() {
	// No additional flags needed - uses global dry-run and verbose flags
}
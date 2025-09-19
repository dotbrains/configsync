package cmd

import (
	"fmt"

	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/internal/symlink"
	"github.com/spf13/cobra"
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

func runSync(_ *cobra.Command, args []string) error {
	manager := config.NewManager(homeDir)

	if !manager.ConfigExists() {
		return fmt.Errorf("ConfigSync is not initialized. Run 'configsync init' first")
	}

	cfg, err := manager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	appsToSync, err := selectAppsToSync(cfg, args)
	if err != nil {
		return err
	}

	if len(appsToSync) == 0 {
		fmt.Println("No applications configured. Use 'configsync add <app>' to add applications.")
		return nil
	}

	symlinkManager := symlink.NewManager(homeDir, cfg.StorePath, cfg.BackupPath, dryRun, verbose)
	successful, failed := syncApplications(symlinkManager, appsToSync)

	if !dryRun && len(successful) > 0 {
		if err := manager.UpdateLastSync(); err != nil {
			fmt.Printf("Warning: failed to update last sync time: %v\n", err)
		}
	}

	showSyncSummary(successful, failed)

	if len(failed) > 0 && len(successful) == 0 {
		return fmt.Errorf("failed to sync any applications")
	}

	return nil
}

// selectAppsToSync determines which applications to sync based on arguments
func selectAppsToSync(cfg *config.Config, args []string) (map[string]*config.AppConfig, error) {
	if len(args) == 0 {
		if verbose {
			fmt.Printf("Syncing all %d configured applications...\n", len(cfg.Apps))
		}
		return cfg.Apps, nil
	}

	appsToSync := make(map[string]*config.AppConfig)
	for _, appName := range args {
		if app, exists := cfg.Apps[appName]; exists {
			appsToSync[appName] = app
		} else {
			return nil, fmt.Errorf("application %s is not configured. Use 'configsync add %s' first", appName, appName)
		}
	}
	if verbose {
		fmt.Printf("Syncing %d specified applications...\n", len(appsToSync))
	}
	return appsToSync, nil
}

// syncApplications syncs all provided applications and returns successful and failed lists
func syncApplications(symlinkManager *symlink.Manager, apps map[string]*config.AppConfig) ([]string, []string) {
	var successful, failed []string

	for _, appConfig := range apps {
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

	return successful, failed
}

// showSyncSummary displays the sync results summary
func showSyncSummary(successful, failed []string) {
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
	}

	if dryRun && len(successful) > 0 {
		fmt.Println("\nRun without --dry-run to apply these changes.")
	}
}

func init() {
	// No additional flags needed - uses global dry-run and verbose flags
}

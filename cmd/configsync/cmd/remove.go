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

func runRemove(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one application name is required")
	}

	manager := config.NewManager(homeDir)

	if !manager.ConfigExists() {
		return fmt.Errorf("ConfigSync is not initialized. Run 'configsync init' first")
	}

	cfg, err := manager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	symlinkManager := symlink.NewManager(homeDir, cfg.StorePath, cfg.BackupPath, dryRun, verbose)
	successful, failed := removeApplications(manager, symlinkManager, cfg, args)

	showRemoveSummary(successful, failed)

	if len(failed) > 0 && len(successful) == 0 {
		return fmt.Errorf("failed to remove any applications")
	}

	return nil
}

// removeApplications processes the removal of applications
func removeApplications(manager *config.Manager, symlinkManager *symlink.Manager, cfg *config.Config, args []string) ([]string, []string) {
	var successful, failed []string

	for _, appName := range args {
		if verbose {
			fmt.Printf("Processing application: %s\n", appName)
		}

		appConfig, exists := cfg.Apps[appName]
		if !exists {
			if verbose {
				fmt.Printf("  ✗ Application %s is not configured\n", appName)
			}
			failed = append(failed, appName)
			continue
		}

		if err := removeApplication(manager, symlinkManager, appName, appConfig); err != nil {
			failed = append(failed, appConfig.DisplayName)
		} else {
			successful = append(successful, appConfig.DisplayName)
		}
	}

	return successful, failed
}

// removeApplication removes a single application
func removeApplication(manager *config.Manager, symlinkManager *symlink.Manager, appName string, appConfig *config.AppConfig) error {
	if verbose {
		fmt.Printf("  Removing symlinks for %s\n", appConfig.DisplayName)
	}

	if err := symlinkManager.UnsyncApp(appConfig); err != nil {
		if verbose {
			fmt.Printf("  ✗ Failed to unsync %s: %v\n", appConfig.DisplayName, err)
		}
		return err
	}

	if !dryRun {
		if err := manager.RemoveApp(appName); err != nil {
			if verbose {
				fmt.Printf("  ✗ Failed to remove %s from config: %v\n", appConfig.DisplayName, err)
			}
			return err
		}
	}

	if verbose {
		fmt.Printf("  ✓ Successfully removed %s\n", appConfig.DisplayName)
	}
	return nil
}

// showRemoveSummary displays the removal results summary
func showRemoveSummary(successful, failed []string) {
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
	}

	if dryRun && len(successful) > 0 {
		fmt.Println("\nRun without --dry-run to apply these changes.")
	}
}

func init() {
	// No additional flags needed for remove command
}

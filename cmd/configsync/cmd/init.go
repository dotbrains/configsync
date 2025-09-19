package cmd

import (
	"fmt"

	"github.com/dotbrains/configsync/internal/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ConfigSync in the current user directory",
	Long: `Initialize ConfigSync by creating the necessary directory structure
and configuration files in ~/.configsync.

This command creates:
- Configuration directory (~/.configsync)
- Central storage directory (store/)
- Backup directory (backups/)
- Log directory (logs/)
- Initial configuration file (config.yaml)`,
	RunE: runInit,
}

func runInit(_ *cobra.Command, _ []string) error {
	if verbose {
		fmt.Printf("Initializing ConfigSync in %s\n", configDir)
	}

	// Create configuration manager
	manager := config.NewManager(homeDir)

	// Check if already initialized
	if manager.ConfigExists() {
		return fmt.Errorf("ConfigSync is already initialized in %s", configDir)
	}

	// Initialize the configuration
	if err := manager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize ConfigSync: %w", err)
	}

	fmt.Printf("✓ ConfigSync initialized successfully in %s\n", configDir)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add applications: configsync add <app>")
	fmt.Println("  2. Sync configurations: configsync sync")
	fmt.Println("  3. Check status: configsync status")
	fmt.Println()
	fmt.Println("For help with supported applications:")
	fmt.Println("  configsync add --help")

	return nil
}

func init() {
	// No additional flags needed for init command
}

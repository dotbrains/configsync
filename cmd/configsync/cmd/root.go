// Package cmd provides the command-line interface commands for ConfigSync.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	homeDir   string
	configDir string
	verbose   bool
	dryRun    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "configsync",
	Short: "Manage macOS application configurations with centralized storage and syncing",
	Long: `ConfigSync is a command-line tool for managing macOS application settings
and configurations with centralized storage and syncing across multiple Mac systems.

It helps you:
- Store app configurations in a central location
- Use symlinks to sync settings between the central store and app locations
- Deploy configurations to new Mac systems easily
- Create backups before making changes
- Support version control integration`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", "", "home directory (default is $HOME)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without actually doing it")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(deployCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set home directory
	if homeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		homeDir = home
	}

	// Set config directory
	configDir = filepath.Join(homeDir, ".configsync")
}

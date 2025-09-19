package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/internal/util"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all managed configurations",
	Long: `Show the current status of all managed application configurations,
including sync status, last sync time, and any issues.`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	// Show general information
	fmt.Println("ConfigSync Status")
	fmt.Println("=================")
	fmt.Printf("Configuration: %s\n", filepath.Join(manager.GetConfigDir(), "config.yaml"))
	fmt.Printf("Store Path: %s\n", cfg.StorePath)
	fmt.Printf("Backup Path: %s\n", cfg.BackupPath)
	
	if !cfg.LastSync.IsZero() {
		fmt.Printf("Last Sync: %s\n", cfg.LastSync.Format(time.RFC3339))
	} else {
		fmt.Printf("Last Sync: Never\n")
	}
	
	fmt.Printf("Total Apps: %d\n", len(cfg.Apps))
	
	if len(cfg.Apps) == 0 {
		fmt.Println("\nNo applications configured. Use 'configsync add <app>' to add applications.")
		return nil
	}

	fmt.Println("\nApplication Status:")
	fmt.Println("===================")

	// Check status of each application
	for appName, appConfig := range cfg.Apps {
		fmt.Printf("\n%s (%s)\n", appConfig.DisplayName, appName)
		fmt.Printf("  Enabled: %t\n", appConfig.Enabled)
		fmt.Printf("  Paths: %d\n", len(appConfig.Paths))
		
		if !appConfig.LastSynced.IsZero() {
			fmt.Printf("  Last Synced: %s\n", appConfig.LastSynced.Format(time.RFC3339))
		} else {
			fmt.Printf("  Last Synced: Never\n")
		}

		// Check sync status for each path
		syncedCount := 0
		for _, path := range appConfig.Paths {
			sourcePath := expandPath(path.Source, homeDir)
			storePath := filepath.Join(cfg.StorePath, path.Destination)
			
			status := getPathStatus(sourcePath, storePath)
			if status == "synced" {
				syncedCount++
			}
			
			if verbose {
				fmt.Printf("    %s -> %s (%s)\n", path.Source, path.Destination, status)
			}
		}
		
		fmt.Printf("  Sync Status: %d/%d paths synced\n", syncedCount, len(appConfig.Paths))
	}

	return nil
}

func getPathStatus(sourcePath, storePath string) string {
	// Check if source exists
	sourceExists := util.PathExists(sourcePath)
	storeExists := util.PathExists(storePath)
	
	if !sourceExists && !storeExists {
		return "missing"
	}
	
	if !sourceExists && storeExists {
		return "not_synced"
	}
	
	if sourceExists && !storeExists {
		return "not_synced"
	}
	
	// Both exist - check if source is a symlink to store
	if isSymlink(sourcePath) {
		link, err := os.Readlink(sourcePath)
		if err != nil {
			return "error"
		}
		
		// Resolve relative paths
		if !filepath.IsAbs(link) {
			baseDir := filepath.Dir(sourcePath)
			link = filepath.Join(baseDir, link)
		}
		
		// Clean both paths for comparison
		link = filepath.Clean(link)
		storePath = filepath.Clean(storePath)
		
		if link == storePath {
			return "synced"
		}
		return "wrong_link"
	}
	
	return "not_synced"
}

func expandPath(path, homeDir string) string {
	if filepath.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}


func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

func init() {
	// No additional flags needed for status command
}
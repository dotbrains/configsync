package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/dotbrains/configsync/internal/backup"
	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/internal/deploy"
	"github.com/dotbrains/configsync/internal/util"
)

var (
	backupKeepDays int
	backupValidate bool
	restoreAll     bool
	exportOutput   string
	exportApps     []string
	importForce    bool
	deployForce    bool
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup [app1] [app2] ...",
	Short: "Create and manage backups of configurations",
	Long: `Create backups of original configurations before symlinking.

If no app names are provided, all managed applications will be backed up.

Examples:
  configsync backup              # Backup all apps
  configsync backup vscode       # Backup only VS Code
  configsync backup --validate   # Validate existing backups
  configsync backup --cleanup --keep-days 30  # Clean old backups`,
	RunE: runBackup,
}

func runBackup(cmd *cobra.Command, args []string) error {
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

	// Create backup manager
	backupManager := backup.NewManager(cfg.BackupPath, homeDir, verbose)

	if backupValidate {
		return validateBackups(backupManager, args, cfg)
	}

	if cmd.Flags().Changed("keep-days") {
		return cleanupBackups(backupManager, args, cfg)
	}

	return createBackups(backupManager, args, cfg)
}

func createBackups(backupManager *backup.Manager, args []string, cfg *config.Config) error {
	// Get applications to backup
	var appsToBackup map[string]*config.AppConfig
	if len(args) == 0 {
		appsToBackup = cfg.Apps
		if len(appsToBackup) == 0 {
			fmt.Println("No applications configured. Use 'configsync add <app>' to add applications.")
			return nil
		}
	} else {
		appsToBackup = make(map[string]*config.AppConfig)
		for _, appName := range args {
			if app, exists := cfg.Apps[appName]; exists {
				appsToBackup[appName] = app
			} else {
				return fmt.Errorf("application %s is not configured", appName)
			}
		}
	}

	// Create backups for each application
	var successful []string
	var failed []string

	for appName, appConfig := range appsToBackup {
		if verbose {
			fmt.Printf("\n=== %s ===\n", appConfig.DisplayName)
		}

		pathErrors := 0
		for _, path := range appConfig.Paths {
			if err := backupManager.BackupPath(appName, &path); err != nil {
				if verbose {
					fmt.Printf("  ✗ Failed to backup %s: %v\n", path.Source, err)
				}
				pathErrors++
			}
		}

		if pathErrors == 0 {
			successful = append(successful, appConfig.DisplayName)
			if verbose {
				fmt.Printf("✓ Backed up %s\n", appConfig.DisplayName)
			}
		} else {
			failed = append(failed, appConfig.DisplayName)
		}
	}

	// Show results
	fmt.Println()
	if len(successful) > 0 {
		fmt.Printf("✓ Successfully backed up %d application(s):\n", len(successful))
		for _, name := range successful {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(failed) > 0 {
		fmt.Printf("\n✗ Failed to backup %d application(s):\n", len(failed))
		for _, name := range failed {
			fmt.Printf("  - %s\n", name)
		}
	}

	return nil
}

func validateBackups(backupManager *backup.Manager, args []string, cfg *config.Config) error {
	if len(args) == 0 {
		// Validate all apps
		for appName := range cfg.Apps {
			args = append(args, appName)
		}
	}

	if len(args) == 0 {
		fmt.Println("No applications to validate.")
		return nil
	}

	var totalBackups int
	var validBackups int
	var invalidBackups int

	for _, appName := range args {
		backups, err := backupManager.ListBackups(appName)
		if err != nil {
			fmt.Printf("Error listing backups for %s: %v\n", appName, err)
			continue
		}

		if len(backups) == 0 {
			if verbose {
				fmt.Printf("%s: No backups found\n", appName)
			}
			continue
		}

		for _, backup := range backups {
			totalBackups++
			if err := backupManager.ValidateBackup(backup); err != nil {
				fmt.Printf("✗ %s: %s - %v\n", appName, filepath.Base(backup.OriginalPath), err)
				invalidBackups++
			} else {
				if verbose {
					fmt.Printf("✓ %s: %s\n", appName, filepath.Base(backup.OriginalPath))
				}
				validBackups++
			}
		}
	}

	fmt.Printf("\nBackup validation complete: %d valid, %d invalid (total: %d)\n", 
		validBackups, invalidBackups, totalBackups)

	return nil
}

func cleanupBackups(backupManager *backup.Manager, args []string, cfg *config.Config) error {
	if len(args) == 0 {
		// Cleanup all apps
		for appName := range cfg.Apps {
			args = append(args, appName)
		}
	}

	if len(args) == 0 {
		fmt.Println("No applications to cleanup.")
		return nil
	}

	for _, appName := range args {
		if err := backupManager.CleanupBackups(appName, backupKeepDays); err != nil {
			fmt.Printf("Error cleaning up backups for %s: %v\n", appName, err)
		}
	}

	return nil
}

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore [app1] [app2] ...",
	Short: "Restore original configurations from backup",
	Long: `Restore original configurations from backup for specified applications.

Examples:
  configsync restore vscode      # Restore VS Code from backup
  configsync restore git ssh     # Restore multiple apps
  configsync restore --all       # Restore all backed up configurations`,
	RunE: runRestore,
}

func runRestore(cmd *cobra.Command, args []string) error {
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

	// Create backup manager
	backupManager := backup.NewManager(cfg.BackupPath, homeDir, verbose)

	// Get applications to restore
	var appsToRestore []string
	if restoreAll {
		// Restore all apps that have backups
		for appName := range cfg.Apps {
			backups, err := backupManager.ListBackups(appName)
			if err != nil {
				continue
			}
			if len(backups) > 0 {
				appsToRestore = append(appsToRestore, appName)
			}
		}
	} else if len(args) > 0 {
		appsToRestore = args
	} else {
		return fmt.Errorf("specify applications to restore or use --all flag")
	}

	if len(appsToRestore) == 0 {
		fmt.Println("No applications with backups found.")
		return nil
	}

	// Restore each application
	var successful []string
	var failed []string

	for _, appName := range appsToRestore {
		appConfig, exists := cfg.Apps[appName]
		if !exists {
			failed = append(failed, appName)
			if verbose {
				fmt.Printf("Application %s is not configured\n", appName)
			}
			continue
		}

		if verbose {
			fmt.Printf("\n=== %s ===\n", appConfig.DisplayName)
		}

		pathErrors := 0
		for _, path := range appConfig.Paths {
			if err := backupManager.RestorePath(appName, &path); err != nil {
				if verbose {
					fmt.Printf("  ✗ Failed to restore %s: %v\n", path.Source, err)
				}
				pathErrors++
			}
		}

		if pathErrors == 0 {
			successful = append(successful, appConfig.DisplayName)
			if verbose {
				fmt.Printf("✓ Restored %s\n", appConfig.DisplayName)
			}
		} else {
			failed = append(failed, appConfig.DisplayName)
		}
	}

	// Show results
	fmt.Println()
	if len(successful) > 0 {
		fmt.Printf("✓ Successfully restored %d application(s):\n", len(successful))
		for _, name := range successful {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(failed) > 0 {
		fmt.Printf("\n✗ Failed to restore %d application(s):\n", len(failed))
		for _, name := range failed {
			fmt.Printf("  - %s\n", name)
		}
	}

	return nil
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [--output bundle.tar.gz] [--apps app1,app2]",
	Short: "Export configuration bundle for deployment",
	Long: `Export configuration bundle that can be imported on another Mac.

Examples:
  configsync export                           # Export all apps to default location
  configsync export --output my-config.tar.gz # Export to specific file
  configsync export --apps vscode,git        # Export specific apps only`,
	RunE: runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
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

	// Create deploy manager
	deployManager := deploy.NewManager(homeDir, cfg.StorePath, cfg.BackupPath, verbose)

	// Determine output file
	outputFile := exportOutput
	if outputFile == "" {
		outputFile = "configsync-bundle.tar.gz"
	}

	// Convert output to absolute path
	if !filepath.IsAbs(outputFile) {
		cwd, _ := os.Getwd()
		outputFile = filepath.Join(cwd, outputFile)
	}

	// Create bundle
	if err := deployManager.ExportBundle(outputFile, exportApps, manager); err != nil {
		return fmt.Errorf("failed to export bundle: %w", err)
	}

	fmt.Printf("\n✓ Configuration bundle exported to: %s\n", outputFile)
	fmt.Println("\nTo import on another Mac:")
	fmt.Printf("  configsync import %s\n", filepath.Base(outputFile))
	fmt.Printf("  configsync deploy\n")

	return nil
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <bundle.tar.gz>",
	Short: "Import configuration bundle",
	Long: `Import configuration bundle from another Mac.

This extracts and validates the bundle but doesn't deploy it yet.
Use 'configsync deploy' to apply the imported configurations.

Examples:
  configsync import my-bundle.tar.gz
  configsync import --force bundle.tar.gz   # Force import even with conflicts`,
	RunE: runImport,
	Args: cobra.ExactArgs(1),
}

func runImport(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
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

	// Create deploy manager
	deployManager := deploy.NewManager(homeDir, cfg.StorePath, cfg.BackupPath, verbose)

	// Create import directory
	importDir := filepath.Join(configDir, "import")
	if err := os.RemoveAll(importDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean import directory: %w", err)
	}

	// Import bundle
	bundle, err := deployManager.ImportBundle(bundlePath, importDir)
	if err != nil {
		return fmt.Errorf("failed to import bundle: %w", err)
	}

	fmt.Printf("\n✓ Bundle imported successfully\n")
	fmt.Printf("  Created: %s by %s\n", bundle.CreatedAt.Format("2006-01-02 15:04"), bundle.CreatedBy)
	fmt.Printf("  Platform: %s\n", bundle.Metadata["platform"])
	fmt.Printf("  Applications: %d\n", len(bundle.Apps))

	fmt.Println("\nNext step: Run 'configsync deploy' to apply these configurations")

	return nil
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy imported configurations to current system",
	Long: `Deploy configurations to the current system from the last imported bundle.

This command applies the configurations that were imported with 'configsync import'.
Use --force to override any conflicts with existing configurations.

Examples:
  configsync deploy              # Deploy imported configurations
  configsync deploy --force      # Force deploy even with conflicts`,
	RunE: runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
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

	// Check if import directory exists
	importDir := filepath.Join(configDir, "import")
	if !util.PathExists(importDir) {
		return fmt.Errorf("no imported bundle found. Run 'configsync import <bundle>' first")
	}

	// Load bundle metadata
	bundleFile := filepath.Join(importDir, "bundle.yaml")
	if !util.PathExists(bundleFile) {
		return fmt.Errorf("invalid import directory. Run 'configsync import <bundle>' first")
	}

	// Load bundle
	deployManager := deploy.NewManager(homeDir, cfg.StorePath, cfg.BackupPath, verbose)
	bundle, err := deployManager.ImportBundle("", importDir) // Empty bundle path since already extracted
	if err != nil {
		return fmt.Errorf("failed to load imported bundle: %w", err)
	}

	// Deploy bundle
	if err := deployManager.DeployBundle(bundle, importDir, manager, deployForce); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	return nil
}


func init() {
	// Backup command flags
	backupCmd.Flags().IntVar(&backupKeepDays, "keep-days", 30, "cleanup backups older than N days")
	backupCmd.Flags().BoolVar(&backupValidate, "validate", false, "validate existing backups")

	// Restore command flags  
	restoreCmd.Flags().BoolVar(&restoreAll, "all", false, "restore all backed up applications")

	// Export command flags
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file for bundle (default: configsync-bundle.tar.gz)")
	exportCmd.Flags().StringSliceVar(&exportApps, "apps", []string{}, "comma-separated list of apps to export (default: all)")

	// Import command flags
	importCmd.Flags().BoolVar(&importForce, "force", false, "force import even with conflicts")

	// Deploy command flags
	deployCmd.Flags().BoolVar(&deployForce, "force", false, "force deploy even with conflicts")
}

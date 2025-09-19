package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/pkg/apps"
	"github.com/spf13/cobra"
)

var (
	discoverAutoAdd bool
	discoverList    bool
	discoverFilter  string
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover applications installed on the system and their configurations",
	Long: `The discover command scans your Mac for installed applications and automatically
detects their configuration files and directories. It uses multiple detection methods:

1. Known Applications: Uses built-in knowledge of popular Mac apps
2. System Profiler: Scans system application database
3. Spotlight Search: Uses mdfind to locate .app bundles
4. Directory Scanning: Scans common app installation locations
5. Smart Pattern Detection: Automatically detects config paths using common patterns

Examples:
  # List all discovered applications
  configsync discover --list

  # Auto-add all detected applications to your configuration
  configsync discover --auto-add

  # Filter results to specific apps
  configsync discover --filter="chrome,slack,vscode"

  # Discover and show details in dry-run mode
  configsync discover --dry-run --verbose`,
	RunE: runDiscover,
}

func init() {
	discoverCmd.Flags().BoolVar(&discoverAutoAdd, "auto-add", false, "automatically add discovered apps to configuration")
	discoverCmd.Flags().BoolVar(&discoverList, "list", false, "list all discovered applications")
	discoverCmd.Flags().StringVar(&discoverFilter, "filter", "", "comma-separated list of app names to filter results")
}

func runDiscover(_ *cobra.Command, _ []string) error {
	// Initialize detector
	detector := apps.NewAppDetector(homeDir)

	if verbose {
		fmt.Printf("Scanning for installed applications...\n")
	}

	// Scan for installed apps
	installedApps, err := detector.ScanInstalledApps()
	if err != nil {
		return fmt.Errorf("failed to scan installed apps: %v", err)
	}

	if verbose {
		fmt.Printf("Found %d installed applications\n\n", len(installedApps))
	}

	// Auto-detect configurations
	detectedConfigs, err := detector.AutoDetectApps()
	if err != nil {
		return fmt.Errorf("failed to auto-detect app configurations: %v", err)
	}

	// Filter results if requested
	var filteredConfigs []*config.AppConfig
	if discoverFilter != "" {
		filterNames := strings.Split(strings.ToLower(discoverFilter), ",")
		for _, appConfig := range detectedConfigs {
			for _, filterName := range filterNames {
				if strings.Contains(strings.ToLower(appConfig.Name), strings.TrimSpace(filterName)) ||
					strings.Contains(strings.ToLower(appConfig.DisplayName), strings.TrimSpace(filterName)) {
					filteredConfigs = append(filteredConfigs, appConfig)
					break
				}
			}
		}
		detectedConfigs = filteredConfigs
	}

	if discoverList {
		return printDiscoveredApps(detectedConfigs, installedApps)
	}

	if discoverAutoAdd {
		return autoAddDiscoveredApps(detectedConfigs)
	}

	// Default behavior: show summary and ask for confirmation
	return showDiscoveryResults(detectedConfigs)
}

func printDiscoveredApps(detectedConfigs []*config.AppConfig, installedApps []apps.InstalledApp) error {
	if len(detectedConfigs) == 0 {
		fmt.Println("No applications with configuration files were discovered.")
		return nil
	}

	// Create a map for quick lookup of installed apps
	installedMap := make(map[string]apps.InstalledApp)
	for _, app := range installedApps {
		installedMap[app.Name] = app
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "NAME\tDISPLAY NAME\tBUNDLE ID\tPATHS FOUND\tSTATUS"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "----\t------------\t---------\t-----------\t------"); err != nil {
		return err
	}

	for _, appConfig := range detectedConfigs {
		status := "Unknown"
		if installedApp, exists := installedMap[appConfig.Name]; exists {
			if installedApp.BundleID != "" {
				status = "Installed"
			} else {
				status = "Detected"
			}
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			appConfig.Name,
			appConfig.DisplayName,
			appConfig.BundleID,
			len(appConfig.Paths),
			status); err != nil {
			return err
		}

		if verbose {
			for _, path := range appConfig.Paths {
				if _, err := fmt.Fprintf(w, "\t↳ %s\t%s\t%s\t\n", path.Source, path.Type, ""); err != nil {
					return err
				}
			}
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func showDiscoveryResults(detectedConfigs []*config.AppConfig) error {
	if len(detectedConfigs) == 0 {
		fmt.Println("No applications with configuration files were discovered.")
		fmt.Println("\nTry running with --verbose to see more details about the scanning process.")
		return nil
	}

	fmt.Printf("🔍 Discovered %d applications with configuration files:\n\n", len(detectedConfigs))

	for i, appConfig := range detectedConfigs {
		fmt.Printf("%d. %s (%s)\n", i+1, appConfig.DisplayName, appConfig.Name)
		if appConfig.BundleID != "" {
			fmt.Printf("   Bundle ID: %s\n", appConfig.BundleID)
		}
		fmt.Printf("   Configuration paths found: %d\n", len(appConfig.Paths))

		if verbose {
			for _, path := range appConfig.Paths {
				fmt.Printf("   ↳ %s (%s)\n", path.Source, path.Type)
			}
		}
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("• Run 'configsync discover --auto-add' to add all discovered apps")
	fmt.Println("• Run 'configsync discover --filter=\"app1,app2\"' to filter specific apps")
	fmt.Println("• Run 'configsync add <app-name>' to manually add specific applications")
	fmt.Println("• Run 'configsync discover --list --verbose' for detailed path information")

	return nil
}

func autoAddDiscoveredApps(detectedConfigs []*config.AppConfig) error {
	if len(detectedConfigs) == 0 {
		fmt.Println("No applications were discovered for auto-adding.")
		return nil
	}

	// Load existing configuration
	configManager := config.NewManager(configDir)
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	added := 0
	skipped := 0

	fmt.Printf("🚀 Auto-adding %d discovered applications...\n\n", len(detectedConfigs))

	for _, appConfig := range detectedConfigs {
		// Check if app already exists in configuration
		if _, exists := cfg.Apps[appConfig.Name]; exists {
			if verbose {
				fmt.Printf("⏭️  Skipping %s (already configured)\n", appConfig.DisplayName)
			}
			skipped++
			continue
		}

		if dryRun {
			fmt.Printf("🔍 Would add: %s (%d paths)\n", appConfig.DisplayName, len(appConfig.Paths))
			added++
			continue
		}

		// Add the application to configuration
		cfg.Apps[appConfig.Name] = appConfig
		fmt.Printf("✅ Added: %s (%d paths)\n", appConfig.DisplayName, len(appConfig.Paths))
		added++
	}

	if !dryRun && added > 0 {
		// Save the updated configuration
		if err := configManager.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %v", err)
		}

		fmt.Printf("\n🎉 Successfully added %d applications to your configuration!\n", added)
		if skipped > 0 {
			fmt.Printf("⏭️  Skipped %d applications (already configured)\n", skipped)
		}

		fmt.Println("\nNext steps:")
		fmt.Println("• Run 'configsync sync' to create symlinks for the new applications")
		fmt.Println("• Run 'configsync status' to check the current sync status")
	} else if dryRun {
		fmt.Printf("\n🔍 Dry run complete. Would have added %d applications.\n", added)
		if skipped > 0 {
			fmt.Printf("⏭️  Would have skipped %d applications (already configured)\n", skipped)
		}
	}

	return nil
}

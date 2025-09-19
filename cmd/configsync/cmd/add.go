package cmd

import (
	"fmt"
	"sort"

	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/pkg/apps"
	"github.com/spf13/cobra"
)

var (
	listSupported bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [app1] [app2] ...",
	Short: "Add application(s) to configuration management",
	Long: `Add one or more applications to ConfigSync management.

ConfigSync will automatically detect common configuration paths for known applications.
You can also specify custom paths using the --path flag.

Examples:
  configsync add vscode
  configsync add "Google Chrome" Firefox
  configsync add Terminal iTerm2
  configsync add --list-supported`,
	RunE: runAdd,
}

func runAdd(_ *cobra.Command, args []string) error {
	if listSupported {
		return showSupportedApps()
	}

	if len(args) == 0 {
		return fmt.Errorf("at least one application name is required\nUse 'configsync add --list-supported' to see supported applications")
	}

	manager := config.NewManager(homeDir)
	detector := apps.NewAppDetector(homeDir)

	if !manager.ConfigExists() {
		return fmt.Errorf("ConfigSync is not initialized. Run 'configsync init' first")
	}

	successful, failed := addApplications(manager, detector, args)
	showAddResults(successful, failed)

	if len(failed) > 0 && len(successful) == 0 {
		return fmt.Errorf("failed to add any applications")
	}

	return nil
}

// addApplications processes adding applications and returns successful and failed lists
func addApplications(manager *config.Manager, detector *apps.AppDetector, args []string) ([]string, []string) {
	var successful, failed []string

	for _, appName := range args {
		if verbose {
			fmt.Printf("Processing application: %s\n", appName)
		}

		appConfig, err := detector.DetectApp(appName)
		if err != nil {
			if verbose {
				fmt.Printf("  ✗ Failed to detect %s: %v\n", appName, err)
			}
			failed = append(failed, appName)
			continue
		}

		if err := manager.AddApp(appConfig); err != nil {
			if verbose {
				fmt.Printf("  ✗ Failed to add %s: %v\n", appName, err)
			}
			failed = append(failed, appName)
			continue
		}

		if verbose {
			fmt.Printf("  ✓ Successfully added %s (%d paths)\n", appConfig.DisplayName, len(appConfig.Paths))
			for _, path := range appConfig.Paths {
				fmt.Printf("    - %s\n", path.Source)
			}
		}
		successful = append(successful, appConfig.DisplayName)
	}

	return successful, failed
}

// showAddResults displays the add operation results
func showAddResults(successful, failed []string) {
	if len(successful) > 0 {
		fmt.Printf("✓ Successfully added %d application(s):\n", len(successful))
		for _, name := range successful {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(failed) > 0 {
		fmt.Printf("\n✗ Failed to add %d application(s):\n", len(failed))
		for _, name := range failed {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println("\nTip: Use 'configsync add --list-supported' to see supported applications")
	}

	if len(successful) > 0 {
		fmt.Println("\nNext step: Run 'configsync sync' to create symlinks")
	}
}

func showSupportedApps() error {
	detector := apps.NewAppDetector(homeDir)
	supportedApps := detector.GetSupportedApps()

	if len(supportedApps) == 0 {
		fmt.Println("No supported applications found")
		return nil
	}

	// Sort apps alphabetically
	sort.Strings(supportedApps)

	fmt.Println("Supported applications:")
	fmt.Println("======================")

	for _, app := range supportedApps {
		fmt.Printf("  %s\n", app)
	}

	fmt.Printf("\nTotal: %d applications\n", len(supportedApps))
	fmt.Println("\nNote: ConfigSync can also auto-detect other applications")
	fmt.Println("by searching for preference files and configurations.")

	return nil
}

func init() {
	addCmd.Flags().BoolVar(&listSupported, "list-supported", false, "list all supported applications")
}

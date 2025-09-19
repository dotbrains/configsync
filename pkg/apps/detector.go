// Package apps provides application detection and configuration functionality.
package apps

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dotbrains/configsync/internal/config"
	"github.com/dotbrains/configsync/internal/util"
)

// AppDetector handles detection and configuration of macOS applications
type AppDetector struct {
	lastScanTime  time.Time
	homeDir       string
	installedApps []InstalledApp
	cacheDuration time.Duration
}

// NewAppDetector creates a new application detector
func NewAppDetector(homeDir string) *AppDetector {
	return &AppDetector{
		homeDir:       homeDir,
		installedApps: []InstalledApp{},
		cacheDuration: 5 * time.Minute, // Cache for 5 minutes
	}
}

// DetectApp attempts to detect an application and its configuration paths
func (d *AppDetector) DetectApp(appName string) (*config.AppConfig, error) {
	// Normalize app name
	normalizedName := strings.ToLower(strings.ReplaceAll(appName, " ", ""))

	// Try to find app configuration using various strategies
	if appConfig := d.detectKnownApp(normalizedName); appConfig != nil {
		return appConfig, nil
	}

	if appConfig := d.detectByBundleID(appName); appConfig != nil {
		return appConfig, nil
	}

	if appConfig := d.detectByPreferences(appName); appConfig != nil {
		return appConfig, nil
	}

	return nil, fmt.Errorf("could not detect configuration for app: %s", appName)
}

// InstalledApp represents an application installed on the system
type InstalledApp struct {
	Name        string `json:"name"`
	BundleID    string `json:"bundle_id"`
	Path        string `json:"path"`
	Version     string `json:"version"`
	DisplayName string `json:"display_name"`
}

// ScanInstalledApps scans the system for installed applications using system_profiler and mdfind
func (d *AppDetector) ScanInstalledApps() ([]InstalledApp, error) {
	// Check cache first
	if time.Since(d.lastScanTime) < d.cacheDuration && len(d.installedApps) > 0 {
		return d.installedApps, nil
	}

	var allApps []InstalledApp
	var scanCounts []int

	// Method 1: Use system_profiler to get installed applications
	if apps, err := d.scanWithSystemProfiler(); err == nil {
		allApps = append(allApps, apps...)
		scanCounts = append(scanCounts, len(apps))
	} else {
		scanCounts = append(scanCounts, 0)
	}

	// Method 2: Use mdfind to find .app bundles
	if apps, err := d.scanWithMdfind(); err == nil {
		allApps = append(allApps, apps...)
		scanCounts = append(scanCounts, len(apps))
	} else {
		scanCounts = append(scanCounts, 0)
	}

	// Method 3: Scan common application directories
	apps := d.scanCommonDirectories()
	allApps = append(allApps, apps...)
	scanCounts = append(scanCounts, len(apps))

	beforeDeduplication := len(allApps)

	// Remove duplicates based on bundle ID
	uniqueApps := d.removeDuplicateApps(allApps)

	afterDeduplication := len(uniqueApps)
	duplicatesRemoved := beforeDeduplication - afterDeduplication

	// Debug info: only visible in verbose mode (would need to be passed down)
	// For now, this is just internal tracking
	_ = scanCounts
	_ = duplicatesRemoved

	// Cache the results
	d.installedApps = uniqueApps
	d.lastScanTime = time.Now()

	return uniqueApps, nil
}

// scanWithSystemProfiler uses system_profiler to get application information
func (d *AppDetector) scanWithSystemProfiler() ([]InstalledApp, error) {
	cmd := exec.Command("system_profiler", "SPApplicationsDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run system_profiler: %v", err)
	}

	var result struct {
		SPApplicationsDataType []struct {
			Name    string `json:"_name"`
			Path    string `json:"path"`
			Version string `json:"version"`
		} `json:"SPApplicationsDataType"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse system_profiler output: %v", err)
	}

	var apps []InstalledApp
	for _, app := range result.SPApplicationsDataType {
		if app.Name != "" {
			installedApp := InstalledApp{
				Name:        strings.ToLower(strings.ReplaceAll(app.Name, " ", "")),
				DisplayName: app.Name,
				Path:        app.Path,
				Version:     app.Version,
				BundleID:    d.extractBundleID(app.Path),
			}
			apps = append(apps, installedApp)
		}
	}

	return apps, nil
}

// scanWithMdfind uses mdfind to locate .app bundles
func (d *AppDetector) scanWithMdfind() ([]InstalledApp, error) {
	cmd := exec.Command("mdfind", "kMDItemContentType == 'com.apple.application-bundle'")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run mdfind: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var apps []InstalledApp

	for _, line := range lines {
		if line == "" || !strings.HasSuffix(line, ".app") {
			continue
		}

		appName := filepath.Base(line)
		appName = strings.TrimSuffix(appName, ".app")

		installedApp := InstalledApp{
			Name:        strings.ToLower(strings.ReplaceAll(appName, " ", "")),
			DisplayName: appName,
			Path:        line,
			BundleID:    d.extractBundleID(line),
		}

		apps = append(apps, installedApp)
	}

	return apps, nil
}

// scanCommonDirectories scans common application installation directories
func (d *AppDetector) scanCommonDirectories() []InstalledApp {
	commonDirs := []string{
		"/Applications",
		filepath.Join(d.homeDir, "Applications"),
		"/System/Applications",
		"/System/Library/CoreServices",
	}

	var apps []InstalledApp

	for _, dir := range commonDirs {
		if !util.PathExists(dir) {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
				continue
			}

			appPath := filepath.Join(dir, entry.Name())
			appName := strings.TrimSuffix(entry.Name(), ".app")

			installedApp := InstalledApp{
				Name:        strings.ToLower(strings.ReplaceAll(appName, " ", "")),
				DisplayName: appName,
				Path:        appPath,
				BundleID:    d.extractBundleID(appPath),
			}

			apps = append(apps, installedApp)
		}
	}

	return apps
}

// extractBundleID extracts the bundle ID from an application path
func (d *AppDetector) extractBundleID(appPath string) string {
	if appPath == "" {
		return ""
	}

	infoPath := filepath.Join(appPath, "Contents", "Info.plist")
	if !util.PathExists(infoPath) {
		return ""
	}

	// Use plutil to extract bundle ID from Info.plist
	cmd := exec.Command("plutil", "-extract", "CFBundleIdentifier", "raw", infoPath)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// removeDuplicateApps removes duplicate apps based on bundle ID, name, and path
func (d *AppDetector) removeDuplicateApps(apps []InstalledApp) []InstalledApp {
	seen := make(map[string]bool)
	var unique []InstalledApp

	for _, app := range apps {
		// Create a comprehensive key for deduplication
		key := d.generateAppKey(app)

		if !seen[key] {
			seen[key] = true
			unique = append(unique, app)
		} else {
			// If we see a duplicate, prefer the one with more complete information
			for i, existingApp := range unique {
				if d.generateAppKey(existingApp) == key {
					if d.shouldReplaceApp(existingApp, app) {
						unique[i] = app
					}
					break
				}
			}
		}
	}

	return unique
}

// generateAppKey creates a unique key for app deduplication
func (d *AppDetector) generateAppKey(app InstalledApp) string {
	// Priority order for key generation:
	// 1. Bundle ID (most reliable)
	// 2. App path (for apps without bundle ID)
	// 3. Normalized name (fallback)

	if app.BundleID != "" {
		return "bundle:" + app.BundleID
	}

	if app.Path != "" {
		// Use clean path to normalize /Applications vs /Applications/ differences
		cleanPath := filepath.Clean(app.Path)
		return "path:" + cleanPath
	}

	return "name:" + app.Name
}

// shouldReplaceApp determines if the new app should replace the existing one
func (d *AppDetector) shouldReplaceApp(existing, newApp InstalledApp) bool {
	// Prefer apps with bundle IDs
	if existing.BundleID == "" && newApp.BundleID != "" {
		return true
	}
	if existing.BundleID != "" && newApp.BundleID == "" {
		return false
	}

	// Prefer apps with version information
	if existing.Version == "" && newApp.Version != "" {
		return true
	}
	if existing.Version != "" && newApp.Version == "" {
		return false
	}

	// Prefer apps in /Applications over other locations
	if !strings.HasPrefix(existing.Path, "/Applications/") && strings.HasPrefix(newApp.Path, "/Applications/") {
		return true
	}

	return false
}

// AutoDetectApps automatically detects applications and generates configurations
func (d *AppDetector) AutoDetectApps() ([]*config.AppConfig, error) {
	installedApps, err := d.ScanInstalledApps()
	if err != nil {
		return nil, fmt.Errorf("failed to scan installed apps: %v", err)
	}

	var detectedConfigs []*config.AppConfig

	for _, app := range installedApps {
		// First try to detect using known apps
		if appConfig := d.detectKnownApp(app.Name); appConfig != nil {
			// Enhance with bundle ID from installed app if available
			if appConfig.BundleID == "" && app.BundleID != "" {
				appConfig.BundleID = app.BundleID
			}
			detectedConfigs = append(detectedConfigs, appConfig)
			continue
		}

		// Try smart detection using bundle ID and common patterns
		if appConfig := d.smartDetectApp(app); appConfig != nil {
			detectedConfigs = append(detectedConfigs, appConfig)
		}
	}

	// Remove duplicate configurations
	deduplicatedConfigs := d.removeDuplicateConfigs(detectedConfigs)

	return deduplicatedConfigs, nil
}

// removeDuplicateConfigs removes duplicate app configurations
func (d *AppDetector) removeDuplicateConfigs(configs []*config.AppConfig) []*config.AppConfig {
	seen := make(map[string]bool)
	var unique []*config.AppConfig

	for _, cfg := range configs {
		key := d.generateConfigKey(cfg)

		if !seen[key] {
			seen[key] = true
			unique = append(unique, cfg)
		} else {
			// If we see a duplicate, prefer the one with more configuration paths
			for i, existingConfig := range unique {
				if d.generateConfigKey(existingConfig) == key {
					if len(cfg.Paths) > len(existingConfig.Paths) {
						unique[i] = cfg
					} else if len(cfg.Paths) == len(existingConfig.Paths) && cfg.BundleID != "" && existingConfig.BundleID == "" {
						// Prefer config with bundle ID
						unique[i] = cfg
					}
					break
				}
			}
		}
	}

	return unique
}

// generateConfigKey creates a unique key for configuration deduplication
func (d *AppDetector) generateConfigKey(cfg *config.AppConfig) string {
	// Priority order for key generation:
	// 1. Bundle ID (most reliable)
	// 2. Normalized name (fallback)

	if cfg.BundleID != "" {
		return "bundle:" + cfg.BundleID
	}

	return "name:" + cfg.Name
}

// smartDetectApp uses smart heuristics to detect app configuration
func (d *AppDetector) smartDetectApp(app InstalledApp) *config.AppConfig {
	appConfig := config.NewAppConfig(app.Name, app.DisplayName)
	appConfig.BundleID = app.BundleID

	var foundPaths []localPath

	// Pattern 1: Check for preferences in ~/Library/Preferences/
	if app.BundleID != "" {
		prefsPath := filepath.Join(d.homeDir, "Library", "Preferences", app.BundleID+".plist")
		if util.PathExists(prefsPath) {
			relPath := filepath.Join("Library", "Preferences", app.BundleID+".plist")
			foundPaths = append(foundPaths, localPath{
				Source:      prefsPath,
				Destination: relPath,
				Type:        config.PathTypeFile,
				Required:    false,
			})
		}
	}

	// Pattern 2: Check for application support directory
	appSupportPaths := []string{
		filepath.Join(d.homeDir, "Library", "Application Support", app.DisplayName),
		filepath.Join(d.homeDir, "Library", "Application Support", app.Name),
	}

	for _, appSupportPath := range appSupportPaths {
		if util.PathExists(appSupportPath) {
			relPath, _ := filepath.Rel(d.homeDir, appSupportPath)
			foundPaths = append(foundPaths, localPath{
				Source:      appSupportPath,
				Destination: relPath,
				Type:        config.PathTypeDirectory,
				Required:    false,
			})
			break
		}
	}

	// Pattern 3: Check for containers (sandboxed apps)
	if app.BundleID != "" {
		containerPaths := []string{
			filepath.Join(d.homeDir, "Library", "Containers", app.BundleID),
			filepath.Join(d.homeDir, "Library", "Group Containers", app.BundleID),
		}

		for _, containerPath := range containerPaths {
			if util.PathExists(containerPath) {
				relPath, _ := filepath.Rel(d.homeDir, containerPath)
				foundPaths = append(foundPaths, localPath{
					Source:      containerPath,
					Destination: relPath,
					Type:        config.PathTypeDirectory,
					Required:    false,
				})
			}
		}
	}

	// Pattern 4: Check for dotfiles in home directory (for CLI tools)
	if !strings.Contains(app.Path, "Applications") {
		potentialDotfiles := []string{
			filepath.Join(d.homeDir, "."+app.Name+"rc"),
			filepath.Join(d.homeDir, "."+app.Name),
			filepath.Join(d.homeDir, "."+app.Name+".conf"),
			filepath.Join(d.homeDir, ".config", app.Name),
		}

		for _, dotfile := range potentialDotfiles {
			if util.PathExists(dotfile) {
				relPath, _ := filepath.Rel(d.homeDir, dotfile)
				foundPaths = append(foundPaths, localPath{
					Source:      dotfile,
					Destination: relPath,
					Type:        config.PathTypeFile,
					Required:    false,
				})
			}
		}
	}

	// Convert localPath to the expected Path format
	for _, cp := range foundPaths {
		appConfig.AddPath(cp.Source, cp.Destination, cp.Type, cp.Required)
	}

	// Only return if we found at least one configuration path
	if len(foundPaths) > 0 {
		return appConfig
	}

	return nil
}

// localPath represents a configuration path (internal struct for smart detection)
type localPath struct {
	Source      string
	Destination string
	Type        config.PathType
	Required    bool
}

// GetInstalledApps returns the cached list of installed apps
func (d *AppDetector) GetInstalledApps() []InstalledApp {
	return d.installedApps
}

// GetSupportedApps returns a list of known supported applications
func (d *AppDetector) GetSupportedApps() []string {
	var apps []string
	for appName := range knownApps {
		apps = append(apps, appName)
	}
	return apps
}

// detectKnownApp detects configuration for known applications
func (d *AppDetector) detectKnownApp(normalizedName string) *config.AppConfig {
	appInfo, exists := knownApps[normalizedName]
	if !exists {
		return nil
	}

	appConfig := config.NewAppConfig(appInfo.Name, appInfo.DisplayName)
	appConfig.BundleID = appInfo.BundleID

	// Add paths from the known app configuration
	for _, pathInfo := range appInfo.Paths {
		sourcePath := d.expandPath(pathInfo.Source)
		destPath := pathInfo.Destination

		// Only add path if source exists (unless it's required)
		if pathInfo.Required || util.PathExists(sourcePath) {
			appConfig.AddPath(sourcePath, destPath, pathInfo.Type, pathInfo.Required)
		}
	}

	// Only return if we found at least one valid path
	if len(appConfig.Paths) > 0 {
		return appConfig
	}

	return nil
}

// detectByBundleID attempts to detect app by bundle ID patterns
func (d *AppDetector) detectByBundleID(appName string) *config.AppConfig {
	// Common bundle ID patterns
	patterns := []string{
		"com.%s.%s",
		"org.%s.%s",
		"com.%s",
		"org.%s",
	}

	normalizedName := strings.ToLower(strings.ReplaceAll(appName, " ", ""))
	prefsDir := filepath.Join(d.homeDir, "Library", "Preferences")

	for _, pattern := range patterns {
		bundleID := fmt.Sprintf(pattern, normalizedName, normalizedName)
		plistPath := filepath.Join(prefsDir, bundleID+".plist")

		if util.PathExists(plistPath) {
			appConfig := config.NewAppConfig(normalizedName, appName)
			appConfig.BundleID = bundleID

			destPath := filepath.Join("Library", "Preferences", bundleID+".plist")
			appConfig.AddPath(plistPath, destPath, config.PathTypeFile, false)

			return appConfig
		}
	}

	return nil
}

// detectByPreferences searches for preference files matching the app name
func (d *AppDetector) detectByPreferences(appName string) *config.AppConfig {
	normalizedName := strings.ToLower(strings.ReplaceAll(appName, " ", ""))
	prefsDir := filepath.Join(d.homeDir, "Library", "Preferences")

	entries, err := os.ReadDir(prefsDir)
	if err != nil {
		return nil
	}

	var foundPaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := strings.ToLower(entry.Name())
		if strings.Contains(fileName, normalizedName) && strings.HasSuffix(fileName, ".plist") {
			foundPaths = append(foundPaths, filepath.Join(prefsDir, entry.Name()))
		}
	}

	if len(foundPaths) > 0 {
		appConfig := config.NewAppConfig(normalizedName, appName)

		for _, path := range foundPaths {
			relPath, _ := filepath.Rel(d.homeDir, path)
			appConfig.AddPath(path, relPath, config.PathTypeFile, false)
		}

		return appConfig
	}

	return nil
}

// expandPath expands ~ to home directory and other path expansions
func (d *AppDetector) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(d.homeDir, path[2:])
	}
	return path
}

// AppInfo represents information about a known application
type AppInfo struct {
	Name        string
	DisplayName string
	BundleID    string
	Paths       []PathInfo
}

// PathInfo represents a configuration path for an application
type PathInfo struct {
	Source      string
	Destination string
	Type        config.PathType
	Required    bool
}

// knownApps contains configuration information for commonly used macOS applications
var knownApps = map[string]*AppInfo{
	"vscode": {
		Name:        "vscode",
		DisplayName: "Visual Studio Code",
		BundleID:    "com.microsoft.VSCode",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Application Support/Code/User/settings.json",
				Destination: "Library/Application Support/Code/User/settings.json",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/Code/User/keybindings.json",
				Destination: "Library/Application Support/Code/User/keybindings.json",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/Code/User/snippets",
				Destination: "Library/Application Support/Code/User/snippets",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"googlechrome": {
		Name:        "googlechrome",
		DisplayName: "Google Chrome",
		BundleID:    "com.google.Chrome",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.google.Chrome.plist",
				Destination: "Library/Preferences/com.google.Chrome.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/Google/Chrome/Default/Preferences",
				Destination: "Library/Application Support/Google/Chrome/Default/Preferences",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"firefox": {
		Name:        "firefox",
		DisplayName: "Firefox",
		BundleID:    "org.mozilla.firefox",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/org.mozilla.firefox.plist",
				Destination: "Library/Preferences/org.mozilla.firefox.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/Firefox/Profiles",
				Destination: "Library/Application Support/Firefox/Profiles",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"terminal": {
		Name:        "terminal",
		DisplayName: "Terminal",
		BundleID:    "com.apple.Terminal",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.apple.Terminal.plist",
				Destination: "Library/Preferences/com.apple.Terminal.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"iterm2": {
		Name:        "iterm2",
		DisplayName: "iTerm2",
		BundleID:    "com.googlecode.iterm2",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.googlecode.iterm2.plist",
				Destination: "Library/Preferences/com.googlecode.iterm2.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"sublimetext": {
		Name:        "sublimetext",
		DisplayName: "Sublime Text",
		BundleID:    "com.sublimetext.4",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Application Support/Sublime Text/Packages/User",
				Destination: "Library/Application Support/Sublime Text/Packages/User",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"git": {
		Name:        "git",
		DisplayName: "Git",
		BundleID:    "",
		Paths: []PathInfo{
			{
				Source:      "~/.gitconfig",
				Destination: ".gitconfig",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/.gitignore_global",
				Destination: ".gitignore_global",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"ssh": {
		Name:        "ssh",
		DisplayName: "SSH",
		BundleID:    "",
		Paths: []PathInfo{
			{
				Source:      "~/.ssh/config",
				Destination: ".ssh/config",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"bartender4": {
		Name:        "bartender4",
		DisplayName: "Bartender 4",
		BundleID:    "com.surteesstudios.Bartender",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.surteesstudios.Bartender.plist",
				Destination: "Library/Preferences/com.surteesstudios.Bartender.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"rectangle": {
		Name:        "rectangle",
		DisplayName: "Rectangle",
		BundleID:    "com.knollsoft.Rectangle",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.knollsoft.Rectangle.plist",
				Destination: "Library/Preferences/com.knollsoft.Rectangle.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"alfred": {
		Name:        "alfred",
		DisplayName: "Alfred",
		BundleID:    "com.runningwithcrayons.Alfred",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.runningwithcrayons.Alfred-Preferences.plist",
				Destination: "Library/Preferences/com.runningwithcrayons.Alfred-Preferences.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/Alfred",
				Destination: "Library/Application Support/Alfred",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"1password": {
		Name:        "1password",
		DisplayName: "1Password 7 - Password Manager",
		BundleID:    "com.1password.1password",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.1password.1password.plist",
				Destination: "Library/Preferences/com.1password.1password.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Group Containers/2BUA8C4S2C.com.1password",
				Destination: "Library/Group Containers/2BUA8C4S2C.com.1password",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"1password8": {
		Name:        "1password8",
		DisplayName: "1Password 8",
		BundleID:    "com.1password.1password8",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.1password.1password8.plist",
				Destination: "Library/Preferences/com.1password.1password8.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"finder": {
		Name:        "finder",
		DisplayName: "Finder",
		BundleID:    "com.apple.finder",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.apple.finder.plist",
				Destination: "Library/Preferences/com.apple.finder.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"dock": {
		Name:        "dock",
		DisplayName: "Dock",
		BundleID:    "com.apple.dock",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.apple.dock.plist",
				Destination: "Library/Preferences/com.apple.dock.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"spotify": {
		Name:        "spotify",
		DisplayName: "Spotify",
		BundleID:    "com.spotify.client",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.spotify.client.plist",
				Destination: "Library/Preferences/com.spotify.client.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/Spotify",
				Destination: "Library/Application Support/Spotify",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"slack": {
		Name:        "slack",
		DisplayName: "Slack",
		BundleID:    "com.tinyspeck.slackmacgap",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.tinyspeck.slackmacgap.plist",
				Destination: "Library/Preferences/com.tinyspeck.slackmacgap.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/Slack",
				Destination: "Library/Application Support/Slack",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"discord": {
		Name:        "discord",
		DisplayName: "Discord",
		BundleID:    "com.hnc.Discord",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.hnc.Discord.plist",
				Destination: "Library/Preferences/com.hnc.Discord.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/Library/Application Support/discord",
				Destination: "Library/Application Support/discord",
				Type:        config.PathTypeDirectory,
				Required:    false,
			},
		},
	},
	"homebrew": {
		Name:        "homebrew",
		DisplayName: "Homebrew",
		BundleID:    "",
		Paths: []PathInfo{
			{
				Source:      "~/.zprofile",
				Destination: ".zprofile",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/.zshrc",
				Destination: ".zshrc",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/.bashrc",
				Destination: ".bashrc",
				Type:        config.PathTypeFile,
				Required:    false,
			},
			{
				Source:      "~/.bash_profile",
				Destination: ".bash_profile",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"magnet": {
		Name:        "magnet",
		DisplayName: "Magnet",
		BundleID:    "com.crowdcafe.windowmagnet",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.crowdcafe.windowmagnet.plist",
				Destination: "Library/Preferences/com.crowdcafe.windowmagnet.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
	"cleanmymac": {
		Name:        "cleanmymac",
		DisplayName: "CleanMyMac X",
		BundleID:    "com.macpaw.CleanMyMac4",
		Paths: []PathInfo{
			{
				Source:      "~/Library/Preferences/com.macpaw.CleanMyMac4.plist",
				Destination: "Library/Preferences/com.macpaw.CleanMyMac4.plist",
				Type:        config.PathTypeFile,
				Required:    false,
			},
		},
	},
}

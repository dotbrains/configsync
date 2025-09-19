package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotbrains/configsync/internal/constants"
)

func TestNewManager(t *testing.T) {
	homeDir := "/test/home"
	manager := NewManager(homeDir)

	expectedConfigDir := filepath.Join(homeDir, ".configsync")
	if manager.GetConfigDir() != expectedConfigDir {
		t.Errorf("Expected configDir %s, got %s", expectedConfigDir, manager.GetConfigDir())
	}

	expectedConfigPath := filepath.Join(expectedConfigDir, "config.yaml")
	if manager.ConfigPath() != expectedConfigPath {
		t.Errorf("Expected configPath %s, got %s", expectedConfigPath, manager.ConfigPath())
	}
}

func TestManagerWithRealFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create config directory
	err := os.MkdirAll(manager.GetConfigDir(), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Test ConfigExists with non-existent config
	if manager.ConfigExists() {
		t.Error("Expected ConfigExists to return false for non-existent config")
	}

	// Create a test configuration
	cfg := NewDefaultConfig(
		filepath.Join(tempDir, "store"),
		filepath.Join(tempDir, "backups"),
		filepath.Join(tempDir, "logs"),
	)

	// Add a test app
	testApp := NewAppConfig(constants.TestAppName, "Test Application")
	testApp.BundleID = constants.TestBundleID
	testApp.AddPath("/test/source", "test/dest", PathTypeFile, false)
	cfg.Apps[constants.TestAppName] = testApp

	// Test Save
	err = manager.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test ConfigExists with existing config
	if !manager.ConfigExists() {
		t.Error("Expected ConfigExists to return true for existing config")
	}

	// Test Load
	loadedCfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config
	if loadedCfg.Version != cfg.Version {
		t.Errorf("Expected version %s, got %s", cfg.Version, loadedCfg.Version)
	}

	if loadedCfg.StorePath != cfg.StorePath {
		t.Errorf("Expected store path %s, got %s", cfg.StorePath, loadedCfg.StorePath)
	}

	if len(loadedCfg.Apps) != 1 {
		t.Errorf("Expected 1 app, got %d", len(loadedCfg.Apps))
	}

	if _, exists := loadedCfg.Apps[constants.TestAppName]; !exists {
		t.Error("Expected testapp to exist in loaded config")
	}

	loadedApp := loadedCfg.Apps[constants.TestAppName]
	if loadedApp.DisplayName != "Test Application" {
		t.Errorf("Expected display name 'Test Application', got '%s'", loadedApp.DisplayName)
	}

	if loadedApp.BundleID != constants.TestBundleID {
		t.Errorf("Expected bundle ID '%s', got '%s'", constants.TestBundleID, loadedApp.BundleID)
	}

	if len(loadedApp.Paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(loadedApp.Paths))
	}
}

func TestManagerSaveLoadTimestamps(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create config directory
	err := os.MkdirAll(manager.GetConfigDir(), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create configuration with timestamps
	cfg := NewDefaultConfig(
		filepath.Join(tempDir, "store"),
		filepath.Join(tempDir, "backups"),
		filepath.Join(tempDir, "logs"),
	)

	// Set timestamps
	now := time.Now().Truncate(time.Second)
	cfg.CreatedAt = now
	cfg.UpdatedAt = now
	cfg.LastSync = now.Add(-1 * time.Hour)

	// Test save and load cycle for timestamps
	err = manager.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify timestamps are preserved
	if !loadedCfg.CreatedAt.Equal(cfg.CreatedAt) {
		t.Errorf("CreatedAt mismatch: expected %v, got %v", cfg.CreatedAt, loadedCfg.CreatedAt)
	}

	if !loadedCfg.LastSync.Equal(cfg.LastSync) {
		t.Errorf("LastSync mismatch: expected %v, got %v", cfg.LastSync, loadedCfg.LastSync)
	}
}

func TestManagerSaveLoadMultipleApps(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create config directory
	err := os.MkdirAll(manager.GetConfigDir(), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	cfg := NewDefaultConfig(
		filepath.Join(tempDir, "store"),
		filepath.Join(tempDir, "backups"),
		filepath.Join(tempDir, "logs"),
	)

	now := time.Now().Truncate(time.Second)

	// Add test apps
	cfg.Apps["vscode"] = createTestApp("vscode", "Visual Studio Code", "com.microsoft.VSCode", true, now)
	cfg.Apps["chrome"] = createTestApp("chrome", "Google Chrome", "com.google.Chrome", false, now)

	// Test save and load cycle
	err = manager.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify apps are preserved
	verifyAppsEqual(t, cfg.Apps, loadedCfg.Apps)
}

func TestManagerSaveLoadAppPaths(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create config directory
	err := os.MkdirAll(manager.GetConfigDir(), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	cfg := NewDefaultConfig(
		filepath.Join(tempDir, "store"),
		filepath.Join(tempDir, "backups"),
		filepath.Join(tempDir, "logs"),
	)

	now := time.Now().Truncate(time.Second)
	testApp := createTestAppWithPaths("testapp", "Test App", now)
	cfg.Apps["testapp"] = testApp

	// Test save and load cycle
	err = manager.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify paths are preserved
	loadedApp := loadedCfg.Apps["testapp"]
	verifyPathsEqual(t, testApp.Paths, loadedApp.Paths)
}

// Helper functions for test refactoring
func createTestApp(name, displayName, bundleID string, enabled bool, now time.Time) *AppConfig {
	return &AppConfig{
		Name:         name,
		DisplayName:  displayName,
		BundleID:     bundleID,
		Enabled:      enabled,
		BackupBefore: true,
		Paths:        []Path{},
		Metadata:     map[string]string{"version": "1.0.0"},
		AddedAt:      now.Add(-1 * time.Hour),
		LastSynced:   now.Add(-30 * time.Minute),
	}
}

func createTestAppWithPaths(name, displayName string, now time.Time) *AppConfig {
	return &AppConfig{
		Name:        name,
		DisplayName: displayName,
		Enabled:     true,
		Paths: []Path{
			{
				Source:      "/test/settings.json",
				Destination: "test/settings.json",
				Type:        PathTypeFile,
				Required:    false,
				BackedUp:    true,
				Synced:      true,
				SyncedAt:    now,
			},
			{
				Source:      "/test/config",
				Destination: "test/config",
				Type:        PathTypeDirectory,
				Required:    false,
				BackedUp:    false,
				Synced:      false,
			},
		},
		Metadata: map[string]string{"version": "1.80.0"},
		AddedAt:  now.Add(-2 * time.Hour),
	}
}

func verifyAppsEqual(t *testing.T, original, loaded map[string]*AppConfig) {
	for appName, originalApp := range original {
		loadedApp, exists := loaded[appName]
		if !exists {
			t.Errorf("App %s not found in loaded config", appName)
			continue
		}

		if loadedApp.Enabled != originalApp.Enabled {
			t.Errorf("App %s enabled mismatch: expected %t, got %t", appName, originalApp.Enabled, loadedApp.Enabled)
		}
	}
}

func verifyPathsEqual(t *testing.T, original, loaded []Path) {
	if len(loaded) != len(original) {
		t.Errorf("Paths count mismatch: expected %d, got %d", len(original), len(loaded))
		return
	}

	for i, originalPath := range original {
		loadedPath := loaded[i]
		if loadedPath.Source != originalPath.Source {
			t.Errorf("Path source mismatch: expected %s, got %s", originalPath.Source, loadedPath.Source)
		}
		if loadedPath.Synced != originalPath.Synced {
			t.Errorf("Path synced mismatch: expected %t, got %t", originalPath.Synced, loadedPath.Synced)
		}
		if loadedPath.BackedUp != originalPath.BackedUp {
			t.Errorf("Path backed up mismatch: expected %t, got %t", originalPath.BackedUp, loadedPath.BackedUp)
		}
	}
}

func TestManagerLoadError(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create config directory
	err := os.MkdirAll(manager.GetConfigDir(), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create an invalid YAML file
	invalidYAML := "invalid: yaml: content: [unclosed"
	err = os.WriteFile(manager.ConfigPath(), []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	// Test loading invalid YAML
	_, err = manager.Load()
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestManagerSaveError(t *testing.T) {
	// Create a directory that we don't have permission to write to
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "readonly")
	err := os.Mkdir(configDir, 0444) // Read-only directory
	if err != nil {
		t.Fatalf("Failed to create readonly directory: %v", err)
	}

	manager := NewManager(configDir)
	cfg := NewDefaultConfig("/test/store", "/test/backup", "/test/log")

	// Test save error due to permissions
	err = manager.Save(cfg)
	if err == nil {
		t.Error("Expected error when saving to readonly directory")
	}

	// Clean up by changing permissions
	_ = os.Chmod(configDir, 0755)
}

func TestManagerInitialize(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Test Initialize
	err := manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Verify directories were created
	expectedDirs := []string{
		filepath.Join(tempDir, ".configsync"),
		filepath.Join(tempDir, ".configsync", "store"),
		filepath.Join(tempDir, ".configsync", "backups"),
		filepath.Join(tempDir, ".configsync", "logs"),
		filepath.Join(tempDir, ".configsync", "store", "Library"),
		filepath.Join(tempDir, ".configsync", "store", "Library", "Preferences"),
		filepath.Join(tempDir, ".configsync", "store", "Library", "Application Support"),
		filepath.Join(tempDir, ".configsync", "store", ".config"),
	}

	for _, dir := range expectedDirs {
		if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	// Test that config file was created
	if !manager.ConfigExists() {
		t.Error("Expected config file to exist after initialize")
	}
}

func TestManagerAppOperations(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Initialize first
	err := manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test AddApp
	testApp := NewAppConfig("testapp", "Test App")
	testApp.AddPath("/test/source", "test/dest", PathTypeFile, false)
	err = manager.AddApp(testApp)
	if err != nil {
		t.Fatalf("Failed to add app: %v", err)
	}

	// Test GetApp
	retrievedApp, err := manager.GetApp("testapp")
	if err != nil {
		t.Fatalf("Failed to get app: %v", err)
	}
	if retrievedApp.Name != "testapp" {
		t.Errorf("Expected app name 'testapp', got '%s'", retrievedApp.Name)
	}

	// Test GetApp with non-existent app
	_, err = manager.GetApp("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent app")
	}

	// Test ListApps
	apps, err := manager.ListApps()
	if err != nil {
		t.Fatalf("Failed to list apps: %v", err)
	}
	if len(apps) != 1 {
		t.Errorf("Expected 1 app, got %d", len(apps))
	}
	if _, exists := apps["testapp"]; !exists {
		t.Error("Expected testapp to exist in apps list")
	}

	// Test RemoveApp
	err = manager.RemoveApp("testapp")
	if err != nil {
		t.Fatalf("Failed to remove app: %v", err)
	}

	// Verify app was removed
	apps, err = manager.ListApps()
	if err != nil {
		t.Fatalf("Failed to list apps after removal: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("Expected 0 apps after removal, got %d", len(apps))
	}
}

func TestManagerPaths(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Initialize first
	err := manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test GetStorePath
	storePath, err := manager.GetStorePath()
	if err != nil {
		t.Fatalf("Failed to get store path: %v", err)
	}
	expectedStorePath := filepath.Join(tempDir, ".configsync", "store")
	if storePath != expectedStorePath {
		t.Errorf("Expected store path %s, got %s", expectedStorePath, storePath)
	}

	// Test GetBackupPath
	backupPath, err := manager.GetBackupPath()
	if err != nil {
		t.Fatalf("Failed to get backup path: %v", err)
	}
	expectedBackupPath := filepath.Join(tempDir, ".configsync", "backups")
	if backupPath != expectedBackupPath {
		t.Errorf("Expected backup path %s, got %s", expectedBackupPath, backupPath)
	}
}

func TestManagerSettings(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Initialize first
	err := manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test GetSettings
	settings, err := manager.GetSettings()
	if err != nil {
		t.Fatalf("Failed to get settings: %v", err)
	}
	if !settings.AutoBackup {
		t.Error("Expected AutoBackup to be true by default")
	}

	// Test UpdateSettings
	settings.AutoBackup = false
	settings.DryRun = true
	err = manager.UpdateSettings(settings)
	if err != nil {
		t.Fatalf("Failed to update settings: %v", err)
	}

	// Verify settings were updated
	updatedSettings, err := manager.GetSettings()
	if err != nil {
		t.Fatalf("Failed to get updated settings: %v", err)
	}
	if updatedSettings.AutoBackup {
		t.Error("Expected AutoBackup to be false after update")
	}
	if !updatedSettings.DryRun {
		t.Error("Expected DryRun to be true after update")
	}
}

func TestManagerSyncOperations(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Initialize first
	err := manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test UpdateLastSync
	err = manager.UpdateLastSync()
	if err != nil {
		t.Fatalf("Failed to update last sync: %v", err)
	}

	// Verify last sync was updated
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config after sync update: %v", err)
	}
	if cfg.LastSync.IsZero() {
		t.Error("Expected LastSync to be set")
	}
}

func TestManagerBackupIntegration(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create config directory
	err := os.MkdirAll(manager.GetConfigDir(), 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create initial config
	cfg := NewDefaultConfig(
		filepath.Join(tempDir, "store"),
		filepath.Join(tempDir, "backups"),
		filepath.Join(tempDir, "logs"),
	)

	testApp := NewAppConfig("testapp", "Test App")
	testApp.AddPath("/test/path", "test/path", PathTypeFile, false)
	cfg.Apps["testapp"] = testApp

	// Save initial config
	err = manager.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Modify and save again (simulating backup scenario)
	cfg.Apps["testapp"].Enabled = false
	cfg.UpdatedAt = time.Now()

	err = manager.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save modified config: %v", err)
	}

	// Load and verify the changes were saved
	loadedCfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load modified config: %v", err)
	}

	if loadedCfg.Apps["testapp"].Enabled != false {
		t.Error("Expected testapp to be disabled after modification")
	}
}

func TestPathTypeConstants(t *testing.T) {
	// Test that path type constants have expected values
	if PathTypeFile != "file" {
		t.Errorf("Expected PathTypeFile to be 'file', got '%s'", PathTypeFile)
	}

	if PathTypeDirectory != "directory" {
		t.Errorf("Expected PathTypeDirectory to be 'directory', got '%s'", PathTypeDirectory)
	}

	if PathTypeGlob != "glob" {
		t.Errorf("Expected PathTypeGlob to be 'glob', got '%s'", PathTypeGlob)
	}
}

// Benchmark the save/load cycle
func BenchmarkManagerSaveLoad(b *testing.B) {
	tempDir := b.TempDir()
	manager := NewManager(tempDir)

	// Create config directory
	err := os.MkdirAll(manager.GetConfigDir(), 0755)
	if err != nil {
		b.Fatalf("Failed to create config directory: %v", err)
	}

	cfg := NewDefaultConfig(
		filepath.Join(tempDir, "store"),
		filepath.Join(tempDir, "backups"),
		filepath.Join(tempDir, "logs"),
	)

	// Add some apps to make it realistic
	for i := 0; i < 10; i++ {
		appName := "testapp" + string(rune(i))
		app := NewAppConfig(appName, "Test App "+string(rune(i)))
		app.AddPath("/test/path"+string(rune(i)), "path"+string(rune(i)), PathTypeFile, false)
		cfg.Apps[appName] = app
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := manager.Save(cfg)
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}

		_, err = manager.Load()
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

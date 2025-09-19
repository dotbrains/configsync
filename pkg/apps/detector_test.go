package apps

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dotbrains/configsync/internal/config"
)

func TestNewAppDetector(t *testing.T) {
	homeDir := "/test/home"
	detector := NewAppDetector(homeDir)

	if detector.homeDir != homeDir {
		t.Errorf("Expected homeDir %s, got %s", homeDir, detector.homeDir)
	}

	if detector.installedApps == nil {
		t.Error("Expected installedApps to be initialized")
	}

	if detector.cacheDuration != 5*time.Minute {
		t.Errorf("Expected cacheDuration 5 minutes, got %v", detector.cacheDuration)
	}
}

func TestGenerateAppKey(t *testing.T) {
	detector := NewAppDetector("/test/home")

	tests := []struct {
		name     string
		app      InstalledApp
		expected string
	}{
		{
			name: "App with bundle ID",
			app: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
				Path:     "/Applications/Test.app",
			},
			expected: "bundle:com.test.app",
		},
		{
			name: "App without bundle ID but with path",
			app: InstalledApp{
				Name: "testapp",
				Path: "/Applications/Test.app",
			},
			expected: "path:/Applications/Test.app",
		},
		{
			name: "App with only name",
			app: InstalledApp{
				Name: "testapp",
			},
			expected: "name:testapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.generateAppKey(tt.app)
			if result != tt.expected {
				t.Errorf("Expected key %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestShouldReplaceApp(t *testing.T) {
	detector := NewAppDetector("/test/home")

	tests := []struct {
		name     string
		existing InstalledApp
		new      InstalledApp
		expected bool
	}{
		{
			name: "New app has bundle ID, existing doesn't",
			existing: InstalledApp{
				Name: "testapp",
			},
			new: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
			},
			expected: true,
		},
		{
			name: "Existing app has bundle ID, new doesn't",
			existing: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
			},
			new: InstalledApp{
				Name: "testapp",
			},
			expected: false,
		},
		{
			name: "New app has version, existing doesn't",
			existing: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
			},
			new: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
				Version:  "1.0",
			},
			expected: true,
		},
		{
			name: "New app is in /Applications, existing isn't",
			existing: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
				Version:  "1.0",
				Path:     "/Users/test/Applications/Test.app",
			},
			new: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
				Version:  "1.0",
				Path:     "/Applications/Test.app",
			},
			expected: true,
		},
		{
			name: "Both apps equal, prefer existing",
			existing: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
				Version:  "1.0",
				Path:     "/Applications/Test.app",
			},
			new: InstalledApp{
				Name:     "testapp",
				BundleID: "com.test.app",
				Version:  "1.0",
				Path:     "/Applications/Test.app",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.shouldReplaceApp(tt.existing, tt.new)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestRemoveDuplicateApps(t *testing.T) {
	detector := NewAppDetector("/test/home")

	apps := []InstalledApp{
		{
			Name:     "testapp1",
			BundleID: "com.test.app1",
			Path:     "/Applications/Test1.app",
		},
		{
			Name:     "testapp1", // Duplicate by bundle ID
			BundleID: "com.test.app1",
			Path:     "/Applications/Test1.app",
			Version:  "1.0", // Has version, should replace first
		},
		{
			Name: "testapp2",
			Path: "/Applications/Test2.app",
		},
		{
			Name: "testapp2", // Duplicate by path
			Path: "/Applications/Test2.app",
		},
		{
			Name: "testapp3", // Unique
			Path: "/Applications/Test3.app",
		},
	}

	result := detector.removeDuplicateApps(apps)

	if len(result) != 3 {
		t.Errorf("Expected 3 unique apps, got %d", len(result))
	}

	// Check that the app with version was kept for testapp1
	for _, app := range result {
		if app.Name == "testapp1" && app.Version != "1.0" {
			t.Error("Expected version 1.0 for testapp1, indicating the better duplicate was kept")
		}
	}
}

func TestGenerateConfigKey(t *testing.T) {
	detector := NewAppDetector("/test/home")

	tests := []struct {
		name     string
		config   *config.AppConfig
		expected string
	}{
		{
			name: "Config with bundle ID",
			config: &config.AppConfig{
				Name:     "testapp",
				BundleID: "com.test.app",
			},
			expected: "bundle:com.test.app",
		},
		{
			name: "Config without bundle ID",
			config: &config.AppConfig{
				Name: "testapp",
			},
			expected: "name:testapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.generateConfigKey(tt.config)
			if result != tt.expected {
				t.Errorf("Expected key %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRemoveDuplicateConfigs(t *testing.T) {
	detector := NewAppDetector("/test/home")

	configs := []*config.AppConfig{
		{
			Name:     "testapp1",
			BundleID: "com.test.app1",
			Paths: []config.ConfigPath{
				{Source: "/test/path1", Destination: "path1"},
			},
		},
		{
			Name:     "testapp1", // Duplicate by bundle ID but with more paths
			BundleID: "com.test.app1",
			Paths: []config.ConfigPath{
				{Source: "/test/path1", Destination: "path1"},
				{Source: "/test/path2", Destination: "path2"},
			},
		},
		{
			Name: "testapp2", // No bundle ID
			Paths: []config.ConfigPath{
				{Source: "/test/path3", Destination: "path3"},
			},
		},
		{
			Name:     "testapp2", // Same name but with bundle ID, should be preferred
			BundleID: "com.test.app2",
			Paths: []config.ConfigPath{
				{Source: "/test/path3", Destination: "path3"},
			},
		},
		{
			Name: "testapp3", // Unique app
			Paths: []config.ConfigPath{
				{Source: "/test/path4", Destination: "path4"},
			},
		},
	}

	result := detector.removeDuplicateConfigs(configs)

	// We expect 4 configs because:
	// - testapp1 with bundle ID (deduplicated to the one with 2 paths)
	// - testapp2 without bundle ID (unique by name key)
	// - testapp2 with bundle ID (unique by bundle key)
	// - testapp3 without bundle ID (unique)
	if len(result) != 4 {
		t.Errorf("Expected 4 configs (no deduplication between different key types), got %d", len(result))
	}

	// Check that config with more paths was kept for testapp1
	testapp1Found := false
	for _, cfg := range result {
		if cfg.Name == "testapp1" && cfg.BundleID == "com.test.app1" {
			if len(cfg.Paths) != 2 {
				t.Error("Expected config with 2 paths for testapp1, indicating the better duplicate was kept")
			}
			testapp1Found = true
		}
	}
	if !testapp1Found {
		t.Error("Expected to find deduplicated testapp1 config")
	}
}

func TestDetectKnownApp(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test preference files
	prefsDir := filepath.Join(tempDir, "Library", "Preferences")
	err := os.MkdirAll(prefsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create prefs directory: %v", err)
	}

	// Create VS Code directory structure
	vscodeAppSupportDir := filepath.Join(tempDir, "Library", "Application Support", "Code", "User")
	err = os.MkdirAll(vscodeAppSupportDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create VS Code app support directory: %v", err)
	}

	// Create test VS Code files that match the known configuration
	vscodeSettings := filepath.Join(vscodeAppSupportDir, "settings.json")
	_, err = os.Create(vscodeSettings)
	if err != nil {
		t.Fatalf("Failed to create VS Code settings file: %v", err)
	}

	detector := NewAppDetector(tempDir)

	// Test detecting VS Code
	appConfig := detector.detectKnownApp("vscode")
	if appConfig == nil {
		t.Error("Expected to detect VS Code configuration")
		return
	}

	if appConfig.Name != "vscode" {
		t.Errorf("Expected app name 'vscode', got '%s'", appConfig.Name)
	}

	if appConfig.DisplayName != "Visual Studio Code" {
		t.Errorf("Expected display name 'Visual Studio Code', got '%s'", appConfig.DisplayName)
	}

	if appConfig.BundleID != "com.microsoft.VSCode" {
		t.Errorf("Expected bundle ID 'com.microsoft.VSCode', got '%s'", appConfig.BundleID)
	}

	// Should have at least one path (the one we created)
	if len(appConfig.Paths) == 0 {
		t.Error("Expected at least one configuration path")
	}

	// Test app that doesn't exist
	nonExistentConfig := detector.detectKnownApp("nonexistentapp")
	if nonExistentConfig != nil {
		t.Error("Expected nil for non-existent app")
	}
}

func TestSmartDetectApp(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test preference files
	prefsDir := filepath.Join(tempDir, "Library", "Preferences")
	err := os.MkdirAll(prefsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create prefs directory: %v", err)
	}

	appSupportDir := filepath.Join(tempDir, "Library", "Application Support", "TestApp")
	err = os.MkdirAll(appSupportDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create app support directory: %v", err)
	}

	// Create test files
	testPlist := filepath.Join(prefsDir, "com.test.app.plist")
	_, err = os.Create(testPlist)
	if err != nil {
		t.Fatalf("Failed to create test plist: %v", err)
	}

	testAppSupportFile := filepath.Join(appSupportDir, "config.json")
	_, err = os.Create(testAppSupportFile)
	if err != nil {
		t.Fatalf("Failed to create app support file: %v", err)
	}

	detector := NewAppDetector(tempDir)

	testApp := InstalledApp{
		Name:        "testapp",
		DisplayName: "TestApp",
		BundleID:    "com.test.app",
		Path:        "/Applications/TestApp.app",
	}

	appConfig := detector.smartDetectApp(testApp)

	if appConfig == nil {
		t.Fatal("Expected to detect app configuration")
	}

	if appConfig.Name != "testapp" {
		t.Errorf("Expected app name 'testapp', got '%s'", appConfig.Name)
	}

	if appConfig.BundleID != "com.test.app" {
		t.Errorf("Expected bundle ID 'com.test.app', got '%s'", appConfig.BundleID)
	}

	// Should detect both preference file and application support directory
	if len(appConfig.Paths) < 1 {
		t.Error("Expected at least 1 configuration path")
	}

	// Check that we found the preference file
	foundPref := false
	foundAppSupport := false
	for _, path := range appConfig.Paths {
		if strings.Contains(path.Source, "com.test.app.plist") {
			foundPref = true
		}
		if strings.Contains(path.Source, "Application Support/TestApp") {
			foundAppSupport = true
		}
	}

	if !foundPref {
		t.Error("Expected to find preference file path")
	}
	if !foundAppSupport {
		t.Error("Expected to find application support path")
	}
}

func TestGetSupportedApps(t *testing.T) {
	detector := NewAppDetector("/test/home")

	supportedApps := detector.GetSupportedApps()

	if len(supportedApps) == 0 {
		t.Error("Expected some supported apps")
	}

	// Check for some apps we know should be in the list
	expectedApps := []string{"vscode", "googlechrome", "firefox", "terminal", "git"}
	for _, expectedApp := range expectedApps {
		found := false
		for _, supportedApp := range supportedApps {
			if supportedApp == expectedApp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find %s in supported apps list", expectedApp)
		}
	}
}

func TestExpandPath(t *testing.T) {
	detector := NewAppDetector("/test/home")

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Home directory path",
			path:     "~/Library/Preferences/test.plist",
			expected: "/test/home/Library/Preferences/test.plist",
		},
		{
			name:     "Absolute path",
			path:     "/Applications/Test.app",
			expected: "/Applications/Test.app",
		},
		{
			name:     "Relative path",
			path:     "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.expandPath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected path %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCaching(t *testing.T) {
	detector := NewAppDetector("/test/home")

	// Set a short cache duration for testing
	detector.cacheDuration = 100 * time.Millisecond

	// Mock some installed apps
	testApps := []InstalledApp{
		{Name: "testapp1", BundleID: "com.test.1"},
		{Name: "testapp2", BundleID: "com.test.2"},
	}

	// Manually set cache
	detector.installedApps = testApps
	detector.lastScanTime = time.Now()

	// Should return cached results
	result, _ := detector.ScanInstalledApps()
	if !reflect.DeepEqual(result, testApps) {
		t.Error("Expected cached results to be returned")
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Now it should try to scan again (will fail but cache should be expired)
	detector.lastScanTime = time.Now().Add(-10 * time.Minute) // Force cache expiry
	detector.installedApps = []InstalledApp{}                 // Clear cache

	// This will call the actual scan methods which may fail in test environment,
	// but we're testing the cache logic
	_, _ = detector.ScanInstalledApps()

	// The cache should be updated (even if empty due to scan failures)
	if time.Since(detector.lastScanTime) > time.Second {
		t.Error("Expected cache timestamp to be updated")
	}
}

// Benchmark tests for performance
func BenchmarkGenerateAppKey(b *testing.B) {
	detector := NewAppDetector("/test/home")
	app := InstalledApp{
		Name:     "testapp",
		BundleID: "com.test.app",
		Path:     "/Applications/Test.app",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.generateAppKey(app)
	}
}

func BenchmarkRemoveDuplicateApps(b *testing.B) {
	detector := NewAppDetector("/test/home")

	// Create a list of apps with some duplicates
	apps := make([]InstalledApp, 1000)
	for i := 0; i < 1000; i++ {
		apps[i] = InstalledApp{
			Name:     "testapp" + string(rune(i%100)), // Creates duplicates
			BundleID: "com.test.app" + string(rune(i%100)),
			Path:     "/Applications/Test" + string(rune(i)) + ".app",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.removeDuplicateApps(apps)
	}
}

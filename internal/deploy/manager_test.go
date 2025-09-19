package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotbrains/configsync/internal/config"
)

func TestNewManager(t *testing.T) {
	homeDir := "/test/home"
	storeDir := "/test/store"
	backupDir := "/test/backup"
	verbose := true

	manager := NewManager(homeDir, storeDir, backupDir, verbose)

	if manager.homeDir != homeDir {
		t.Errorf("Expected homeDir %s, got %s", homeDir, manager.homeDir)
	}

	if manager.storeDir != storeDir {
		t.Errorf("Expected storeDir %s, got %s", storeDir, manager.storeDir)
	}

	if manager.backupDir != backupDir {
		t.Errorf("Expected backupDir %s, got %s", backupDir, manager.backupDir)
	}

	if manager.verbose != verbose {
		t.Errorf("Expected verbose %t, got %t", verbose, manager.verbose)
	}
}

func TestExportBundle(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")
	configDir := filepath.Join(tempDir, "config")

	// Set up directories
	err := os.MkdirAll(storeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create store dir: %v", err)
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	manager := NewManager(homeDir, storeDir, backupDir, true)

	// Create test configuration files in store
	testFile1 := filepath.Join(storeDir, "test1.conf")
	testFile2 := filepath.Join(storeDir, "test2.conf")
	err = os.WriteFile(testFile1, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	err = os.WriteFile(testFile2, []byte("test content 2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Create config manager and add test apps
	configManager := config.NewManager(homeDir)
	err = configManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

	// Create test app configs
	app1 := config.NewAppConfig("testapp1", "Test App 1")
	app1.AddPath("/test/source1.conf", "test1.conf", config.PathTypeFile, false)

	app2 := config.NewAppConfig("testapp2", "Test App 2")
	app2.AddPath("/test/source2.conf", "test2.conf", config.PathTypeFile, false)

	err = configManager.AddApp(app1)
	if err != nil {
		t.Fatalf("Failed to add app 1: %v", err)
	}

	err = configManager.AddApp(app2)
	if err != nil {
		t.Fatalf("Failed to add app 2: %v", err)
	}

	// Test export all apps
	bundlePath := filepath.Join(tempDir, "test-bundle.tar.gz")
	err = manager.ExportBundle(bundlePath, []string{}, configManager)
	if err != nil {
		t.Fatalf("ExportBundle failed: %v", err)
	}

	// Verify bundle file exists
	if !manager.pathExists(bundlePath) {
		t.Error("Bundle file should exist")
	}

	// Verify bundle is not empty
	size, err := manager.getFileSize(bundlePath)
	if err != nil {
		t.Fatalf("Failed to get bundle size: %v", err)
	}
	if size == 0 {
		t.Error("Bundle should not be empty")
	}
}

func TestExportBundleSpecificApps(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")
	configDir := filepath.Join(tempDir, "config")

	// Set up directories
	err := os.MkdirAll(storeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create store dir: %v", err)
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	manager := NewManager(homeDir, storeDir, backupDir, false)

	// Create test configuration files in store
	testFile1 := filepath.Join(storeDir, "test1.conf")
	err = os.WriteFile(testFile1, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	// Create config manager and add test apps
	configManager := config.NewManager(homeDir)
	err = configManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

	app1 := config.NewAppConfig("testapp1", "Test App 1")
	app1.AddPath("/test/source1.conf", "test1.conf", config.PathTypeFile, false)

	err = configManager.AddApp(app1)
	if err != nil {
		t.Fatalf("Failed to add app 1: %v", err)
	}

	// Test export specific app
	bundlePath := filepath.Join(tempDir, "specific-bundle.tar.gz")
	err = manager.ExportBundle(bundlePath, []string{"testapp1"}, configManager)
	if err != nil {
		t.Fatalf("ExportBundle failed: %v", err)
	}

	// Verify bundle file exists
	if !manager.pathExists(bundlePath) {
		t.Error("Bundle file should exist")
	}
}

func TestExportBundleNonExistentApp(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")
	configDir := filepath.Join(tempDir, "config")

	// Set up directories
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	manager := NewManager(homeDir, storeDir, backupDir, false)
	configManager := config.NewManager(homeDir)
	err = configManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

	// Test export non-existent app
	bundlePath := filepath.Join(tempDir, "nonexistent-bundle.tar.gz")
	err = manager.ExportBundle(bundlePath, []string{"nonexistent"}, configManager)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}

	if !strings.Contains(err.Error(), "application not found") {
		t.Errorf("Expected 'application not found' error, got: %v", err)
	}
}

func TestExportBundleNoApps(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")
	configDir := filepath.Join(tempDir, "config")

	// Set up directories
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	manager := NewManager(homeDir, storeDir, backupDir, false)
	configManager := config.NewManager(homeDir)
	err = configManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

	// Test export with no apps configured
	bundlePath := filepath.Join(tempDir, "empty-bundle.tar.gz")
	err = manager.ExportBundle(bundlePath, []string{}, configManager)
	if err == nil {
		t.Error("Expected error when no apps to export")
	}

	if !strings.Contains(err.Error(), "no applications to export") {
		t.Errorf("Expected 'no applications to export' error, got: %v", err)
	}
}

func TestImportBundle(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")
	configDir := filepath.Join(tempDir, "config")

	// Set up directories
	err := os.MkdirAll(storeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create store dir: %v", err)
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	manager := NewManager(homeDir, storeDir, backupDir, true)

	// First create a bundle to import
	testFile1 := filepath.Join(storeDir, "test1.conf")
	err = os.WriteFile(testFile1, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	configManager := config.NewManager(homeDir)
	err = configManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}
	app1 := config.NewAppConfig("testapp1", "Test App 1")
	app1.AddPath("/test/source1.conf", "test1.conf", config.PathTypeFile, false)

	err = configManager.AddApp(app1)
	if err != nil {
		t.Fatalf("Failed to add app 1: %v", err)
	}

	bundlePath := filepath.Join(tempDir, "test-bundle.tar.gz")
	err = manager.ExportBundle(bundlePath, []string{}, configManager)
	if err != nil {
		t.Fatalf("ExportBundle failed: %v", err)
	}

	// Now test importing the bundle
	importDir := filepath.Join(tempDir, "import")
	bundle, err := manager.ImportBundle(bundlePath, importDir)
	if err != nil {
		t.Fatalf("ImportBundle failed: %v", err)
	}

	// Verify bundle metadata
	if bundle == nil {
		t.Fatal("Expected bundle metadata")
	}

	if bundle.Version == "" {
		t.Error("Bundle should have a version")
	}

	if len(bundle.Apps) == 0 {
		t.Error("Bundle should contain apps")
	}

	if _, exists := bundle.Apps["testapp1"]; !exists {
		t.Error("Bundle should contain testapp1")
	}

	// Verify files were extracted
	bundleYaml := filepath.Join(importDir, "bundle.yaml")
	if !manager.pathExists(bundleYaml) {
		t.Error("bundle.yaml should exist in import directory")
	}

	filesDir := filepath.Join(importDir, "files")
	if !manager.pathExists(filesDir) {
		t.Error("files directory should exist in import directory")
	}
}

func TestImportBundleNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false)

	// Test importing non-existent bundle
	bundlePath := filepath.Join(tempDir, "nonexistent-bundle.tar.gz")
	importDir := filepath.Join(tempDir, "import")

	_, err := manager.ImportBundle(bundlePath, importDir)
	if err == nil {
		t.Error("Expected error for non-existent bundle")
	}

	if !strings.Contains(err.Error(), "bundle file not found") {
		t.Errorf("Expected 'bundle file not found' error, got: %v", err)
	}
}

func TestDeployBundle(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")
	configDir := filepath.Join(tempDir, "config")

	// Set up directories
	err := os.MkdirAll(storeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create store dir: %v", err)
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	manager := NewManager(homeDir, storeDir, backupDir, true)

	// Create and export a test bundle
	testFile1 := filepath.Join(storeDir, "test1.conf")
	err = os.WriteFile(testFile1, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	configManager := config.NewManager(homeDir)
	err = configManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}
	app1 := config.NewAppConfig("testapp1", "Test App 1")
	app1.AddPath("/test/source1.conf", "test1.conf", config.PathTypeFile, false)

	err = configManager.AddApp(app1)
	if err != nil {
		t.Fatalf("Failed to add app 1: %v", err)
	}

	bundlePath := filepath.Join(tempDir, "test-bundle.tar.gz")
	err = manager.ExportBundle(bundlePath, []string{}, configManager)
	if err != nil {
		t.Fatalf("ExportBundle failed: %v", err)
	}

	// Import the bundle
	importDir := filepath.Join(tempDir, "import")
	bundle, err := manager.ImportBundle(bundlePath, importDir)
	if err != nil {
		t.Fatalf("ImportBundle failed: %v", err)
	}

	// Clear existing configuration to test deployment
	_ = os.RemoveAll(configDir)
	_ = os.MkdirAll(configDir, 0755)
	_ = os.RemoveAll(storeDir)
	_ = os.MkdirAll(storeDir, 0755)

	newConfigManager := config.NewManager(homeDir)
	err = newConfigManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize new config manager: %v", err)
	}

	// Deploy the bundle
	err = manager.DeployBundle(bundle, importDir, newConfigManager, false)
	if err != nil {
		t.Fatalf("DeployBundle failed: %v", err)
	}

	// Verify app was added to configuration
	cfg, err := newConfigManager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration after deployment: %v", err)
	}

	if _, exists := cfg.Apps["testapp1"]; !exists {
		t.Error("Deployed app should exist in configuration")
	}

	// Verify files were copied to store
	deployedFile := filepath.Join(storeDir, "test1.conf")
	if !manager.pathExists(deployedFile) {
		t.Error("Deployed file should exist in store")
	}

	content, err := os.ReadFile(deployedFile)
	if err != nil {
		t.Fatalf("Failed to read deployed file: %v", err)
	}

	if string(content) != "test content 1" {
		t.Errorf("Deployed file content mismatch: expected 'test content 1', got '%s'", string(content))
	}
}

func TestDeployBundleWithConflicts(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")
	configDir := filepath.Join(tempDir, "config")

	// Set up directories
	err := os.MkdirAll(storeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create store dir: %v", err)
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	manager := NewManager(homeDir, storeDir, backupDir, false)
	configManager := config.NewManager(homeDir)
	err = configManager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

	// Create existing app configuration
	existingApp := config.NewAppConfig("testapp1", "Test App 1")
	existingApp.LastSynced = time.Now() // Recent sync time
	err = configManager.AddApp(existingApp)
	if err != nil {
		t.Fatalf("Failed to add existing app: %v", err)
	}

	// Create bundle with same app but older timestamp
	bundle := &config.DeploymentBundle{
		Version:   "1.0",
		CreatedAt: time.Now().Add(-24 * time.Hour), // Bundle is 1 day old
		CreatedBy: "test",
		Apps: map[string]*config.AppConfig{
			"testapp1": {
				Name:        "testapp1",
				DisplayName: "Test App 1",
				Enabled:     true,
				Paths:       []config.Path{},
				LastSynced:  time.Now().Add(-24 * time.Hour),
			},
		},
		Metadata: map[string]string{},
	}

	// Test deploy without force (should fail due to conflicts)
	bundleDir := tempDir
	err = manager.DeployBundle(bundle, bundleDir, configManager, false)
	if err == nil {
		t.Error("Expected deployment to fail due to conflicts")
	}

	if !strings.Contains(err.Error(), "use --force to override conflicts") {
		t.Errorf("Expected conflict error, got: %v", err)
	}

	// Test deploy with force (should succeed)
	err = manager.DeployBundle(bundle, bundleDir, configManager, true)
	if err != nil {
		t.Fatalf("DeployBundle with force failed: %v", err)
	}
}

func TestDetectConflicts(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false)

	// Create current config with newer timestamp
	currentCfg := &config.Config{
		Apps: map[string]*config.AppConfig{
			"testapp1": {
				Name:        "testapp1",
				DisplayName: "Test App 1",
				Enabled:     true,
				LastSynced:  time.Now(),
				Paths:       []config.Path{{Source: "path1"}, {Source: "path2"}},
			},
		},
	}

	// Create bundle with older timestamp and different path count
	bundle := &config.DeploymentBundle{
		Version:   "1.0",
		CreatedAt: time.Now().Add(-24 * time.Hour), // Bundle is 1 day old
		Apps: map[string]*config.AppConfig{
			"testapp1": {
				Name:        "testapp1",
				DisplayName: "Test App 1",
				Enabled:     true,
				LastSynced:  time.Now().Add(-24 * time.Hour),
				Paths:       []config.Path{{Source: "path1"}}, // Different path count
			},
		},
	}

	conflicts := manager.detectConflicts(bundle, currentCfg)

	if len(conflicts) == 0 {
		t.Error("Expected conflicts to be detected")
	}

	// Should detect both timestamp and path count conflicts
	if len(conflicts) < 2 {
		t.Errorf("Expected at least 2 conflicts, got %d", len(conflicts))
	}

	// Check conflict messages
	foundTimestampConflict := false
	foundPathCountConflict := false

	for _, conflict := range conflicts {
		if conflict.AppName != "testapp1" {
			t.Errorf("Unexpected app name in conflict: %s", conflict.AppName)
		}

		if strings.Contains(conflict.Message, "local configuration is newer") {
			foundTimestampConflict = true
		}

		if strings.Contains(conflict.Message, "path count differs") {
			foundPathCountConflict = true
		}
	}

	if !foundTimestampConflict {
		t.Error("Expected timestamp conflict")
	}

	if !foundPathCountConflict {
		t.Error("Expected path count conflict")
	}
}

func TestPathExists(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false)

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !manager.pathExists(testFile) {
		t.Error("Should recognize existing file")
	}

	// Test non-existent file
	if manager.pathExists(filepath.Join(tempDir, "nonexistent.txt")) {
		t.Error("Should not recognize non-existent file")
	}
}

func TestGetUserInfo(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false)

	// Test with USER env var set
	originalUser := os.Getenv("USER")
	defer func() { _ = os.Setenv("USER", originalUser) }()

	_ = os.Setenv("USER", "testuser")
	userInfo := manager.getUserInfo()
	if userInfo != "testuser" {
		t.Errorf("Expected 'testuser', got '%s'", userInfo)
	}

	// Test with USER env var unset
	_ = os.Unsetenv("USER")
	userInfo = manager.getUserInfo()
	if userInfo != "unknown" {
		t.Errorf("Expected 'unknown', got '%s'", userInfo)
	}
}

func TestGetSystemInfo(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false)

	systemInfo := manager.getSystemInfo()
	if systemInfo == "" {
		t.Error("System info should not be empty")
	}

	// Should contain hostname (or at least some text)
	if len(systemInfo) < 1 {
		t.Error("System info should contain some information")
	}
}

func TestGetFileSize(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false)

	// Create test file with known content
	testFile := filepath.Join(tempDir, "test.txt")
	content := "Hello, World!"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test getting file size
	size, err := manager.getFileSize(testFile)
	if err != nil {
		t.Fatalf("Failed to get file size: %v", err)
	}

	expectedSize := int64(len(content))
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}

	// Test non-existent file
	_, err = manager.getFileSize(filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false)

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "test content"
	err := os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tempDir, "destination.txt")
	err = manager.copyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	// Verify destination file exists and has correct content
	if !manager.pathExists(dstFile) {
		t.Error("Destination file should exist")
	}

	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != content {
		t.Errorf("Content mismatch: expected '%s', got '%s'", content, string(dstContent))
	}

	// Verify permissions are copied
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		t.Fatalf("Failed to stat source file: %v", err)
	}

	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permissions mismatch: expected %v, got %v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyDir(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false)

	// Create source directory structure
	srcDir := filepath.Join(tempDir, "source")
	err := os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create files in source directory
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "subdir", "file2.txt")

	err = os.WriteFile(file1, []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(file2), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	err = os.WriteFile(file2, []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Copy directory
	dstDir := filepath.Join(tempDir, "destination")
	err = manager.copyDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("Failed to copy directory: %v", err)
	}

	// Verify destination structure
	dstFile1 := filepath.Join(dstDir, "file1.txt")
	dstFile2 := filepath.Join(dstDir, "subdir", "file2.txt")

	if !manager.pathExists(dstFile1) {
		t.Error("Destination file1 should exist")
	}

	if !manager.pathExists(dstFile2) {
		t.Error("Destination file2 should exist")
	}

	// Verify content
	content1, err := os.ReadFile(dstFile1)
	if err != nil {
		t.Fatalf("Failed to read destination file1: %v", err)
	}

	if string(content1) != "content1" {
		t.Errorf("File1 content mismatch: expected 'content1', got '%s'", string(content1))
	}

	content2, err := os.ReadFile(dstFile2)
	if err != nil {
		t.Fatalf("Failed to read destination file2: %v", err)
	}

	if string(content2) != "content2" {
		t.Errorf("File2 content mismatch: expected 'content2', got '%s'", string(content2))
	}
}

func TestValidateBundle(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false)

	// Create valid bundle structure
	filesDir := filepath.Join(tempDir, "files")
	appFilesDir := filepath.Join(filesDir, "testapp")
	err := os.MkdirAll(appFilesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create bundle structure: %v", err)
	}

	// Create required file
	requiredFile := filepath.Join(appFilesDir, "required.conf")
	err = os.WriteFile(requiredFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create required file: %v", err)
	}

	// Test valid bundle
	validBundle := &config.DeploymentBundle{
		Version: "1.0",
		Apps: map[string]*config.AppConfig{
			"testapp": {
				Name: "testapp",
				Paths: []config.Path{
					{Destination: "required.conf", Required: true},
					{Destination: "optional.conf", Required: false},
				},
			},
		},
	}

	err = manager.validateBundle(validBundle, tempDir)
	if err != nil {
		t.Errorf("Valid bundle should pass validation: %v", err)
	}

	// Test bundle with missing version
	invalidBundle := &config.DeploymentBundle{
		Version: "", // Missing version
		Apps:    map[string]*config.AppConfig{},
	}

	err = manager.validateBundle(invalidBundle, tempDir)
	if err == nil {
		t.Error("Bundle with missing version should fail validation")
	}

	if !strings.Contains(err.Error(), "missing bundle version") {
		t.Errorf("Expected 'missing bundle version' error, got: %v", err)
	}

	// Test bundle with missing files directory
	invalidBundleDir := filepath.Join(tempDir, "invalid")
	err = os.MkdirAll(invalidBundleDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create invalid bundle dir: %v", err)
	}

	err = manager.validateBundle(validBundle, invalidBundleDir)
	if err == nil {
		t.Error("Bundle with missing files directory should fail validation")
	}

	if !strings.Contains(err.Error(), "bundle files directory missing") {
		t.Errorf("Expected 'bundle files directory missing' error, got: %v", err)
	}

	// Test bundle with missing required file
	bundleWithMissingFile := &config.DeploymentBundle{
		Version: "1.0",
		Apps: map[string]*config.AppConfig{
			"testapp": {
				Name: "testapp",
				Paths: []config.Path{
					{Destination: "missing.conf", Required: true},
				},
			},
		},
	}

	err = manager.validateBundle(bundleWithMissingFile, tempDir)
	if err == nil {
		t.Error("Bundle with missing required file should fail validation")
	}

	if !strings.Contains(err.Error(), "required file missing") {
		t.Errorf("Expected 'required file missing' error, got: %v", err)
	}
}

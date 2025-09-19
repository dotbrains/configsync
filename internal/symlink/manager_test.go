package symlink

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dotbrains/configsync/internal/config"
)

func TestNewManager(t *testing.T) {
	homeDir := "/test/home"
	storeDir := "/test/store"
	backupDir := "/test/backup"
	dryRun := true
	verbose := false

	manager := NewManager(homeDir, storeDir, backupDir, dryRun, verbose)

	if manager.homeDir != homeDir {
		t.Errorf("Expected homeDir %s, got %s", homeDir, manager.homeDir)
	}

	if manager.storeDir != storeDir {
		t.Errorf("Expected storeDir %s, got %s", storeDir, manager.storeDir)
	}

	if manager.backupDir != backupDir {
		t.Errorf("Expected backupDir %s, got %s", backupDir, manager.backupDir)
	}

	if manager.dryRun != dryRun {
		t.Errorf("Expected dryRun %t, got %t", dryRun, manager.dryRun)
	}

	if manager.verbose != verbose {
		t.Errorf("Expected verbose %t, got %t", verbose, manager.verbose)
	}

	if manager.backupManager == nil {
		t.Error("Expected backup manager to be initialized")
	}
}

func TestSyncAppEnabled(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, false)

	// Create source file
	sourceFile := filepath.Join(tempDir, "test.conf")
	sourceContent := "test configuration"
	err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create app config
	appConfig := config.NewAppConfig("testapp", "Test Application")
	appConfig.AddPath(sourceFile, "test.conf", config.PathTypeFile, false)

	// Test sync
	err = manager.SyncApp(appConfig)
	if err != nil {
		t.Fatalf("SyncApp failed: %v", err)
	}

	// Verify symlink was created
	if !manager.isSymlink(sourceFile) {
		t.Error("Expected source file to be a symlink")
	}

	// Verify store file exists
	storeFile := filepath.Join(storeDir, "test.conf")
	if !manager.pathExists(storeFile) {
		t.Error("Store file should exist")
	}

	// Verify store file content
	storeContent, err := os.ReadFile(storeFile)
	if err != nil {
		t.Fatalf("Failed to read store file: %v", err)
	}

	if string(storeContent) != sourceContent {
		t.Errorf("Store content mismatch: expected %q, got %q", sourceContent, string(storeContent))
	}

	// Verify symlink target
	target, err := os.Readlink(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read symlink target: %v", err)
	}

	// Resolve to absolute path for comparison
	absTarget, err := filepath.Abs(target)
	if err != nil {
		t.Fatalf("Failed to resolve absolute target: %v", err)
	}

	absStoreFile, err := filepath.Abs(storeFile)
	if err != nil {
		t.Fatalf("Failed to resolve absolute store file: %v", err)
	}

	if absTarget != absStoreFile {
		t.Errorf("Symlink target mismatch: expected %s, got %s", absStoreFile, absTarget)
	}
}

func TestSyncAppDisabled(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, true) // verbose=true

	// Create source file
	sourceFile := filepath.Join(tempDir, "test.conf")
	sourceContent := "test configuration"
	err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create disabled app config
	appConfig := config.NewAppConfig("testapp", "Test Application")
	appConfig.Enabled = false
	appConfig.AddPath(sourceFile, "test.conf", config.PathTypeFile, false)

	// Test sync (should do nothing)
	err = manager.SyncApp(appConfig)
	if err != nil {
		t.Fatalf("SyncApp failed: %v", err)
	}

	// Verify original file still exists and is not a symlink
	if manager.isSymlink(sourceFile) {
		t.Error("Source file should not be a symlink for disabled app")
	}

	// Verify store file does not exist
	storeFile := filepath.Join(storeDir, "test.conf")
	if manager.pathExists(storeFile) {
		t.Error("Store file should not exist for disabled app")
	}
}

func TestSyncAppDryRun(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, true, true) // dryRun=true, verbose=true

	// Create source file
	sourceFile := filepath.Join(tempDir, "test.conf")
	sourceContent := "test configuration"
	err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create app config
	appConfig := config.NewAppConfig("testapp", "Test Application")
	appConfig.AddPath(sourceFile, "test.conf", config.PathTypeFile, false)

	// Test dry run sync
	err = manager.SyncApp(appConfig)
	if err != nil {
		t.Fatalf("SyncApp failed in dry run: %v", err)
	}

	// Verify no actual changes were made
	if manager.isSymlink(sourceFile) {
		t.Error("Source file should not be a symlink in dry run")
	}

	// Verify store file does not exist
	storeFile := filepath.Join(storeDir, "test.conf")
	if manager.pathExists(storeFile) {
		t.Error("Store file should not exist in dry run")
	}

	// Verify original file content unchanged
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	if string(content) != sourceContent {
		t.Errorf("Source file content should be unchanged in dry run")
	}
}

func TestSyncAppAlreadySynced(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, true) // verbose=true

	// Create store file
	storeFile := filepath.Join(storeDir, "test.conf")
	err := os.MkdirAll(filepath.Dir(storeFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create store directory: %v", err)
	}

	storeContent := "store content"
	err = os.WriteFile(storeFile, []byte(storeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create store file: %v", err)
	}

	// Create source symlink pointing to store
	sourceFile := filepath.Join(tempDir, "test.conf")
	err = os.Symlink(storeFile, sourceFile)
	if err != nil {
		t.Skipf("Skipping test due to symlink creation failure: %v", err)
	}

	// Create app config
	appConfig := config.NewAppConfig("testapp", "Test Application")
	appConfig.AddPath(sourceFile, "test.conf", config.PathTypeFile, false)

	// Test sync (should do nothing since already synced)
	err = manager.SyncApp(appConfig)
	if err != nil {
		t.Fatalf("SyncApp failed: %v", err)
	}

	// Verify symlink still exists and points to correct target
	if !manager.isSymlink(sourceFile) {
		t.Error("Source file should still be a symlink")
	}

	target, err := os.Readlink(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read symlink target: %v", err)
	}

	if target != storeFile {
		t.Errorf("Symlink target mismatch: expected %s, got %s", storeFile, target)
	}
}

func TestSyncAppRequiredMissing(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, false)

	// Create app config with required but missing path
	appConfig := config.NewAppConfig("testapp", "Test Application")
	missingFile := filepath.Join(tempDir, "missing.conf")
	appConfig.AddPath(missingFile, "missing.conf", config.PathTypeFile, true) // required=true

	// Test sync (should fail)
	err := manager.SyncApp(appConfig)
	if err == nil {
		t.Error("Expected SyncApp to fail for required missing path")
	}

	if !strings.Contains(err.Error(), "required path does not exist") {
		t.Errorf("Expected 'required path does not exist' error, got: %v", err)
	}
}

func TestSyncAppOptionalMissing(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, true) // verbose=true

	// Create app config with optional missing path
	appConfig := config.NewAppConfig("testapp", "Test Application")
	missingFile := filepath.Join(tempDir, "missing.conf")
	appConfig.AddPath(missingFile, "missing.conf", config.PathTypeFile, false) // required=false

	// Test sync (should succeed but do nothing)
	err := manager.SyncApp(appConfig)
	if err != nil {
		t.Fatalf("SyncApp should not fail for optional missing path: %v", err)
	}
}

func TestSyncAppDirectory(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, false)

	// Create source directory with files
	sourceDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Add files to source directory
	file1 := filepath.Join(sourceDir, "file1.txt")
	file2 := filepath.Join(sourceDir, "file2.txt")
	err = os.WriteFile(file1, []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	err = os.WriteFile(file2, []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create app config for directory
	appConfig := config.NewAppConfig("testapp", "Test Application")
	appConfig.AddPath(sourceDir, "config", config.PathTypeDirectory, false)

	// Test sync
	err = manager.SyncApp(appConfig)
	if err != nil {
		t.Fatalf("SyncApp failed for directory: %v", err)
	}

	// Verify symlink was created
	if !manager.isSymlink(sourceDir) {
		t.Error("Source directory should be a symlink")
	}

	// Verify store directory exists
	storeConfigDir := filepath.Join(storeDir, "config")
	if !manager.pathExists(storeConfigDir) {
		t.Error("Store directory should exist")
	}

	// Verify store directory contents
	storeFile1 := filepath.Join(storeConfigDir, "file1.txt")
	storeFile2 := filepath.Join(storeConfigDir, "file2.txt")

	content1, err := os.ReadFile(storeFile1)
	if err != nil {
		t.Errorf("Failed to read store file1: %v", err)
	}
	if string(content1) != "content1" {
		t.Errorf("Store file1 content mismatch")
	}

	content2, err := os.ReadFile(storeFile2)
	if err != nil {
		t.Errorf("Failed to read store file2: %v", err)
	}
	if string(content2) != "content2" {
		t.Errorf("Store file2 content mismatch")
	}
}

func TestUnsyncApp(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, true) // verbose=true

	// Create and sync an app first
	sourceFile := filepath.Join(tempDir, "test.conf")
	sourceContent := "test configuration"
	err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	appConfig := config.NewAppConfig("testapp", "Test Application")
	appConfig.AddPath(sourceFile, "test.conf", config.PathTypeFile, false)

	err = manager.SyncApp(appConfig)
	if err != nil {
		t.Fatalf("Failed to sync app: %v", err)
	}

	// Verify it's synced
	if !manager.isSymlink(sourceFile) {
		t.Fatal("Source file should be a symlink after sync")
	}

	// Now unsync
	err = manager.UnsyncApp(appConfig)
	if err != nil {
		t.Fatalf("UnsyncApp failed: %v", err)
	}

	// Verify file was copied back from store (should exist as regular file)
	if !manager.pathExists(sourceFile) {
		t.Error("Source file should exist after unsync (copied back from store)")
	}

	// Verify it's no longer a symlink
	if manager.isSymlink(sourceFile) {
		t.Error("Source file should not be a symlink after unsync")
	}

	// Verify store file still exists
	storeFile := filepath.Join(storeDir, "test.conf")
	if !manager.pathExists(storeFile) {
		t.Error("Store file should still exist after unsync")
	}

	// Verify content matches
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file after unsync: %v", err)
	}

	if string(content) != sourceContent {
		t.Errorf("Source file content mismatch after unsync")
	}
}

func TestUnsyncAppNotSynced(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, true) // verbose=true

	// Create regular file (not synced)
	sourceFile := filepath.Join(tempDir, "test.conf")
	sourceContent := "test configuration"
	err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	appConfig := config.NewAppConfig("testapp", "Test Application")
	appConfig.AddPath(sourceFile, "test.conf", config.PathTypeFile, false)

	// Try to unsync non-synced app (should do nothing)
	err = manager.UnsyncApp(appConfig)
	if err != nil {
		t.Fatalf("UnsyncApp should not fail for non-synced app: %v", err)
	}

	// Verify original file still exists
	if !manager.pathExists(sourceFile) {
		t.Error("Original file should still exist")
	}

	if manager.isSymlink(sourceFile) {
		t.Error("Original file should not be a symlink")
	}
}

func TestExpandPath(t *testing.T) {
	manager := NewManager("/test/home", "/store", "/backup", false, false)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Home directory path",
			input:    "~/config/test.conf",
			expected: "/test/home/config/test.conf",
		},
		{
			name:     "Absolute path",
			input:    "/absolute/path/test.conf",
			expected: "/absolute/path/test.conf",
		},
		{
			name:     "Relative path",
			input:    "relative/path/test.conf",
			expected: "relative/path/test.conf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsCorrectSymlink(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false, false)

	// Create target file
	targetFile := filepath.Join(tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create correct symlink
	symlinkFile := filepath.Join(tempDir, "symlink.txt")
	err = os.Symlink(targetFile, symlinkFile)
	if err != nil {
		t.Skipf("Skipping symlink test: %v", err)
	}

	// Test correct symlink
	if !manager.isCorrectSymlink(symlinkFile, targetFile) {
		t.Error("Should recognize correct symlink")
	}

	// Create wrong target file
	wrongTarget := filepath.Join(tempDir, "wrong.txt")
	err = os.WriteFile(wrongTarget, []byte("wrong content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create wrong target file: %v", err)
	}

	// Test incorrect symlink
	if manager.isCorrectSymlink(symlinkFile, wrongTarget) {
		t.Error("Should not recognize incorrect symlink")
	}

	// Test non-symlink file
	regularFile := filepath.Join(tempDir, "regular.txt")
	err = os.WriteFile(regularFile, []byte("regular content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}

	if manager.isCorrectSymlink(regularFile, targetFile) {
		t.Error("Should not recognize regular file as symlink")
	}

	// Test non-existent file
	if manager.isCorrectSymlink(filepath.Join(tempDir, "nonexistent.txt"), targetFile) {
		t.Error("Should not recognize non-existent file as symlink")
	}
}

func TestPathExists(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false, false)

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

func TestIsSymlink(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, "store"), filepath.Join(tempDir, "backup"), false, false)

	// Create target file
	targetFile := filepath.Join(tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create symlink
	symlinkFile := filepath.Join(tempDir, "symlink.txt")
	err = os.Symlink(targetFile, symlinkFile)
	if err != nil {
		t.Skipf("Skipping symlink test: %v", err)
	}

	// Test symlink
	if !manager.isSymlink(symlinkFile) {
		t.Error("Should recognize symlink")
	}

	// Test regular file
	if manager.isSymlink(targetFile) {
		t.Error("Should not recognize regular file as symlink")
	}

	// Test non-existent file
	if manager.isSymlink(filepath.Join(tempDir, "nonexistent.txt")) {
		t.Error("Should not recognize non-existent file as symlink")
	}
}

// Test error conditions
func TestSyncAppErrors(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, false)

	// Create app config with multiple paths, some will fail
	appConfig := config.NewAppConfig("testapp", "Test Application")
	
	// Add a valid path
	validFile := filepath.Join(tempDir, "valid.conf")
	err := os.WriteFile(validFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid file: %v", err)
	}
	appConfig.AddPath(validFile, "valid.conf", config.PathTypeFile, false)
	
	// Add an invalid required path
	invalidFile := filepath.Join(tempDir, "invalid.conf")
	appConfig.AddPath(invalidFile, "invalid.conf", config.PathTypeFile, true) // required=true

	// Test sync (should fail due to missing required file)
	err = manager.SyncApp(appConfig)
	if err == nil {
		t.Error("Expected SyncApp to fail with mixed valid/invalid paths")
	}

	// Error should mention the app display name (used in error messages)
	if !strings.Contains(err.Error(), "Test Application") {
		t.Errorf("Expected error to mention app display name 'Test Application', got: %v", err)
	}
}

// Benchmark tests
func BenchmarkSyncFile(b *testing.B) {
	tempDir := b.TempDir()
	homeDir := tempDir
	storeDir := filepath.Join(tempDir, "store")
	backupDir := filepath.Join(tempDir, "backup")

	manager := NewManager(homeDir, storeDir, backupDir, false, false)

	// Create test file
	testContent := strings.Repeat("benchmark content ", 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create unique files for each benchmark iteration
		sourceFile := filepath.Join(tempDir, "benchmark"+string(rune('0'+i%10))+".txt")
		err := os.WriteFile(sourceFile, []byte(testContent), 0644)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}

		appConfig := config.NewAppConfig("benchmarkapp", "Benchmark App")
		appConfig.AddPath(sourceFile, "benchmark"+string(rune('0'+i%10))+".txt", config.PathTypeFile, false)

		err = manager.SyncApp(appConfig)
		if err != nil {
			b.Fatalf("SyncApp failed: %v", err)
		}
	}
}
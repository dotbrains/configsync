package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotbrains/configsync/internal/config"
)

func TestNewManager(t *testing.T) {
	backupDir := "/test/backup"
	homeDir := "/test/home"
	verbose := true

	manager := NewManager(backupDir, homeDir, verbose)

	if manager.backupDir != backupDir {
		t.Errorf("Expected backupDir %s, got %s", backupDir, manager.backupDir)
	}

	if manager.homeDir != homeDir {
		t.Errorf("Expected homeDir %s, got %s", homeDir, manager.homeDir)
	}

	if manager.verbose != verbose {
		t.Errorf("Expected verbose %t, got %t", verbose, manager.verbose)
	}
}

func TestBackupPath(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create a test file to backup
	testFile := filepath.Join(tempDir, "test.conf")
	testContent := "test configuration content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config path
	configPath := &config.ConfigPath{
		Source:      testFile,
		Destination: "test.conf",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	// Test backup
	err = manager.BackupPath("testapp", configPath)
	if err != nil {
		t.Fatalf("BackupPath failed: %v", err)
	}

	// Verify backup was created
	backupPath := manager.getBackupPath("testapp", "test.conf")
	if !manager.pathExists(backupPath) {
		t.Errorf("Backup file not created: %s", backupPath)
	}

	// Verify backup content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != testContent {
		t.Errorf("Backup content mismatch: expected %q, got %q", testContent, string(backupContent))
	}

	// Verify backup info was saved
	backups, err := manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backups))
	}

	backup := backups[0]
	if backup.AppName != "testapp" {
		t.Errorf("Expected app name 'testapp', got %s", backup.AppName)
	}

	if backup.OriginalPath != testFile {
		t.Errorf("Expected original path %s, got %s", testFile, backup.OriginalPath)
	}
}

func TestBackupPathDirectory(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create a test directory with files
	testDir := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile1 := filepath.Join(testDir, "file1.txt")
	testFile2 := filepath.Join(testDir, "file2.txt")
	err = os.WriteFile(testFile1, []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}
	err = os.WriteFile(testFile2, []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}

	// Create config path for directory
	configPath := &config.ConfigPath{
		Source:      testDir,
		Destination: "testdir",
		Type:        config.PathTypeDirectory,
		Required:    false,
	}

	// Test directory backup
	err = manager.BackupPath("testapp", configPath)
	if err != nil {
		t.Fatalf("BackupPath failed for directory: %v", err)
	}

	// Verify backup directory was created
	backupPath := manager.getBackupPath("testapp", "testdir")
	if !manager.pathExists(backupPath) {
		t.Errorf("Backup directory not created: %s", backupPath)
	}

	// Verify backup directory contents
	backupFile1 := filepath.Join(backupPath, "file1.txt")
	backupFile2 := filepath.Join(backupPath, "file2.txt")

	content1, err := os.ReadFile(backupFile1)
	if err != nil {
		t.Errorf("Failed to read backup file1: %v", err)
	}
	if string(content1) != "content1" {
		t.Errorf("Backup file1 content mismatch")
	}

	content2, err := os.ReadFile(backupFile2)
	if err != nil {
		t.Errorf("Failed to read backup file2: %v", err)
	}
	if string(content2) != "content2" {
		t.Errorf("Backup file2 content mismatch")
	}
}

func TestBackupPathNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create config path for non-existent file
	configPath := &config.ConfigPath{
		Source:      filepath.Join(tempDir, "nonexistent.conf"),
		Destination: "nonexistent.conf",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	// Test backup of non-existent file (should succeed but do nothing)
	err := manager.BackupPath("testapp", configPath)
	if err != nil {
		t.Errorf("BackupPath should not fail for non-existent file: %v", err)
	}

	// Verify no backup was created
	backups, err := manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("Expected 0 backups for non-existent file, got %d", len(backups))
	}
}

func TestBackupPathSymlink(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create a test file and symlink to it
	testFile := filepath.Join(tempDir, "target.conf")
	testSymlink := filepath.Join(tempDir, "link.conf")
	
	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.Symlink(testFile, testSymlink)
	if err != nil {
		t.Skipf("Skipping symlink test: %v", err)
	}

	// Create config path for symlink
	configPath := &config.ConfigPath{
		Source:      testSymlink,
		Destination: "link.conf",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	// Test backup of symlink (should succeed but do nothing)
	err = manager.BackupPath("testapp", configPath)
	if err != nil {
		t.Errorf("BackupPath should not fail for symlink: %v", err)
	}

	// Verify no backup was created (symlinks are not backed up)
	backups, err := manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("Expected 0 backups for symlink, got %d", len(backups))
	}
}

func TestRestorePath(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create a test file and backup it
	originalFile := filepath.Join(tempDir, "original.conf")
	originalContent := "original content"
	err := os.WriteFile(originalFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create original file: %v", err)
	}

	configPath := &config.ConfigPath{
		Source:      originalFile,
		Destination: "original.conf",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	// Create backup
	err = manager.BackupPath("testapp", configPath)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify the original file
	modifiedContent := "modified content"
	err = os.WriteFile(originalFile, []byte(modifiedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to modify original file: %v", err)
	}

	// Restore from backup
	err = manager.RestorePath("testapp", configPath)
	if err != nil {
		t.Fatalf("RestorePath failed: %v", err)
	}

	// Verify restoration
	restoredContent, err := os.ReadFile(originalFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(restoredContent) != originalContent {
		t.Errorf("Restored content mismatch: expected %q, got %q", originalContent, string(restoredContent))
	}
}

func TestRestorePathNonExistentBackup(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	configPath := &config.ConfigPath{
		Source:      filepath.Join(tempDir, "nonexistent.conf"),
		Destination: "nonexistent.conf",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	// Try to restore non-existent backup
	err := manager.RestorePath("testapp", configPath)
	if err == nil {
		t.Error("Expected error when restoring non-existent backup")
	}

	if !strings.Contains(err.Error(), "backup does not exist") {
		t.Errorf("Expected 'backup does not exist' error, got: %v", err)
	}
}

func TestListBackups(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Test with no backups
	backups, err := manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("Expected 0 backups for new app, got %d", len(backups))
	}

	// Create some backups
	for i := 1; i <= 3; i++ {
		testFile := filepath.Join(tempDir, "test"+string(rune('0'+i))+".conf")
		err = os.WriteFile(testFile, []byte("content"+string(rune('0'+i))), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		configPath := &config.ConfigPath{
			Source:      testFile,
			Destination: "test" + string(rune('0'+i)) + ".conf",
			Type:        config.PathTypeFile,
			Required:    false,
		}

		err = manager.BackupPath("testapp", configPath)
		if err != nil {
			t.Fatalf("Failed to create backup %d: %v", i, err)
		}
	}

	// List backups
	backups, err = manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(backups))
	}

	// Verify backup details
	for _, backup := range backups {
		if backup.AppName != "testapp" {
			t.Errorf("Expected app name 'testapp', got %s", backup.AppName)
		}

		if backup.Size <= 0 {
			t.Errorf("Expected positive size, got %d", backup.Size)
		}

		if backup.Checksum == "" {
			t.Error("Expected non-empty checksum")
		}

		if backup.CreatedAt.IsZero() {
			t.Error("Expected non-zero creation time")
		}
	}
}

func TestCleanupBackups(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create a test file and backup it
	testFile := filepath.Join(tempDir, "test.conf")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	configPath := &config.ConfigPath{
		Source:      testFile,
		Destination: "test.conf",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	err = manager.BackupPath("testapp", configPath)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup exists
	backups, err := manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("Expected 1 backup, got %d", len(backups))
	}

	// Cleanup backups older than 0 days (should remove all)
	err = manager.CleanupBackups("testapp", 0)
	if err != nil {
		t.Fatalf("CleanupBackups failed: %v", err)
	}

	// Verify backup was removed
	backups, err = manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("Failed to list backups after cleanup: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups after cleanup, got %d", len(backups))
	}
}

func TestValidateBackup(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create a test file and backup it
	testFile := filepath.Join(tempDir, "test.conf")
	testContent := "test content for validation"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	configPath := &config.ConfigPath{
		Source:      testFile,
		Destination: "test.conf",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	err = manager.BackupPath("testapp", configPath)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Get backup info
	backups, err := manager.ListBackups("testapp")
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("Expected 1 backup, got %d", len(backups))
	}

	backup := backups[0]

	// Validate backup (should succeed)
	err = manager.ValidateBackup(backup)
	if err != nil {
		t.Errorf("ValidateBackup failed: %v", err)
	}

	// Corrupt the backup file (keep same size to trigger checksum validation)
	// Original: "test content for validation" (27 chars)
	corruptedContent := "corrupt content validations" // 27 chars to match original
	if len(corruptedContent) != len(testContent) {
		t.Fatalf("Corrupted content must have same length as original for this test (original: %d, corrupted: %d)", len(testContent), len(corruptedContent))
	}
	err = os.WriteFile(backup.BackupPath, []byte(corruptedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to corrupt backup file: %v", err)
	}

	// Validate backup (should fail due to checksum mismatch)
	err = manager.ValidateBackup(backup)
	if err == nil {
		t.Error("Expected validation to fail for corrupted backup")
	}

	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("Expected checksum mismatch error, got: %v", err)
	}
}

func TestValidateBackupMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")
	homeDir := tempDir

	manager := NewManager(backupDir, homeDir, false)

	// Create fake backup info for non-existent file
	backupInfo := &config.BackupInfo{
		AppName:      "testapp",
		OriginalPath: filepath.Join(tempDir, "original.conf"),
		BackupPath:   filepath.Join(backupDir, "missing.conf"),
		CreatedAt:    time.Now(),
		Size:         100,
		Checksum:     "fakechecksum",
	}

	// Validate missing backup (should fail)
	err := manager.ValidateBackup(backupInfo)
	if err == nil {
		t.Error("Expected validation to fail for missing backup file")
	}

	if !strings.Contains(err.Error(), "backup file missing") {
		t.Errorf("Expected 'backup file missing' error, got: %v", err)
	}
}

func TestExpandPath(t *testing.T) {
	manager := NewManager("/backup", "/test/home", false)

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

func TestCalculateChecksum(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(filepath.Join(tempDir, "backup"), tempDir, false)

	// Create test file with known content
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate checksum
	checksum1, err := manager.calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}

	// Checksum should be consistent
	checksum2, err := manager.calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate checksum again: %v", err)
	}

	if checksum1 != checksum2 {
		t.Errorf("Checksums should be identical: %s != %s", checksum1, checksum2)
	}

	// Checksum should be non-empty and hexadecimal
	if checksum1 == "" {
		t.Error("Checksum should not be empty")
	}

	if len(checksum1) != 64 { // SHA256 produces 64 character hex string
		t.Errorf("Expected 64 character checksum, got %d", len(checksum1))
	}
}

func TestCalculateSize(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(filepath.Join(tempDir, "backup"), tempDir, false)

	// Test file size
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size, err := manager.calculateSize(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate file size: %v", err)
	}

	expectedSize := int64(len(testContent))
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}

	// Test directory size
	testDir := filepath.Join(tempDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Add files to directory
	for i := 1; i <= 3; i++ {
		fileName := filepath.Join(testDir, "file"+string(rune('0'+i))+".txt")
		content := strings.Repeat("x", i*10)
		err = os.WriteFile(fileName, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %d: %v", i, err)
		}
	}

	dirSize, err := manager.calculateSize(testDir)
	if err != nil {
		t.Fatalf("Failed to calculate directory size: %v", err)
	}

	expectedDirSize := int64(10 + 20 + 30) // 10 + 20 + 30 bytes
	if dirSize != expectedDirSize {
		t.Errorf("Expected directory size %d, got %d", expectedDirSize, dirSize)
	}
}

// Benchmark tests
func BenchmarkBackupFile(b *testing.B) {
	tempDir := b.TempDir()
	backupDir := filepath.Join(tempDir, "backup")
	manager := NewManager(backupDir, tempDir, false)

	// Create test file
	testFile := filepath.Join(tempDir, "benchmark.txt")
	testContent := strings.Repeat("benchmark content ", 1000)
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	configPath := &config.ConfigPath{
		Source:      testFile,
		Destination: "benchmark.txt",
		Type:        config.PathTypeFile,
		Required:    false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different app names to avoid conflicts
		appName := "benchmarkapp" + string(rune('0'+i%10))
		err = manager.BackupPath(appName, configPath)
		if err != nil {
			b.Fatalf("BackupPath failed: %v", err)
		}
	}
}

func BenchmarkCalculateChecksum(b *testing.B) {
	tempDir := b.TempDir()
	manager := NewManager(filepath.Join(tempDir, "backup"), tempDir, false)

	// Create test file
	testFile := filepath.Join(tempDir, "benchmark.txt")
	testContent := strings.Repeat("benchmark content for checksum calculation ", 1000)
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.calculateChecksum(testFile)
		if err != nil {
			b.Fatalf("calculateChecksum failed: %v", err)
		}
	}
}
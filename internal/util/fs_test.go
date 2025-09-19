package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathExists(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "testfile.txt")

	// Create the test file
	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Errorf("Failed to close file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing file",
			path:     tempFile,
			expected: true,
		},
		{
			name:     "Existing directory",
			path:     tempDir,
			expected: true,
		},
		{
			name:     "Non-existent file",
			path:     filepath.Join(tempDir, "nonexistent.txt"),
			expected: false,
		},
		{
			name:     "Non-existent directory",
			path:     filepath.Join(tempDir, "nonexistent"),
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "Root directory",
			path:     "/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathExists(tt.path)
			if result != tt.expected {
				t.Errorf("PathExists(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestPathExistsSymlink(t *testing.T) {
	tempDir := t.TempDir()

	// Create a target file
	targetFile := filepath.Join(tempDir, "target.txt")
	file, err := os.Create(targetFile)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Errorf("Failed to close file: %v", err)
	}

	// Create a symlink to the target file
	symlinkFile := filepath.Join(tempDir, "symlink.txt")
	err = os.Symlink(targetFile, symlinkFile)
	if err != nil {
		t.Skipf("Skipping symlink test: %v", err) // Skip if symlinks not supported
	}

	// Test that PathExists works with symlinks
	if !PathExists(symlinkFile) {
		t.Error("Expected PathExists to return true for valid symlink")
	}

	// Create a broken symlink
	brokenSymlink := filepath.Join(tempDir, "broken.txt")
	err = os.Symlink(filepath.Join(tempDir, "nonexistent.txt"), brokenSymlink)
	if err != nil {
		t.Skipf("Skipping broken symlink test: %v", err)
	}

	// Test that PathExists returns false for broken symlinks
	if PathExists(brokenSymlink) {
		t.Error("Expected PathExists to return false for broken symlink")
	}
}

// Benchmark PathExists performance
func BenchmarkPathExists(b *testing.B) {
	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "benchmark.txt")

	file, err := os.Create(tempFile)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}
	if err := file.Close(); err != nil {
		b.Fatalf("Failed to close benchmark file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PathExists(tempFile)
	}
}

func BenchmarkPathExistsNonExistent(b *testing.B) {
	tempDir := b.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PathExists(nonExistentFile)
	}
}

// Package backup provides functionality for creating, managing, and validating configuration backups.
package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/dotbrains/configsync/internal/config"
)

// Manager handles backup operations for configurations
type Manager struct {
	backupDir string
	homeDir   string
	verbose   bool
}

// NewManager creates a new backup manager
func NewManager(backupDir, homeDir string, verbose bool) *Manager {
	return &Manager{
		backupDir: backupDir,
		homeDir:   homeDir,
		verbose:   verbose,
	}
}

// BackupPath creates a backup of a configuration path before symlinking
func (m *Manager) BackupPath(appName string, path *config.ConfigPath) error {
	sourcePath := m.expandPath(path.Source)

	// Check if source exists
	if !m.pathExists(sourcePath) {
		if m.verbose {
			fmt.Printf("    No backup needed - path does not exist: %s\n", sourcePath)
		}
		return nil
	}

	// Check if it's already a symlink (don't backup symlinks)
	if m.isSymlink(sourcePath) {
		if m.verbose {
			fmt.Printf("    No backup needed - path is already a symlink: %s\n", sourcePath)
		}
		return nil
	}

	// Create backup info
	backupInfo := &config.BackupInfo{
		AppName:      appName,
		OriginalPath: sourcePath,
		CreatedAt:    time.Now(),
	}

	// Calculate checksum
	if path.Type == config.PathTypeFile {
		checksum, err := m.calculateChecksum(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}
		backupInfo.Checksum = checksum
	}

	// Get file/directory size
	size, err := m.calculateSize(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to calculate size: %w", err)
	}
	backupInfo.Size = size

	// Create backup path
	backupPath := m.getBackupPath(appName, path.Destination)
	backupInfo.BackupPath = backupPath

	if m.verbose {
		fmt.Printf("    Creating backup: %s -> %s\n", sourcePath, backupPath)
	}

	// Create backup directory
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy file/directory to backup location
	if err := m.copyPath(sourcePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Save backup metadata
	if err := m.saveBackupInfo(backupInfo); err != nil {
		return fmt.Errorf("failed to save backup info: %w", err)
	}

	if m.verbose {
		fmt.Printf("    Backup created successfully (%d bytes)\n", backupInfo.Size)
	}

	return nil
}

// RestorePath restores a configuration path from backup
func (m *Manager) RestorePath(appName string, path *config.ConfigPath) error {
	sourcePath := m.expandPath(path.Source)
	backupPath := m.getBackupPath(appName, path.Destination)

	if m.verbose {
		fmt.Printf("    Restoring: %s <- %s\n", sourcePath, backupPath)
	}

	// Check if backup exists
	if !m.pathExists(backupPath) {
		return fmt.Errorf("backup does not exist: %s", backupPath)
	}

	// Remove existing file/symlink if it exists
	if m.pathExists(sourcePath) {
		if m.verbose {
			fmt.Printf("    Removing existing: %s\n", sourcePath)
		}
		if err := os.RemoveAll(sourcePath); err != nil {
			return fmt.Errorf("failed to remove existing path: %w", err)
		}
	}

	// Create source directory if needed
	sourceDir := filepath.Dir(sourcePath)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Copy backup back to original location
	if err := m.copyPath(backupPath, sourcePath); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	if m.verbose {
		fmt.Printf("    Restored successfully\n")
	}

	return nil
}

// ListBackups returns information about all backups for an application
func (m *Manager) ListBackups(appName string) ([]*config.BackupInfo, error) {
	backupInfoDir := filepath.Join(m.backupDir, "info", appName)

	if !m.pathExists(backupInfoDir) {
		return []*config.BackupInfo{}, nil
	}

	entries, err := os.ReadDir(backupInfoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup info directory: %w", err)
	}

	var backups []*config.BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		infoPath := filepath.Join(backupInfoDir, entry.Name())
		backupInfo, err := m.loadBackupInfo(infoPath)
		if err != nil {
			if m.verbose {
				fmt.Printf("Warning: failed to load backup info %s: %v\n", infoPath, err)
			}
			continue
		}

		backups = append(backups, backupInfo)
	}

	return backups, nil
}

// CleanupBackups removes old backups for an application
func (m *Manager) CleanupBackups(appName string, keepDays int) error {
	backups, err := m.ListBackups(appName)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -keepDays)
	var removed int

	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoff) {
			if m.verbose {
				fmt.Printf("Removing old backup: %s (created %s)\n",
					backup.BackupPath, backup.CreatedAt.Format(time.RFC3339))
			}

			// Remove backup file/directory
			if err := os.RemoveAll(backup.BackupPath); err != nil {
				if m.verbose {
					fmt.Printf("Warning: failed to remove backup file %s: %v\n", backup.BackupPath, err)
				}
			}

			// Remove backup info file
			infoPath := m.getBackupInfoPath(appName, backup.OriginalPath)
			if err := os.Remove(infoPath); err != nil {
				if m.verbose {
					fmt.Printf("Warning: failed to remove backup info %s: %v\n", infoPath, err)
				}
			}

			removed++
		}
	}

	if m.verbose && removed > 0 {
		fmt.Printf("Cleaned up %d old backup(s) for %s\n", removed, appName)
	}

	return nil
}

// ValidateBackup verifies the integrity of a backup
func (m *Manager) ValidateBackup(backupInfo *config.BackupInfo) error {
	if !m.pathExists(backupInfo.BackupPath) {
		return fmt.Errorf("backup file missing: %s", backupInfo.BackupPath)
	}

	// Verify size
	currentSize, err := m.calculateSize(backupInfo.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to calculate current size: %w", err)
	}

	if currentSize != backupInfo.Size {
		return fmt.Errorf("backup size mismatch: expected %d, got %d", backupInfo.Size, currentSize)
	}

	// Verify checksum for files
	if backupInfo.Checksum != "" {
		currentChecksum, err := m.calculateChecksum(backupInfo.BackupPath)
		if err != nil {
			return fmt.Errorf("failed to calculate current checksum: %w", err)
		}

		if currentChecksum != backupInfo.Checksum {
			return fmt.Errorf("backup checksum mismatch: expected %s, got %s",
				backupInfo.Checksum, currentChecksum)
		}
	}

	return nil
}

// Helper methods

func (m *Manager) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(m.homeDir, path[2:])
	}
	return path
}

func (m *Manager) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (m *Manager) isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

func (m *Manager) getBackupPath(appName, destination string) string {
	// Create a safe filename from the destination path
	safeName := strings.ReplaceAll(destination, "/", "_")
	safeName = strings.ReplaceAll(safeName, " ", "_")

	return filepath.Join(m.backupDir, "files", appName, safeName)
}

func (m *Manager) getBackupInfoPath(appName, originalPath string) string {
	// Create a safe filename from the original path
	safeName := strings.ReplaceAll(originalPath, "/", "_")
	safeName = strings.ReplaceAll(safeName, " ", "_")

	return filepath.Join(m.backupDir, "info", appName, safeName+".yaml")
}

func (m *Manager) copyPath(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return m.copyDir(src, dst)
	}
	return m.copyFile(src, dst)
}

func (m *Manager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func (m *Manager) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return m.copyFile(path, dstPath)
	})
}

func (m *Manager) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (m *Manager) calculateSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if !info.IsDir() {
		return info.Size(), nil
	}

	var size int64
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

func (m *Manager) saveBackupInfo(info *config.BackupInfo) error {
	infoPath := m.getBackupInfoPath(info.AppName, info.OriginalPath)

	// Create info directory
	infoDir := filepath.Dir(infoPath)
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(info)
	if err != nil {
		return err
	}

	return os.WriteFile(infoPath, data, 0644)
}

func (m *Manager) loadBackupInfo(infoPath string) (*config.BackupInfo, error) {
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	var info config.BackupInfo
	if err := yaml.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

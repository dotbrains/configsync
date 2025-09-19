// Package symlink provides functionality for managing symlinks between configuration files and central storage.
package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotbrains/configsync/internal/backup"
	"github.com/dotbrains/configsync/internal/config"
)

// Manager handles symlink operations
type Manager struct {
	backupManager *backup.Manager
	homeDir       string
	storeDir      string
	backupDir     string
	dryRun        bool
	verbose       bool
}

// NewManager creates a new symlink manager
func NewManager(homeDir, storeDir, backupDir string, dryRun, verbose bool) *Manager {
	return &Manager{
		homeDir:       homeDir,
		storeDir:      storeDir,
		backupDir:     backupDir,
		dryRun:        dryRun,
		verbose:       verbose,
		backupManager: backup.NewManager(backupDir, homeDir, verbose),
	}
}

// SyncApp creates symlinks for all paths in an application configuration
func (m *Manager) SyncApp(appConfig *config.AppConfig) error {
	if !appConfig.IsEnabled() {
		if m.verbose {
			fmt.Printf("Skipping disabled app: %s\n", appConfig.DisplayName)
		}
		return nil
	}

	if m.verbose {
		fmt.Printf("Syncing %s...\n", appConfig.DisplayName)
	}

	var errors []string
	for i := range appConfig.Paths {
		path := &appConfig.Paths[i]

		if err := m.syncPath(path); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", path.Source, err))
			continue
		}

		if !m.dryRun {
			path.MarkSynced()
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors syncing %s:\n%s", appConfig.DisplayName, strings.Join(errors, "\n"))
	}

	return nil
}

// UnsyncApp removes symlinks for all paths in an application configuration
func (m *Manager) UnsyncApp(appConfig *config.AppConfig) error {
	if m.verbose {
		fmt.Printf("Unsyncing %s...\n", appConfig.DisplayName)
	}

	var errors []string
	for i := range appConfig.Paths {
		path := &appConfig.Paths[i]

		if err := m.unsyncPath(path); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", path.Source, err))
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors unsyncing %s:\n%s", appConfig.DisplayName, strings.Join(errors, "\n"))
	}

	return nil
}

// syncPath creates a symlink for a single configuration path
func (m *Manager) syncPath(path *config.Path) error {
	sourcePath := m.expandPath(path.Source)
	storePath := filepath.Join(m.storeDir, path.Destination)

	if m.verbose {
		fmt.Printf("  Syncing: %s -> %s\n", sourcePath, storePath)
	}

	if m.isCorrectSymlink(sourcePath, storePath) {
		if m.verbose {
			fmt.Printf("    Already synced correctly\n")
		}
		return nil
	}

	if err := m.ensureStoreDirectory(storePath); err != nil {
		return err
	}

	if err := m.handleExistingSource(sourcePath, storePath, path); err != nil {
		return err
	}

	if !m.pathExists(sourcePath) && !m.pathExists(storePath) {
		return m.handleMissingPath(sourcePath, path)
	}

	return m.createFinalSymlink(sourcePath, storePath)
}

// unsyncPath removes a symlink and restores the original file if backed up
func (m *Manager) unsyncPath(path *config.Path) error {
	sourcePath := m.expandPath(path.Source)
	storePath := filepath.Join(m.storeDir, path.Destination)

	if m.verbose {
		fmt.Printf("  Unsyncing: %s\n", sourcePath)
	}

	// Check if source is a symlink to the store
	if !m.isCorrectSymlink(sourcePath, storePath) {
		if m.verbose {
			fmt.Printf("    Not a valid symlink, skipping\n")
		}
		return nil
	}

	// Remove the symlink
	if m.verbose {
		fmt.Printf("    Removing symlink: %s\n", sourcePath)
	}
	if !m.dryRun {
		if err := os.Remove(sourcePath); err != nil {
			return fmt.Errorf("failed to remove symlink: %w", err)
		}
	} else {
		fmt.Printf("    [DRY RUN] Would remove symlink: %s\n", sourcePath)
	}

	// Copy back from store if it exists
	if m.pathExists(storePath) {
		if m.verbose {
			fmt.Printf("    Copying back from store: %s -> %s\n", storePath, sourcePath)
		}
		if !m.dryRun {
			if err := m.copyFromStore(storePath, sourcePath); err != nil {
				return fmt.Errorf("failed to copy from store: %w", err)
			}
		} else {
			fmt.Printf("    [DRY RUN] Would copy: %s -> %s\n", storePath, sourcePath)
		}
	}

	return nil
}

// ensureStoreDirectory creates the store directory if needed
func (m *Manager) ensureStoreDirectory(storePath string) error {
	storeDir := filepath.Dir(storePath)
	if !m.dryRun {
		if err := os.MkdirAll(storeDir, 0755); err != nil {
			return fmt.Errorf("failed to create store directory: %w", err)
		}
	} else {
		fmt.Printf("    [DRY RUN] Would create directory: %s\n", storeDir)
	}
	return nil
}

// handleExistingSource processes an existing source file or symlink
func (m *Manager) handleExistingSource(sourcePath, storePath string, path *config.Path) error {
	if !m.pathExists(sourcePath) {
		return nil
	}

	if m.isSymlink(sourcePath) {
		return m.removeExistingSymlink(sourcePath)
	}

	return m.moveSourceToStore(sourcePath, storePath, path)
}

// removeExistingSymlink removes an existing symlink
func (m *Manager) removeExistingSymlink(sourcePath string) error {
	if m.verbose {
		fmt.Printf("    Removing existing symlink: %s\n", sourcePath)
	}
	if !m.dryRun {
		if err := os.Remove(sourcePath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	} else {
		fmt.Printf("    [DRY RUN] Would remove symlink: %s\n", sourcePath)
	}
	return nil
}

// moveSourceToStore moves the source file/directory to store with backup
func (m *Manager) moveSourceToStore(sourcePath, storePath string, path *config.Path) error {
	if !m.dryRun {
		if err := m.backupManager.BackupPath("temp", path); err != nil {
			if m.verbose {
				fmt.Printf("    Warning: backup failed: %v\n", err)
			}
		}
	}

	if m.verbose {
		fmt.Printf("    Moving to store: %s -> %s\n", sourcePath, storePath)
	}
	if !m.dryRun {
		if err := m.moveToStore(sourcePath, storePath); err != nil {
			return fmt.Errorf("failed to move to store: %w", err)
		}
		path.MarkBackedUp()
	} else {
		fmt.Printf("    [DRY RUN] Would move: %s -> %s\n", sourcePath, storePath)
	}
	return nil
}

// handleMissingPath handles the case where neither source nor store exists
func (m *Manager) handleMissingPath(sourcePath string, path *config.Path) error {
	if path.Required {
		return fmt.Errorf("required path does not exist: %s", sourcePath)
	}
	if m.verbose {
		fmt.Printf("    Skipping non-existent optional path: %s\n", sourcePath)
	}
	return nil
}

// createFinalSymlink creates the final symlink
func (m *Manager) createFinalSymlink(sourcePath, storePath string) error {
	if m.verbose {
		fmt.Printf("    Creating symlink: %s -> %s\n", sourcePath, storePath)
	}
	if !m.dryRun {
		if err := m.createSymlink(storePath, sourcePath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	} else {
		fmt.Printf("    [DRY RUN] Would create symlink: %s -> %s\n", sourcePath, storePath)
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

func (m *Manager) isCorrectSymlink(sourcePath, targetPath string) bool {
	if !m.isSymlink(sourcePath) {
		return false
	}

	link, err := os.Readlink(sourcePath)
	if err != nil {
		return false
	}

	// Resolve relative paths
	if !filepath.IsAbs(link) {
		baseDir := filepath.Dir(sourcePath)
		link = filepath.Join(baseDir, link)
	}

	// Clean both paths for comparison
	link = filepath.Clean(link)
	targetPath = filepath.Clean(targetPath)

	return link == targetPath
}

func (m *Manager) createSymlink(target, source string) error {
	// Ensure the source directory exists
	sourceDir := filepath.Dir(source)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Create the symlink
	return os.Symlink(target, source)
}

func (m *Manager) moveToStore(sourcePath, storePath string) error {
	// Ensure store directory exists
	storeDir := filepath.Dir(storePath)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Move the file/directory
	return os.Rename(sourcePath, storePath)
}

func (m *Manager) copyFromStore(storePath, sourcePath string) error {
	// Ensure source directory exists
	sourceDir := filepath.Dir(sourcePath)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Check if store path is a directory
	info, err := os.Stat(storePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return m.copyDir(storePath, sourcePath)
	}
	return m.copyFile(storePath, sourcePath)
}

func (m *Manager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	// Copy file contents
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	// Copy permissions
	info, statErr := sourceFile.Stat()
	if statErr != nil {
		return statErr
	}
	return os.Chmod(dst, info.Mode())
}

func (m *Manager) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		return m.copyFile(path, destPath)
	})
}

package deploy

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/dotbrains/configsync/internal/config"
)

// Manager handles deployment operations for configuration bundles
type Manager struct {
	homeDir   string
	storeDir  string
	backupDir string
	verbose   bool
}

// NewManager creates a new deployment manager
func NewManager(homeDir, storeDir, backupDir string, verbose bool) *Manager {
	return &Manager{
		homeDir:   homeDir,
		storeDir:  storeDir,
		backupDir: backupDir,
		verbose:   verbose,
	}
}

// ExportBundle creates a deployment bundle from current configuration
func (m *Manager) ExportBundle(bundlePath string, apps []string, configManager *config.Manager) error {
	if m.verbose {
		fmt.Printf("Creating deployment bundle: %s\n", bundlePath)
	}

	// Load current configuration
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create deployment bundle
	bundle := &config.DeploymentBundle{
		Version:   "1.0",
		CreatedAt: time.Now(),
		CreatedBy: m.getUserInfo(),
		Apps:      make(map[string]*config.AppConfig),
		Metadata:  make(map[string]string),
	}

	// Add system information to metadata
	bundle.Metadata["platform"] = "darwin"
	bundle.Metadata["created_on"] = m.getSystemInfo()

	// Select apps to include
	if len(apps) == 0 {
		// Include all apps
		bundle.Apps = cfg.Apps
		if m.verbose {
			fmt.Printf("Including all %d configured applications\n", len(cfg.Apps))
		}
	} else {
		// Include specified apps
		for _, appName := range apps {
			if appConfig, exists := cfg.Apps[appName]; exists {
				bundle.Apps[appName] = appConfig
				if m.verbose {
					fmt.Printf("Including application: %s\n", appConfig.DisplayName)
				}
			} else {
				return fmt.Errorf("application not found: %s", appName)
			}
		}
	}

	if len(bundle.Apps) == 0 {
		return fmt.Errorf("no applications to export")
	}

	// Create temporary directory for bundle contents
	tempDir, err := os.MkdirTemp("", "configsync-bundle-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Save bundle metadata
	bundleFile := filepath.Join(tempDir, "bundle.yaml")
	if err := m.saveBundleMetadata(bundle, bundleFile); err != nil {
		return fmt.Errorf("failed to save bundle metadata: %w", err)
	}

	// Copy configuration files
	filesDir := filepath.Join(tempDir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return fmt.Errorf("failed to create files directory: %w", err)
	}

	for _, appConfig := range bundle.Apps {
		appFilesDir := filepath.Join(filesDir, appConfig.Name)
		if err := os.MkdirAll(appFilesDir, 0755); err != nil {
			return fmt.Errorf("failed to create app files directory: %w", err)
		}

		for _, path := range appConfig.Paths {
			storePath := filepath.Join(m.storeDir, path.Destination)
			if !m.pathExists(storePath) {
				if m.verbose {
					fmt.Printf("  Skipping missing file: %s\n", storePath)
				}
				continue
			}

			// Create destination path in bundle
			destPath := filepath.Join(appFilesDir, path.Destination)
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("failed to create bundle path directory: %w", err)
			}

			// Copy file/directory
			if err := m.copyPath(storePath, destPath); err != nil {
				return fmt.Errorf("failed to copy %s: %w", storePath, err)
			}

			if m.verbose {
				fmt.Printf("  Added: %s\n", path.Destination)
			}
		}
	}

	// Create compressed bundle
	if err := m.createTarGz(tempDir, bundlePath); err != nil {
		return fmt.Errorf("failed to create bundle archive: %w", err)
	}

	if m.verbose {
		bundleSize, _ := m.getFileSize(bundlePath)
		fmt.Printf("Bundle created successfully: %s (%d bytes)\n", bundlePath, bundleSize)
	}

	return nil
}

// ImportBundle imports a deployment bundle and validates its contents
func (m *Manager) ImportBundle(bundlePath, targetDir string) (*config.DeploymentBundle, error) {
	if m.verbose {
		fmt.Printf("Importing deployment bundle: %s\n", bundlePath)
	}

	// Check if bundle exists
	if !m.pathExists(bundlePath) {
		return nil, fmt.Errorf("bundle file not found: %s", bundlePath)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract bundle
	if err := m.extractTarGz(bundlePath, targetDir); err != nil {
		return nil, fmt.Errorf("failed to extract bundle: %w", err)
	}

	// Load bundle metadata
	bundleFile := filepath.Join(targetDir, "bundle.yaml")
	bundle, err := m.loadBundleMetadata(bundleFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load bundle metadata: %w", err)
	}

	// Validate bundle contents
	if err := m.validateBundle(bundle, targetDir); err != nil {
		return nil, fmt.Errorf("bundle validation failed: %w", err)
	}

	if m.verbose {
		fmt.Printf("Bundle imported successfully: %d applications\n", len(bundle.Apps))
		for _, appConfig := range bundle.Apps {
			fmt.Printf("  %s (%d paths)\n", appConfig.DisplayName, len(appConfig.Paths))
		}
	}

	return bundle, nil
}

// DeployBundle deploys an imported bundle to the current system
func (m *Manager) DeployBundle(bundle *config.DeploymentBundle, bundleDir string, configManager *config.Manager, force bool) error {
	if m.verbose {
		fmt.Printf("Deploying bundle to current system\n")
	}

	// Load current configuration
	currentCfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load current configuration: %w", err)
	}

	// Check for conflicts if not forcing
	if !force {
		conflicts := m.detectConflicts(bundle, currentCfg)
		if len(conflicts) > 0 {
			fmt.Println("Deployment conflicts detected:")
			for _, conflict := range conflicts {
				fmt.Printf("  - %s: %s\n", conflict.AppName, conflict.Message)
			}
			return fmt.Errorf("use --force to override conflicts")
		}
	}

	// Deploy each application
	var deployed []string
	var failed []string

	for appName, bundleAppConfig := range bundle.Apps {
		if m.verbose {
			fmt.Printf("\nDeploying %s...\n", bundleAppConfig.DisplayName)
		}

		// Copy files from bundle to store
		bundleFilesDir := filepath.Join(bundleDir, "files", appName)
		if m.pathExists(bundleFilesDir) {
			if err := m.deployAppFiles(bundleAppConfig, bundleFilesDir); err != nil {
				if m.verbose {
					fmt.Printf("  ✗ Failed to deploy files for %s: %v\n", bundleAppConfig.DisplayName, err)
				}
				failed = append(failed, bundleAppConfig.DisplayName)
				continue
			}
		}

		// Add/update app configuration
		if err := configManager.AddApp(bundleAppConfig); err != nil {
			if m.verbose {
				fmt.Printf("  ✗ Failed to add configuration for %s: %v\n", bundleAppConfig.DisplayName, err)
			}
			failed = append(failed, bundleAppConfig.DisplayName)
			continue
		}

		if m.verbose {
			fmt.Printf("  ✓ Deployed %s successfully\n", bundleAppConfig.DisplayName)
		}
		deployed = append(deployed, bundleAppConfig.DisplayName)
	}

	// Show deployment summary
	fmt.Println()
	if len(deployed) > 0 {
		fmt.Printf("✓ Successfully deployed %d application(s):\n", len(deployed))
		for _, name := range deployed {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(failed) > 0 {
		fmt.Printf("\n✗ Failed to deploy %d application(s):\n", len(failed))
		for _, name := range failed {
			fmt.Printf("  - %s\n", name)
		}
		
		if len(deployed) == 0 {
			return fmt.Errorf("failed to deploy any applications")
		}
	}

	if len(deployed) > 0 {
		fmt.Println("\nNext step: Run 'configsync sync' to create symlinks")
	}

	return nil
}

// Helper methods and types

type Conflict struct {
	AppName string
	Message string
}

func (m *Manager) detectConflicts(bundle *config.DeploymentBundle, currentCfg *config.Config) []Conflict {
	var conflicts []Conflict

	for appName, bundleApp := range bundle.Apps {
		if currentApp, exists := currentCfg.Apps[appName]; exists {
			// Check if versions differ significantly
			if currentApp.LastSynced.After(bundle.CreatedAt) {
				conflicts = append(conflicts, Conflict{
					AppName: appName,
					Message: fmt.Sprintf("local configuration is newer than bundle (local: %s, bundle: %s)",
						currentApp.LastSynced.Format("2006-01-02"), bundle.CreatedAt.Format("2006-01-02")),
				})
			}

			// Check if paths have changed
			if len(currentApp.Paths) != len(bundleApp.Paths) {
				conflicts = append(conflicts, Conflict{
					AppName: appName,
					Message: fmt.Sprintf("path count differs (local: %d, bundle: %d)",
						len(currentApp.Paths), len(bundleApp.Paths)),
				})
			}
		}
	}

	return conflicts
}

func (m *Manager) deployAppFiles(appConfig *config.AppConfig, bundleFilesDir string) error {
	for _, path := range appConfig.Paths {
		bundlePath := filepath.Join(bundleFilesDir, path.Destination)
		if !m.pathExists(bundlePath) {
			if path.Required {
				return fmt.Errorf("required file missing from bundle: %s", path.Destination)
			}
			continue
		}

		// Copy to store
		storePath := filepath.Join(m.storeDir, path.Destination)
		storeDir := filepath.Dir(storePath)
		if err := os.MkdirAll(storeDir, 0755); err != nil {
			return fmt.Errorf("failed to create store directory: %w", err)
		}

		if err := m.copyPath(bundlePath, storePath); err != nil {
			return fmt.Errorf("failed to copy to store: %w", err)
		}

		if m.verbose {
			fmt.Printf("    Copied: %s\n", path.Destination)
		}
	}

	return nil
}

func (m *Manager) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (m *Manager) getUserInfo() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	return "unknown"
}

func (m *Manager) getSystemInfo() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s", hostname)
}

func (m *Manager) getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (m *Manager) copyPath(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return m.copyDir(src, dst)
	} else {
		return m.copyFile(src, dst)
	}
}

func (m *Manager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

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
		} else {
			return m.copyFile(path, dstPath)
		}
	})
}

func (m *Manager) saveBundleMetadata(bundle *config.DeploymentBundle, path string) error {
	data, err := yaml.Marshal(bundle)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *Manager) loadBundleMetadata(path string) (*config.DeploymentBundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bundle config.DeploymentBundle
	if err := yaml.Unmarshal(data, &bundle); err != nil {
		return nil, err
	}

	return &bundle, nil
}

func (m *Manager) validateBundle(bundle *config.DeploymentBundle, bundleDir string) error {
	// Check bundle version
	if bundle.Version == "" {
		return fmt.Errorf("missing bundle version")
	}

	// Check that files directory exists
	filesDir := filepath.Join(bundleDir, "files")
	if !m.pathExists(filesDir) {
		return fmt.Errorf("bundle files directory missing")
	}

	// Validate each app
	for appName, appConfig := range bundle.Apps {
		appFilesDir := filepath.Join(filesDir, appName)
		
		for _, path := range appConfig.Paths {
			if path.Required {
				bundlePath := filepath.Join(appFilesDir, path.Destination)
				if !m.pathExists(bundlePath) {
					return fmt.Errorf("required file missing for %s: %s", appName, path.Destination)
				}
			}
		}
	}

	return nil
}

func (m *Manager) createTarGz(sourceDir, targetPath string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			return err
		}

		return nil
	})
}

func (m *Manager) extractTarGz(sourcePath, targetDir string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(targetDir, header.Name)

		// Security check: ensure path is within target directory
		if !strings.HasPrefix(path, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			_, err = io.Copy(file, tarReader)
			file.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
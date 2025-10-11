package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	yaml "gopkg.in/yaml.v3"
)

const (
	// DefaultConfigDir is the default directory name for ConfigSync configuration
	DefaultConfigDir = ".configsync"
	// DefaultConfigFile is the default filename for ConfigSync configuration
	DefaultConfigFile = "config.yaml"
	// DefaultStoreDir is the default directory name for ConfigSync store
	DefaultStoreDir = "store"
	// DefaultBackupDir is the default directory name for ConfigSync backups
	DefaultBackupDir = "backups"
	// DefaultLogDir is the default directory name for ConfigSync logs
	DefaultLogDir = "logs"
)

// Manager handles configuration file operations
type Manager struct {
	config     *Config
	configDir  string
	configPath string
}

// NewManager creates a new configuration manager
func NewManager(homeDir string) *Manager {
	configDir := filepath.Join(homeDir, DefaultConfigDir)
	configPath := filepath.Join(configDir, DefaultConfigFile)

	return &Manager{
		configDir:  configDir,
		configPath: configPath,
	}
}

// Initialize creates the configuration directory structure and initial config file
func (m *Manager) Initialize() error {
	// Create main config directory
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create subdirectories
	storeDir := filepath.Join(m.configDir, DefaultStoreDir)
	backupDir := filepath.Join(m.configDir, DefaultBackupDir)
	logDir := filepath.Join(m.configDir, DefaultLogDir)

	for _, dir := range []string{storeDir, backupDir, logDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create store subdirectories that mirror macOS structure
	libraryDir := filepath.Join(storeDir, "Library")
	prefsDir := filepath.Join(libraryDir, "Preferences")
	appSupportDir := filepath.Join(libraryDir, "Application Support")
	configHomeDir := filepath.Join(storeDir, ".config")

	for _, dir := range []string{libraryDir, prefsDir, appSupportDir, configHomeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create store directory %s: %w", dir, err)
		}
	}

	// Create initial config file if it doesn't exist
	if !m.configExists() {
		config := NewDefaultConfig(storeDir, backupDir, logDir)
		if err := m.saveConfig(config); err != nil {
			return fmt.Errorf("failed to create initial config file: %w", err)
		}
	}

	return nil
}

// Load loads the configuration from file
func (m *Manager) Load() (*Config, error) {
	if !m.configExists() {
		return nil, fmt.Errorf("configuration file not found: %s", m.configPath)
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	m.config = &config
	return &config, nil
}

// Save saves the configuration to file
func (m *Manager) Save(config *Config) error {
	config.UpdatedAt = time.Now()
	return m.saveConfig(config)
}

// AddApp adds a new application configuration
func (m *Manager) AddApp(appConfig *AppConfig) error {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	m.config.Apps[appConfig.Name] = appConfig
	return m.Save(m.config)
}

// RemoveApp removes an application configuration
func (m *Manager) RemoveApp(appName string) error {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	delete(m.config.Apps, appName)
	return m.Save(m.config)
}

// GetApp retrieves an application configuration
func (m *Manager) GetApp(appName string) (*AppConfig, error) {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	app, exists := m.config.Apps[appName]
	if !exists {
		return nil, fmt.Errorf("application %s not found", appName)
	}

	return app, nil
}

// ListApps returns all application configurations
func (m *Manager) ListApps() (map[string]*AppConfig, error) {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return m.config.Apps, nil
}

// UpdateLastSync updates the last sync timestamp
func (m *Manager) UpdateLastSync() error {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	m.config.LastSync = time.Now()
	return m.Save(m.config)
}

// GetStorePath returns the path to the central store
func (m *Manager) GetStorePath() (string, error) {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return "", fmt.Errorf("failed to load config: %w", err)
		}
	}

	return m.config.StorePath, nil
}

// GetBackupPath returns the path to the backup directory
func (m *Manager) GetBackupPath() (string, error) {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return "", fmt.Errorf("failed to load config: %w", err)
		}
	}

	return m.config.BackupPath, nil
}

// GetSettings returns the global settings
func (m *Manager) GetSettings() (*Settings, error) {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return m.config.Settings, nil
}

// UpdateSettings updates the global settings
func (m *Manager) UpdateSettings(settings *Settings) error {
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	m.config.Settings = settings
	return m.Save(m.config)
}

// ConfigExists checks if the configuration file exists
func (m *Manager) ConfigExists() bool {
	return m.configExists()
}

// GetConfigDir returns the configuration directory path
func (m *Manager) GetConfigDir() string {
	return m.configDir
}

// ConfigPath returns the configuration file path
func (m *Manager) ConfigPath() string {
	return m.configPath
}

// private methods

func (m *Manager) configExists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}

func (m *Manager) saveConfig(config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.config = config
	return nil
}

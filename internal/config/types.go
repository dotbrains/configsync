package config

import (
	"time"
)

// Config represents the main configuration for ConfigSync
type Config struct {
	LastSync   time.Time             `yaml:"last_sync,omitempty"`
	CreatedAt  time.Time             `yaml:"created_at"`
	UpdatedAt  time.Time             `yaml:"updated_at"`
	Apps       map[string]*AppConfig `yaml:"apps"`
	Settings   *Settings             `yaml:"settings"`
	Version    string                `yaml:"version"`
	StorePath  string                `yaml:"store_path"`
	BackupPath string                `yaml:"backup_path"`
	LogPath    string                `yaml:"log_path"`
}

// AppConfig represents configuration for a single application
type AppConfig struct {
	AddedAt      time.Time         `yaml:"added_at"`
	LastSynced   time.Time         `yaml:"last_synced,omitempty"`
	Metadata     map[string]string `yaml:"metadata,omitempty"`
	Name         string            `yaml:"name"`
	DisplayName  string            `yaml:"display_name"`
	BundleID     string            `yaml:"bundle_id,omitempty"`
	Paths        []Path            `yaml:"paths"`
	Enabled      bool              `yaml:"enabled"`
	BackupBefore bool              `yaml:"backup_before"`
}

// Path represents a configuration file or directory path within an application config
type Path struct {
	SyncedAt    time.Time `yaml:"synced_at,omitempty"`
	Source      string    `yaml:"source"`      // Original path (e.g., ~/Library/Preferences/com.app.plist)
	Destination string    `yaml:"destination"` // Path in central store
	Type        PathType  `yaml:"type"`        // file, directory, or glob
	Required    bool      `yaml:"required"`    // Whether this path must exist
	BackedUp    bool      `yaml:"backed_up"`   // Whether original was backed up
	Synced      bool      `yaml:"synced"`      // Whether currently synced
}

// PathType represents the type of configuration path
type PathType string

const (
	// PathTypeFile represents a configuration file
	PathTypeFile PathType = "file"
	// PathTypeDirectory represents a configuration directory
	PathTypeDirectory PathType = "directory"
	// PathTypeGlob represents a configuration glob pattern
	PathTypeGlob PathType = "glob"
)

// Settings represents global settings for ConfigSync
type Settings struct {
	SymlinkMode      string   `yaml:"symlink_mode"`
	ConflictStrategy string   `yaml:"conflict_strategy"`
	ExcludePatterns  []string `yaml:"exclude_patterns"`
	AutoBackup       bool     `yaml:"auto_backup"`
	DryRun           bool     `yaml:"dry_run"`
	VerboseLogging   bool     `yaml:"verbose_logging"`
}

// SyncStatus represents the status of configuration synchronization
type SyncStatus struct {
	LastChecked time.Time `yaml:"last_checked"`
	AppName     string    `yaml:"app_name"`
	Status      string    `yaml:"status"` // "synced", "out_of_sync", "missing", "error"
	Message     string    `yaml:"message,omitempty"`
}

// BackupInfo represents information about a backup
type BackupInfo struct {
	CreatedAt    time.Time `yaml:"created_at"`
	AppName      string    `yaml:"app_name"`
	OriginalPath string    `yaml:"original_path"`
	BackupPath   string    `yaml:"backup_path"`
	Checksum     string    `yaml:"checksum,omitempty"`
	Size         int64     `yaml:"size"`
}

// DeploymentBundle represents a bundle of configurations for deployment
type DeploymentBundle struct {
	CreatedAt time.Time             `yaml:"created_at"`
	Apps      map[string]*AppConfig `yaml:"apps"`
	Metadata  map[string]string     `yaml:"metadata,omitempty"`
	Version   string                `yaml:"version"`
	CreatedBy string                `yaml:"created_by"`
}

// NewDefaultConfig creates a new configuration with default settings
func NewDefaultConfig(storePath, backupPath, logPath string) *Config {
	now := time.Now()
	return &Config{
		Version:    "1.0",
		StorePath:  storePath,
		BackupPath: backupPath,
		LogPath:    logPath,
		Apps:       make(map[string]*AppConfig),
		Settings: &Settings{
			AutoBackup:       true,
			DryRun:           false,
			VerboseLogging:   false,
			SymlinkMode:      "soft",
			ExcludePatterns:  []string{".DS_Store", "*.tmp", "*.log"},
			ConflictStrategy: "ask",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewAppConfig creates a new application configuration
func NewAppConfig(name, displayName string) *AppConfig {
	return &AppConfig{
		Name:         name,
		DisplayName:  displayName,
		Paths:        []Path{},
		Enabled:      true,
		BackupBefore: true,
		Metadata:     make(map[string]string),
		AddedAt:      time.Now(),
	}
}

// AddPath adds a configuration path to an app config
func (ac *AppConfig) AddPath(source, destination string, pathType PathType, required bool) {
	path := Path{
		Source:      source,
		Destination: destination,
		Type:        pathType,
		Required:    required,
		BackedUp:    false,
		Synced:      false,
	}
	ac.Paths = append(ac.Paths, path)
}

// IsEnabled checks if the app configuration is enabled
func (ac *AppConfig) IsEnabled() bool {
	return ac.Enabled
}

// MarkSynced marks a path as synced
func (cp *Path) MarkSynced() {
	cp.Synced = true
	cp.SyncedAt = time.Now()
}

// MarkBackedUp marks a path as backed up
func (cp *Path) MarkBackedUp() {
	cp.BackedUp = true
}

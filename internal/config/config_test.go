// Package config provides comprehensive tests for configuration management functionality.
package config

import (
	"testing"
	"time"
)

func TestNewDefaultConfig(t *testing.T) {
	storePath := "/test/store"
	backupPath := "/test/backup"
	logPath := "/test/log"

	cfg := NewDefaultConfig(storePath, backupPath, logPath)

	if cfg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cfg.Version)
	}

	if cfg.StorePath != storePath {
		t.Errorf("Expected store path %s, got %s", storePath, cfg.StorePath)
	}

	if cfg.BackupPath != backupPath {
		t.Errorf("Expected backup path %s, got %s", backupPath, cfg.BackupPath)
	}

	if cfg.LogPath != logPath {
		t.Errorf("Expected log path %s, got %s", logPath, cfg.LogPath)
	}

	if cfg.Apps == nil {
		t.Error("Expected Apps map to be initialized")
	}

	if cfg.Settings == nil {
		t.Error("Expected Settings to be initialized")
	}

	if !cfg.Settings.AutoBackup {
		t.Error("Expected AutoBackup to be true by default")
	}

	if cfg.Settings.DryRun {
		t.Error("Expected DryRun to be false by default")
	}
}

func TestNewAppConfig(t *testing.T) {
	name := "testapp"
	displayName := "Test App"

	app := NewAppConfig(name, displayName)

	if app.Name != name {
		t.Errorf("Expected name %s, got %s", name, app.Name)
	}

	if app.DisplayName != displayName {
		t.Errorf("Expected display name %s, got %s", displayName, app.DisplayName)
	}

	if !app.Enabled {
		t.Error("Expected app to be enabled by default")
	}

	if !app.BackupBefore {
		t.Error("Expected BackupBefore to be true by default")
	}

	if app.Paths == nil {
		t.Error("Expected Paths to be initialized")
	}

	if app.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestAppConfigAddPath(t *testing.T) {
	app := NewAppConfig("test", "Test")

	source := "/test/source"
	destination := "/test/dest"
	pathType := PathTypeFile
	required := true

	app.AddPath(source, destination, pathType, required)

	if len(app.Paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(app.Paths))
	}

	path := app.Paths[0]
	if path.Source != source {
		t.Errorf("Expected source %s, got %s", source, path.Source)
	}

	if path.Destination != destination {
		t.Errorf("Expected destination %s, got %s", destination, path.Destination)
	}

	if path.Type != pathType {
		t.Errorf("Expected type %s, got %s", pathType, path.Type)
	}

	if path.Required != required {
		t.Errorf("Expected required %t, got %t", required, path.Required)
	}

	if path.BackedUp {
		t.Error("Expected BackedUp to be false initially")
	}

	if path.Synced {
		t.Error("Expected Synced to be false initially")
	}
}

func TestConfigPathMarkSynced(t *testing.T) {
	path := &ConfigPath{
		Source:      "/test/source",
		Destination: "/test/dest",
		Type:        PathTypeFile,
		Required:    false,
	}

	before := time.Now()
	path.MarkSynced()
	after := time.Now()

	if !path.Synced {
		t.Error("Expected path to be marked as synced")
	}

	if path.SyncedAt.IsZero() {
		t.Error("Expected SyncedAt to be set")
	}

	if path.SyncedAt.Before(before) || path.SyncedAt.After(after) {
		t.Error("Expected SyncedAt to be within reasonable time range")
	}
}

func TestConfigPathMarkBackedUp(t *testing.T) {
	path := &ConfigPath{
		Source:      "/test/source",
		Destination: "/test/dest",
		Type:        PathTypeFile,
		Required:    false,
	}

	path.MarkBackedUp()

	if !path.BackedUp {
		t.Error("Expected path to be marked as backed up")
	}
}

func TestAppConfigIsEnabled(t *testing.T) {
	app := NewAppConfig("test", "Test")

	if !app.IsEnabled() {
		t.Error("Expected new app to be enabled")
	}

	app.Enabled = false
	if app.IsEnabled() {
		t.Error("Expected disabled app to return false")
	}
}

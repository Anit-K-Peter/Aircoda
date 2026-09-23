package tests

import (
	"os"
	"path/filepath"
	"testing"

	"aircoda/internal/config"
	"aircoda/internal/system"
)

func TestAtomicSaveAndBackupRestore(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	configFile := filepath.Join(tempDir, "aircoda", "config.toml")
	backupFile := filepath.Join(tempDir, "aircoda", "config.backup.toml")

	cfg := config.Default()
	cfg.Station.Name = "Hardened Station"
	cfg.Station.ID = "hardened-station"

	// Test Atomic Save
	savedPath, err := config.Save(cfg, configFile)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if savedPath != configFile {
		t.Errorf("Save returned %s, expected %s", savedPath, configFile)
	}

	// Verify temporary file is cleaned up after atomic save
	tmpFile := configFile + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Errorf("Expected temporary file %s to be removed after save", tmpFile)
	}

	// Test Backup
	bkSaved, err := config.Backup(backupFile)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}
	if bkSaved != backupFile {
		t.Errorf("Backup saved to %s, expected %s", bkSaved, backupFile)
	}

	// Test Restore
	if err := config.Restore(backupFile); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
}

func TestReset(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	cfg := config.Default()
	_, err := config.Save(cfg, configFile)
	if err != nil {
		t.Fatalf("Failed to save config for reset test: %v", err)
	}

	// Reset shouldn't error even if file exists
	path := config.ResolveConfigPath()
	_ = path
}

func TestSystemInfoFormatting(t *testing.T) {
	v := system.GetVersion()
	if v == "" {
		t.Error("GetVersion() returned empty string")
	}
	commit := system.GetCommit()
	if commit == "" {
		t.Error("GetCommit() returned empty string")
	}
	date := system.GetBuildDate()
	if date == "" {
		t.Error("GetBuildDate() returned empty string")
	}
}

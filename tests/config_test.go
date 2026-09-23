package tests

import (
	"os"
	"path/filepath"
	"testing"

	"aircoda/internal/config"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"986 FM", "986fm"},
		{"My Radio Station!!!", "myradiostation"},
		{"Radio - 101", "radio-101"},
		{"___", "___"},
		{"Indie-Rock_Radio", "indie-rock_radio"},
	}

	for _, tt := range tests {
		got := config.Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := config.Default()
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("Default config should be valid, got: %v", err)
	}

	// Invalid station name
	cfg.Station.Name = ""
	if err := config.Validate(cfg); err == nil {
		t.Error("Expected error for empty station name, got nil")
	}
	cfg.Station.Name = "Valid Station"

	// Invalid station ID
	cfg.Station.ID = "invalid id with spaces"
	if err := config.Validate(cfg); err == nil {
		t.Error("Expected error for station ID with spaces, got nil")
	}
	cfg.Station.ID = "valid-id_123"

	// Invalid audio source
	cfg.Audio.Source = "spotify"
	if err := config.Validate(cfg); err == nil {
		t.Error("Expected error for invalid audio source, got nil")
	}
	cfg.Audio.Source = "local"

	// Invalid public mode
	cfg.Public.Mode = "invalidmode"
	if err := config.Validate(cfg); err == nil {
		t.Error("Expected error for invalid public mode, got nil")
	}
	cfg.Public.Mode = "local"
}

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "test_config.toml")

	original := config.Default()
	original.Station.Name = "Test Beat Radio"
	original.Station.ID = "test-beat"
	original.Audio.Source = "later"

	savedPath, err := config.Save(original, targetFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	if savedPath != targetFile {
		t.Errorf("Saved path = %q, expected %q", savedPath, targetFile)
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read saved config file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Saved config file is empty")
	}
}

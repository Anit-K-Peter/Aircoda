package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var (
	ErrConfigNotFound = errors.New("configuration file not found")
	validIDRegex      = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Default returns a default initial Config structure.
func Default() *Config {
	return &Config{
		Version: 1,
		Station: StationConfig{
			Name: "986 FM",
			ID:   "986fm",
		},
		Audio: AudioConfig{
			Source:    "local",
			Directory: "",
		},
		Playlist: PlaylistConfig{
			Mode:              "shuffle",
			Repeat:            true,
			Recursive:         true,
			StationIDEnabled:  false,
			StationIDInterval: 5,
			StationIDPath:     "",
		},
		Schedule: ScheduleConfig{
			Enabled:       false,
			MorningPath:   "",
			AfternoonPath: "",
			EveningPath:   "",
			NightPath:     "",
		},
		Streaming: StreamingConfig{
			Host:           "127.0.0.1",
			Port:           8000,
			Mount:          "/986fm",
			Bitrate:        128,
			Format:         "mp3",
			SourcePassword: "hackme",
			AdminPassword:  "hackme",
		},
		Public: PublicConfig{
			Mode:     "local",
			Hostname: "",
		},
		Tunnel: TunnelConfig{
			Enabled:    false,
			Provider:   "cloudflare",
			Hostname:   "",
			Target:     "127.0.0.1:8000",
			ConfigPath: "",
		},
		System: SystemConfig{
			StartOnBoot: true,
			LogLevel:    "info",
		},
	}
}

// Load reads and parses configuration from the resolved path.
func Load() (*Config, string, error) {
	path := ResolveConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, path, ErrConfigNotFound
		}
		return nil, path, fmt.Errorf("failed to read config file (%s): %w", path, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, path, fmt.Errorf("configuration file (%s) is invalid: %w", path, err)
	}

	// Ensure sensible playlist defaults for older config schemas
	if cfg.Playlist.Mode == "" {
		cfg.Playlist.Mode = "shuffle"
		cfg.Playlist.Repeat = true
		cfg.Playlist.Recursive = true
	}
	if cfg.Playlist.StationIDInterval <= 0 {
		cfg.Playlist.StationIDInterval = 5
	}

	// Ensure sensible streaming defaults
	if cfg.Streaming.Host == "" {
		cfg.Streaming.Host = "127.0.0.1"
	}
	if cfg.Streaming.Port == 0 {
		cfg.Streaming.Port = 8000
	}
	if cfg.Streaming.Mount == "" {
		mountSlug := Slugify(cfg.Station.Name)
		if mountSlug == "" {
			mountSlug = "station"
		}
		cfg.Streaming.Mount = "/" + mountSlug
	}
	if cfg.Streaming.Bitrate == 0 {
		cfg.Streaming.Bitrate = 128
	}
	if cfg.Streaming.Format == "" {
		cfg.Streaming.Format = "mp3"
	}
	if cfg.Streaming.SourcePassword == "" {
		cfg.Streaming.SourcePassword = "hackme"
	}
	if cfg.Streaming.AdminPassword == "" {
		cfg.Streaming.AdminPassword = "hackme"
	}

	// Ensure sensible public & tunnel defaults
	if cfg.Public.Mode == "" {
		cfg.Public.Mode = "local"
	}
	if cfg.Tunnel.Provider == "" {
		cfg.Tunnel.Provider = "cloudflare"
	}
	if cfg.Tunnel.Target == "" {
		cfg.Tunnel.Target = fmt.Sprintf("%s:%d", cfg.Streaming.Host, cfg.Streaming.Port)
	}

	return &cfg, path, nil
}

// Save writes configuration to specified or resolved path atomically.
func Save(cfg *Config, targetPath string) (string, error) {
	if targetPath == "" {
		targetPath = ResolveConfigPath()
	}

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return targetPath, fmt.Errorf("cannot create configuration directory (%s): %w", dir, err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return targetPath, fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Atomic save via temporary file
	tmpPath := targetPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return targetPath, fmt.Errorf("cannot write temporary configuration file (%s): %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return targetPath, fmt.Errorf("cannot update configuration file (%s): %w", targetPath, err)
	}

	return targetPath, nil
}

// Backup creates a copy of the active configuration file.
func Backup(targetPath string) (string, error) {
	cfg, currentPath, err := Load()
	if err != nil {
		return "", fmt.Errorf("cannot backup unreadable configuration: %w", err)
	}

	if targetPath == "" {
		dir := filepath.Dir(currentPath)
		targetPath = filepath.Join(dir, "config.backup.toml")
	}

	return Save(cfg, targetPath)
}

// Restore validates and restores configuration from a backup file.
func Restore(backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("cannot read backup file (%s): %w", backupPath, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("backup configuration file (%s) is invalid: %w", backupPath, err)
	}

	if err := Validate(&cfg); err != nil {
		return fmt.Errorf("backup configuration validation failed: %w", err)
	}

	_, err = Save(&cfg, ResolveConfigPath())
	return err
}

// Reset removes station configuration file while preserving audio files.
func Reset() error {
	path := ResolveConfigPath()
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to reset configuration file (%s): %w", path, err)
		}
	}
	return nil
}

// Validate checks the validity of configuration fields.
func Validate(cfg *Config) error {
	if cfg == nil {
		return errors.New("configuration is nil")
	}

	if strings.TrimSpace(cfg.Station.Name) == "" {
		return errors.New("station name cannot be empty")
	}

	stationID := strings.TrimSpace(cfg.Station.ID)
	if stationID == "" {
		return errors.New("station ID cannot be empty")
	}
	if !validIDRegex.MatchString(stationID) {
		return errors.New("station ID may only contain letters, numbers, hyphens, and underscores")
	}

	switch cfg.Audio.Source {
	case "local", "youtube", "later":
		// valid options
	default:
		return fmt.Errorf("invalid audio source '%s'", cfg.Audio.Source)
	}

	if cfg.Playlist.Mode == "" {
		cfg.Playlist.Mode = "shuffle"
	}
	switch cfg.Playlist.Mode {
	case "shuffle", "sequential":
		// valid modes
	default:
		return fmt.Errorf("invalid playlist mode '%s', must be 'shuffle' or 'sequential'", cfg.Playlist.Mode)
	}

	if cfg.Public.Mode == "" {
		cfg.Public.Mode = "local"
	}
	switch cfg.Public.Mode {
	case "local", "direct", "tunnel":
		// valid modes
	default:
		return fmt.Errorf("invalid public mode '%s', must be 'local', 'direct', or 'tunnel'", cfg.Public.Mode)
	}

	if cfg.Streaming.Port <= 0 || cfg.Streaming.Port > 65535 {
		return fmt.Errorf("invalid streaming port %d, must be between 1 and 65535", cfg.Streaming.Port)
	}
	if cfg.Streaming.Bitrate <= 0 {
		return fmt.Errorf("invalid streaming bitrate %d", cfg.Streaming.Bitrate)
	}

	if cfg.Audio.Source == "local" && cfg.Audio.Directory != "" {
		dirInfo, err := os.Stat(cfg.Audio.Directory)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("audio directory does not exist: %s", cfg.Audio.Directory)
			}
			return fmt.Errorf("cannot access audio directory: %w", err)
		}
		if !dirInfo.IsDir() {
			return fmt.Errorf("audio path is not a directory: %s", cfg.Audio.Directory)
		}
	}

	return nil
}

// Slugify converts station name into valid station ID slug (e.g., "986 FM" -> "986fm").
func Slugify(name string) string {
	lower := strings.ToLower(name)
	var sb strings.Builder
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			sb.WriteRune(r)
		}
	}
	result := sb.String()
	if result == "" {
		return "station"
	}
	return result
}

package config

// Config represents the root Aircoda configuration structure.
type Config struct {
	Version   int             `toml:"version"`
	Station   StationConfig   `toml:"station"`
	Audio     AudioConfig     `toml:"audio"`
	Playlist  PlaylistConfig  `toml:"playlist"`
	Schedule  ScheduleConfig `toml:"schedule"`
	Streaming StreamingConfig `toml:"streaming"`
	Public    PublicConfig    `toml:"public"`
	Tunnel    TunnelConfig    `toml:"tunnel"`
	System    SystemConfig    `toml:"system"`
}

// StationConfig defines station metadata.
type StationConfig struct {
	Name string `toml:"name"`
	ID   string `toml:"id"`
}

// AudioConfig defines audio source settings.
type AudioConfig struct {
	Source    string `toml:"source"`    // "local", "youtube", "later"
	Directory string `toml:"directory"` // Path to local audio files
}

// PlaylistConfig defines playlist playback engine preferences.
type PlaylistConfig struct {
	Mode              string `toml:"mode"`                // "shuffle", "sequential"
	Repeat            bool   `toml:"repeat"`              // true, false
	Recursive         bool   `toml:"recursive"`           // true, false
	StationIDEnabled  bool   `toml:"station_id_enabled"`  // true, false
	StationIDInterval int    `toml:"station_id_interval"` // e.g. 5 tracks
	StationIDPath     string `toml:"station_id_path"`     // audio file or directory path
}

// ScheduleConfig defines time-based playlist scheduling (dayparting).
type ScheduleConfig struct {
	Enabled       bool   `toml:"enabled"`        // true, false
	MorningPath   string `toml:"morning_path"`   // 06:00 - 12:00
	AfternoonPath string `toml:"afternoon_path"` // 12:00 - 18:00
	EveningPath   string `toml:"evening_path"`   // 18:00 - 00:00
	NightPath     string `toml:"night_path"`     // 00:00 - 06:00
}

// StreamingConfig defines streaming server and audio engine settings.
type StreamingConfig struct {
	Host           string `toml:"host"`
	Port           int    `toml:"port"`
	Mount          string `toml:"mount"`
	Bitrate        int    `toml:"bitrate"` // in kbps (e.g. 128)
	Format         string `toml:"format"`  // "mp3"
	SourcePassword string `toml:"source_password"`
	AdminPassword  string `toml:"admin_password"`
}

// PublicConfig defines public stream exposure settings.
type PublicConfig struct {
	Mode     string `toml:"mode"`     // "local", "direct", "tunnel"
	Hostname string `toml:"hostname"` // Optional domain or public IP
}

// TunnelConfig defines Cloudflare tunnel parameters.
type TunnelConfig struct {
	Enabled    bool   `toml:"enabled"`
	Provider   string `toml:"provider"`    // "cloudflare"
	Hostname   string `toml:"hostname"`    // e.g. "radio.example.com"
	Target     string `toml:"target"`      // e.g. "127.0.0.1:8000"
	ConfigPath string `toml:"config_path"` // Optional path to cloudflared config
}

// SystemConfig defines system-level parameters.
type SystemConfig struct {
	StartOnBoot bool   `toml:"start_on_boot"`
	LogLevel    string `toml:"log_level"`
}

package playlist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aircoda/internal/config"
	"aircoda/internal/library"
)

const queueStateFileName = "queue.json"

// ResolveQueueStatePath returns path to persistent queue JSON file.
func ResolveQueueStatePath() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "aircoda", queueStateFileName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	return filepath.Join(home, ".local", "state", "aircoda", queueStateFileName)
}

// LoadState reads persistent queue state from disk.
func LoadState() (*QueueState, error) {
	path := ResolveQueueStatePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var qs QueueState
	if err := json.Unmarshal(data, &qs); err != nil {
		return nil, err
	}
	return &qs, nil
}

// SaveState writes queue state to disk.
func SaveState(qs *QueueState) error {
	path := ResolveQueueStatePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create queue state directory: %w", err)
	}
	data, err := json.MarshalIndent(qs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize queue state: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ClearState removes persistent queue state JSON file.
func ClearState() {
	_ = os.Remove(ResolveQueueStatePath())
}

// Manager orchestrates library scanning and persistent queue state.
type Manager struct {
	Config *config.Config
}

// NewManager constructs a playlist Manager.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{Config: cfg}
}

// ResolveActiveAudioDirectory returns configured Audio.Directory or current daypart directory if Schedule.Enabled is true.
func ResolveActiveAudioDirectory(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	if !cfg.Schedule.Enabled {
		return cfg.Audio.Directory
	}

	hour := time.Now().Hour()
	var daypartDir string

	switch {
	case hour >= 6 && hour < 12:
		daypartDir = cfg.Schedule.MorningPath
	case hour >= 12 && hour < 18:
		daypartDir = cfg.Schedule.AfternoonPath
	case hour >= 18 && hour < 24:
		daypartDir = cfg.Schedule.EveningPath
	default: // 00:00 - 05:59
		daypartDir = cfg.Schedule.NightPath
	}

	daypartDir = strings.TrimSpace(daypartDir)
	if daypartDir != "" {
		if info, err := os.Stat(daypartDir); err == nil && info.IsDir() {
			return daypartDir
		}
	}

	return cfg.Audio.Directory
}

// GetOrInitQueue loads cached queue state or performs fresh scan.
func (m *Manager) GetOrInitQueue() (*QueueState, []library.Track, error) {
	activeDir := ResolveActiveAudioDirectory(m.Config)
	if strings.TrimSpace(activeDir) == "" {
		return nil, nil, fmt.Errorf("audio directory is not configured. Run 'radio setup'")
	}

	tracks, scanErr := library.ScanDirectory(activeDir, m.Config.Playlist.Recursive)
	if scanErr != nil {
		return nil, nil, scanErr
	}

	qs, err := LoadState()
	if err == nil && len(qs.Tracks) > 0 {
		// Verify cached mode/repeat match config
		qs.Mode = m.Config.Playlist.Mode
		qs.Repeat = m.Config.Playlist.Repeat

		// Check if tracks set changed
		if len(qs.Tracks) == len(tracks) {
			match := true
			for i := range tracks {
				if tracks[i].Path != qs.Tracks[i].Path {
					match = false
					break
				}
			}
			if match {
				return qs, tracks, nil
			}
		}
	}

	// Create fresh queue state
	freshQs := NewQueueState(tracks, m.Config.Playlist.Mode, m.Config.Playlist.Repeat)
	_ = SaveState(freshQs)
	return freshQs, tracks, nil
}

// Rescan forces a directory rescan and updates queue state.
func (m *Manager) Rescan() (*QueueState, []library.Track, error) {
	activeDir := ResolveActiveAudioDirectory(m.Config)
	if strings.TrimSpace(activeDir) == "" {
		return nil, nil, fmt.Errorf("audio directory is not configured")
	}

	tracks, err := library.ScanDirectory(activeDir, m.Config.Playlist.Recursive)
	if err != nil {
		return nil, nil, err
	}

	freshQs := NewQueueState(tracks, m.Config.Playlist.Mode, m.Config.Playlist.Repeat)
	if saveErr := SaveState(freshQs); saveErr != nil {
		// Log warning but don't fail return
	}

	return freshQs, tracks, nil
}

// Current returns current queue track.
func (m *Manager) Current() (*library.Track, error) {
	qs, _, err := m.GetOrInitQueue()
	if err != nil {
		return nil, err
	}
	track, ok := qs.Current()
	if !ok {
		return nil, fmt.Errorf("no tracks available in playlist")
	}
	return track, nil
}

// Next advances queue to next track and persists state.
func (m *Manager) Next() (*library.Track, error) {
	qs, _, err := m.GetOrInitQueue()
	if err != nil {
		return nil, err
	}
	track, ok := qs.Next()
	if !ok {
		return nil, fmt.Errorf("end of playlist reached")
	}
	_ = SaveState(qs)
	return track, nil
}

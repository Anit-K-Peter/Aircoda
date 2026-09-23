package audio

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

const stateFileName = "aircoda_state.json"

// ServerState encapsulates runtime process and streaming metrics.
type ServerState struct {
	IsLive       bool      `json:"is_live"`
	DaemonPID    int       `json:"daemon_pid"`
	IcecastPID   int       `json:"icecast_pid"`
	FFmpegPID    int       `json:"ffmpeg_pid"`
	StartTime    time.Time `json:"start_time"`
	CurrentTrack string    `json:"current_track"`
	NextTrack    string    `json:"next_track"`
	StreamURL    string    `json:"stream_url"`
	Bitrate      int       `json:"bitrate"`
	Listeners    string    `json:"listeners"`
}

// ResolveStatePath returns path to runtime state file.
func ResolveStatePath() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "aircoda", stateFileName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	return filepath.Join(home, ".local", "state", "aircoda", stateFileName)
}

// LoadServerState loads active runtime state from disk.
func LoadServerState() (*ServerState, error) {
	path := ResolveStatePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state ServerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveServerState writes active runtime state to disk.
func SaveServerState(state *ServerState) error {
	path := ResolveStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ClearServerState resets/removes runtime state.
func ClearServerState() {
	_ = os.Remove(ResolveStatePath())
}

// IsProcessRunning checks if a PID is currently alive on Linux.
func IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks if process exists without sending a signal
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// CheckPortAvailable verifies whether host:port is free.
func CheckPortAvailable(host string, port int) error {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	for i := 0; i < 3; i++ {
		conn, err := net.DialTimeout("tcp", address, 300*time.Millisecond)
		if err != nil {
			return nil
		}
		conn.Close()
		if i < 2 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return fmt.Errorf("Port %d is already in use.", port)
}

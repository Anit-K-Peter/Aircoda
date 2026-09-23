package tunnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"aircoda/internal/config"
	"aircoda/internal/logger"
)

var (
	ErrCloudflaredNotFound = errors.New("cloudflared binary not found")
	tunnelMutex            sync.Mutex
	tryCloudflareRegex     = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
)

// TunnelState represents saved runtime status of the Cloudflare tunnel process.
type TunnelState struct {
	IsRunning bool      `json:"is_running"`
	PID       int       `json:"pid"`
	Hostname  string    `json:"hostname"`
	PublicURL string    `json:"public_url,omitempty"`
	Target    string    `json:"target"`
	StartTime time.Time `json:"start_time"`
	Error     string    `json:"error,omitempty"`
}

// CheckCloudflared verifies if cloudflared is installed or auto-downloads it.
func CheckCloudflared() (string, error) {
	return EnsureCloudflared()
}

// GetStatePath returns the resolved path to tunnel_state.json.
func GetStatePath() string {
	configDir := filepath.Dir(config.ResolveConfigPath())
	return filepath.Join(configDir, "tunnel_state.json")
}

// LoadState reads state from tunnel_state.json.
func LoadState() (*TunnelState, error) {
	path := GetStatePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state TunnelState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveState writes current state to tunnel_state.json.
func SaveState(state *TunnelState) error {
	path := GetStatePath()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ClearState removes tunnel_state.json.
func ClearState() {
	_ = os.Remove(GetStatePath())
}

// IsProcessRunning checks if process ID is active.
func IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	if runtime.GOOS == "windows" {
		proc, err := os.FindProcess(pid)
		if err != nil {
			return false
		}
		_ = proc
		return true
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// Start launches cloudflared tunnel process in background and resolves public URL.
func Start(cfg *config.Config) error {
	tunnelMutex.Lock()
	defer tunnelMutex.Unlock()

	cloudflaredBin, err := EnsureCloudflared()
	if err != nil {
		return fmt.Errorf("Cloudflare tunnel dependency error: %w", err)
	}

	state, err := LoadState()
	if err == nil && state != nil && state.IsRunning && IsProcessRunning(state.PID) {
		if state.PublicURL != "" {
			cfg.Public.Mode = "tunnel"
			cfg.Public.Hostname = state.PublicURL
			cfg.Tunnel.Enabled = true
			_, _ = config.Save(cfg, "")
		}
		return nil
	}

	target := cfg.Tunnel.Target
	if target == "" {
		target = fmt.Sprintf("127.0.0.1:%d", cfg.Streaming.Port)
	}
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}

	var args []string
	if cfg.Tunnel.Hostname != "" {
		args = []string{"tunnel", "run", "--url", target}
	} else {
		// Quick tunnel mode
		args = []string{"tunnel", "--url", target}
	}

	logPath := filepath.Join(filepath.Dir(config.ResolveConfigPath()), "tunnel.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	cmd := exec.Command(cloudflaredBin, args...)
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start cloudflared process: %w", err)
	}

	// Poll log file for generated .trycloudflare.com URL
	publicURL := ""
	if cfg.Tunnel.Hostname != "" {
		publicURL = "https://" + cfg.Tunnel.Hostname
	} else {
		publicURL = pollForPublicURL(logPath, 12*time.Second)
	}

	newState := &TunnelState{
		IsRunning: true,
		PID:       cmd.Process.Pid,
		Hostname:  cfg.Tunnel.Hostname,
		PublicURL: publicURL,
		Target:    target,
		StartTime: time.Now(),
	}

	if publicURL != "" {
		cfg.Public.Mode = "tunnel"
		cfg.Public.Hostname = publicURL
		cfg.Tunnel.Enabled = true
		_, _ = config.Save(cfg, "")
	}

	_ = SaveState(newState)
	logger.Info("Cloudflare tunnel process started (PID %d) -> Public URL: %s", cmd.Process.Pid, publicURL)
	return nil
}

func pollForPublicURL(logPath string, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		data, err := os.ReadFile(logPath)
		if err != nil {
			continue
		}
		matches := tryCloudflareRegex.FindString(string(data))
		if matches != "" {
			return matches
		}
	}
	return ""
}

// Stop cleanly terminates cloudflared tunnel process.
func Stop() error {
	tunnelMutex.Lock()
	defer tunnelMutex.Unlock()

	state, err := LoadState()
	if err != nil || state == nil || state.PID <= 0 {
		ClearState()
		return nil
	}

	if IsProcessRunning(state.PID) {
		proc, err := os.FindProcess(state.PID)
		if err == nil {
			_ = proc.Kill()
		}
	}

	ClearState()
	logger.Info("Cloudflare tunnel stopped.")
	return nil
}

// Restart stops and starts the Cloudflare tunnel.
func Restart(cfg *config.Config) error {
	_ = Stop()
	time.Sleep(500 * time.Millisecond)
	return Start(cfg)
}

// GetStatus returns current runtime status of Cloudflare tunnel.
func GetStatus(cfg *config.Config) (*TunnelState, error) {
	state, err := LoadState()
	if err != nil || state == nil {
		return &TunnelState{
			IsRunning: false,
			Hostname:  cfg.Tunnel.Hostname,
			Target:    cfg.Tunnel.Target,
		}, nil
	}

	if !IsProcessRunning(state.PID) {
		state.IsRunning = false
	} else if state.PublicURL == "" && cfg.Tunnel.Hostname == "" {
		logPath := filepath.Join(filepath.Dir(config.ResolveConfigPath()), "tunnel.log")
		if data, err := os.ReadFile(logPath); err == nil {
			matches := tryCloudflareRegex.FindString(string(data))
			if matches != "" {
				state.PublicURL = matches
				cfg.Public.Hostname = matches
				cfg.Public.Mode = "tunnel"
				_ = SaveState(state)
				_, _ = config.Save(cfg, "")
			}
		}
	}

	return state, nil
}

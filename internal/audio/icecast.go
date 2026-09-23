package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"aircoda/internal/config"
	"aircoda/internal/logger"
	"aircoda/internal/system"
)

// GenerateIcecastConfig builds minimalist Icecast XML configuration data.
func GenerateIcecastConfig(cfg *config.Config, logDir string) string {
	mount := strings.TrimPrefix(cfg.Streaming.Mount, "/")
	if mount == "" {
		mount = "station"
	}

	return fmt.Sprintf(`<?xml version="1.0"?>
<icecast>
    <location>Aircoda Radio Server</location>
    <admin>admin@localhost</admin>
    <limits>
        <clients>100</clients>
        <sources>10</sources>
        <queue-size>524288</queue-size>
        <client-timeout>30</client-timeout>
        <header-timeout>15</header-timeout>
        <source-timeout>10</source-timeout>
        <burst-size>65535</burst-size>
    </limits>
    <authentication>
        <source-password>%s</source-password>
        <relay-password>%s</relay-password>
        <admin-user>admin</admin-user>
        <admin-password>%s</admin-password>
    </authentication>
    <hostname>%s</hostname>
    <listen-socket>
        <port>%d</port>
        <bind-address>%s</bind-address>
    </listen-socket>
    <paths>
        <logdir>%s</logdir>
    </paths>
    <logging>
        <accesslog>access.log</accesslog>
        <errorlog>error.log</errorlog>
        <loglevel>3</loglevel>
    </logging>
    <mount type="normal">
        <mount-name>/%s</mount-name>
        <stream-name>%s</stream-name>
        <stream-description>Aircoda Radio Broadcast</stream-description>
        <genre>Radio</genre>
    </mount>
</icecast>`,
		cfg.Streaming.SourcePassword,
		cfg.Streaming.SourcePassword,
		cfg.Streaming.AdminPassword,
		cfg.Streaming.Host,
		cfg.Streaming.Port,
		cfg.Streaming.Host,
		logDir,
		mount,
		cfg.Station.Name,
	)
}

// IcecastProcess manages an Icecast process instance.
type IcecastProcess struct {
	cmd *exec.Cmd
}

// StartIcecast creates XML configuration and starts Icecast process.
func StartIcecast(cfg *config.Config) (*IcecastProcess, error) {
	binary, err := exec.LookPath("icecast")
	if err != nil {
		binary, err = exec.LookPath("icecast2")
		if err != nil {
			return nil, fmt.Errorf("Icecast binary not found in PATH")
		}
	}

	stateDir := filepath.Dir(ResolveStatePath())
	logDir := filepath.Join(stateDir, "icecast_logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create Icecast log directory: %w", err)
	}

	xmlPath := filepath.Join(stateDir, "icecast.xml")
	xmlContent := GenerateIcecastConfig(cfg, logDir)
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Icecast config XML: %w", err)
	}

	cmd := exec.Command(binary, "-c", xmlPath)
	cmd.SysProcAttr = system.GetDetachSysProcAttr()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Icecast server process: %w", err)
	}

	logger.Info("Icecast server process started with PID %d on %s:%d", cmd.Process.Pid, cfg.Streaming.Host, cfg.Streaming.Port)

	// Wait briefly to ensure Icecast binds to port
	time.Sleep(500 * time.Millisecond)
	return &IcecastProcess{cmd: cmd}, nil
}

// Stop terminates the Icecast process cleanly.
func (p *IcecastProcess) Stop() {
	if p != nil && p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
	}
}

// PID returns Icecast process ID.
func (p *IcecastProcess) PID() int {
	if p != nil && p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}

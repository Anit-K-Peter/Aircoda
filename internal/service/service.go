package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	kardianos "github.com/kardianos/service"

	"aircoda/internal/audio"
	"aircoda/internal/config"
	"aircoda/internal/logger"
	"aircoda/internal/playlist"
	"aircoda/internal/system"
)

// ServiceManager interface abstracts background OS service operations.
type ServiceManager interface {
	Install() error
	Uninstall() error
	Start() error
	Stop() error
	Restart() error
	Status() string
	Enable() error
	Disable() error
	SystemName() string
}

// program implements kardianos.Interface for OS background service runner.
type program struct {
	engine *audio.Engine
}

func (p *program) Start(s kardianos.Service) error {
	cfg, _, err := config.Load()
	if err != nil {
		return err
	}
	mgr := playlist.NewManager(cfg)
	p.engine = audio.NewEngine(cfg, mgr)
	go func() {
		_ = p.engine.Start()
	}()
	return nil
}

func (p *program) Stop(s kardianos.Service) error {
	if p.engine != nil {
		p.engine.Stop()
	}
	return nil
}

// OSManager implements ServiceManager using kardianos/service.
type OSManager struct {
	svc kardianos.Service
}

// NewServiceManager constructs an OS-specific ServiceManager instance.
func NewServiceManager() (ServiceManager, error) {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "radio"
	}
	execPath, _ = filepath.Abs(execPath)

	svcConfig := &kardianos.Config{
		Name:        "aircoda",
		DisplayName: "Aircoda Radio Server",
		Description: "Self-hosted 24/7 internet radio server daemon",
		Executable:  execPath,
		Arguments:   []string{"daemon"},
	}

	prg := &program{}
	s, err := kardianos.New(prg, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OS service manager: %w", err)
	}

	return &OSManager{svc: s}, nil
}

func (m *OSManager) SystemName() string {
	return system.GetServiceSystemName()
}

func (m *OSManager) Install() error {
	if err := m.svc.Install(); err != nil {
		if strings.Contains(err.Error(), "not supported") {
			return fmt.Errorf("Background service installation is not supported on this platform yet.\n\nYou can still run Aircoda manually with:\n  radio start")
		}
		return fmt.Errorf("service installation failed: %w", err)
	}
	logger.Info("Aircoda service successfully installed (%s)", m.SystemName())
	return nil
}

func (m *OSManager) Uninstall() error {
	if err := m.svc.Uninstall(); err != nil {
		return fmt.Errorf("service uninstallation failed: %w", err)
	}
	logger.Info("Aircoda service successfully uninstalled")
	return nil
}

func (m *OSManager) Start() error {
	if err := m.svc.Start(); err != nil {
		return fmt.Errorf("failed to start Aircoda service: %w", err)
	}
	return nil
}

func (m *OSManager) Stop() error {
	if err := m.svc.Stop(); err != nil {
		return fmt.Errorf("failed to stop Aircoda service: %w", err)
	}
	return nil
}

func (m *OSManager) Restart() error {
	if err := m.svc.Restart(); err != nil {
		return fmt.Errorf("failed to restart Aircoda service: %w", err)
	}
	return nil
}

func (m *OSManager) Status() string {
	status, err := m.svc.Status()
	if err != nil {
		return "UNKNOWN"
	}
	switch status {
	case kardianos.StatusRunning:
		return "RUNNING"
	case kardianos.StatusStopped:
		return "STOPPED"
	default:
		return "NOT INSTALLED"
	}
}

func (m *OSManager) Enable() error {
	// Enable automatic startup on boot via system service
	return m.Install()
}

func (m *OSManager) Disable() error {
	return m.Uninstall()
}

// IsServiceSupported returns true if native background service management is supported on current OS.
func IsServiceSupported() bool {
	switch runtime.GOOS {
	case "linux", "windows", "darwin":
		return true
	default:
		return false
	}
}

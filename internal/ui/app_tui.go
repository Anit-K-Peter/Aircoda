package ui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"aircoda/internal/audio"
	"aircoda/internal/config"
	"aircoda/internal/system"
)

// StatusFallbackFunc function pointer to execute plain status when non-TTY
var StatusFallbackFunc func() error

// RunTUI launches the interactive terminal user interface.
func RunTUI(cfg *config.Config, forceSetup bool) error {
	// If executed in non-interactive environment (script/subshell/non-TTY), fallback gracefully
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		if StatusFallbackFunc != nil {
			return StatusFallbackFunc()
		}
		return nil
	}

	if cfg == nil || forceSetup {
		setupModel := NewSetupModel()
		p := tea.NewProgram(
			setupModel,
			tea.WithAltScreen(),
			tea.WithInput(os.Stdin),
			tea.WithOutput(os.Stdout),
		)
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("setup wizard TUI failed: %w", err)
		}

		resModel, ok := finalModel.(*SetupModel)
		if !ok || resModel.Cancelled || !resModel.IsFinished {
			return nil
		}

		// Reload configuration after successful setup
		newCfg, _, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration after setup: %w", err)
		}

		// Automatically start station in background after setup completion
		audio.KillExistingProcesses()
		executable, err := os.Executable()
		if err != nil {
			executable = "radio"
		}
		daemonCmd := exec.Command(executable, "daemon")
		logFile, _ := os.OpenFile(config.ResolveLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if logFile != nil {
			daemonCmd.Stdout = logFile
			daemonCmd.Stderr = logFile
		}
		daemonCmd.SysProcAttr = system.GetDetachSysProcAttr()
		_ = daemonCmd.Start()
		time.Sleep(1 * time.Second)

		cfg = newCfg
	}

	dashboardModel := NewDashboardModel(cfg)
	p := tea.NewProgram(
		dashboardModel,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("main dashboard TUI failed: %w", err)
	}

	return nil
}

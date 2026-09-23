package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"aircoda/internal/audio"
	"aircoda/internal/config"
	"aircoda/internal/deps"
	"aircoda/internal/playlist"
	"aircoda/internal/system"
)

var (
	startForeground bool
	logsFollow      bool
)

// RunStart starts station streaming processes.
func RunStart(foreground bool) error {
	cfg, _, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("Aircoda is not configured yet. Run 'radio setup' to initialize")
		}
		return err
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("configuration file is invalid: %w", err)
	}

	// 1. Dependency check
	if _, err := deps.CheckDependencies(); err != nil {
		return err
	}

	// 2. Audio directory check
	mgr := playlist.NewManager(cfg)
	_, tracks, err := mgr.GetOrInitQueue()
	if err != nil {
		return err
	}
	if len(tracks) == 0 {
		return fmt.Errorf("No playable audio files were found.\n\nAdd supported audio files to:\n  %s\n\nThen run:\n  radio playlist scan", cfg.Audio.Directory)
	}

	// 3. Kill existing processes & check port availability
	audio.KillExistingProcesses()
	if err := audio.CheckPortAvailable(cfg.Streaming.Host, cfg.Streaming.Port); err != nil {
		return err
	}

	// 4. Check if already live
	if state, err := audio.LoadServerState(); err == nil && state.IsLive {
		if audio.IsProcessRunning(state.DaemonPID) || audio.IsProcessRunning(state.IcecastPID) || audio.IsProcessRunning(state.FFmpegPID) {
			return fmt.Errorf("station is already running. Stream: %s", state.StreamURL)
		}
	}

	streamURL := audio.BuildStreamURL(cfg)

	if foreground {
		fmt.Println("Starting station in foreground mode...")
		fmt.Printf("Stream: %s\n", streamURL)
		engine := audio.NewEngine(cfg, mgr)
		return engine.StartBlocking()
	}

	// Spawn background daemon process
	executable, err := os.Executable()
	if err != nil {
		executable = "radio"
	}

	logFile, err := os.OpenFile(config.ResolveLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	var daemonCmd *exec.Cmd
	if err == nil {
		daemonCmd = exec.Command(executable, "daemon")
		daemonCmd.Stdout = logFile
		daemonCmd.Stderr = logFile
	} else {
		daemonCmd = exec.Command(executable, "daemon")
	}
	daemonCmd.SysProcAttr = system.GetDetachSysProcAttr()

	if err := daemonCmd.Start(); err != nil {
		return fmt.Errorf("failed to launch station background daemon: %w", err)
	}

	// Wait briefly for daemon to initialize
	time.Sleep(1 * time.Second)

	fmt.Println("Station started.")
	fmt.Println()
	fmt.Println("Stream:")
	fmt.Println(streamURL)
	return nil
}

// RunDaemon executes the long-running streaming engine in background process.
func RunDaemon() error {
	cfg, _, err := config.Load()
	if err != nil {
		return err
	}
	mgr := playlist.NewManager(cfg)
	engine := audio.NewEngine(cfg, mgr)
	return engine.StartBlocking()
}

// RunStop stops station processes.
func RunStop() error {
	audio.KillExistingProcesses()
	fmt.Println("Station stopped.")
	return nil
}

// RunRestart restarts the station.
func RunRestart() error {
	_ = RunStop()
	return RunStart(false)
}

// RunLogs displays recent Aircoda log entries.
func RunLogs(follow bool) error {
	logPath := config.ResolveLogPath()
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No log file found yet.")
			return nil
		}
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	maxLines := 30
	startIdx := 0
	if len(lines) > maxLines {
		startIdx = len(lines) - maxLines
	}

	fmt.Printf("Aircoda Log Tail (%s)\n", logPath)
	fmt.Println("────────────────────────────────────────")
	for i := startIdx; i < len(lines); i++ {
		fmt.Println(lines[i])
	}

	if !follow {
		return nil
	}

	// Tail follow loop
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
		fmt.Print(line)
	}

	return nil
}

// Subcommand constructs
func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start station live streaming",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunStart(startForeground)
		},
	}
	cmd.Flags().BoolVarP(&startForeground, "foreground", "f", false, "Run station in foreground mode")
	return cmd
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop station live streaming",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunStop()
		},
	}
}

func newRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart station live streaming",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunRestart()
		},
	}
}

func newLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Display recent Aircoda logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunLogs(logsFollow)
		},
	}
	cmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output continuously")
	return cmd
}

func newDaemonCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "daemon",
		Short:  "Internal daemon loop for live streaming",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDaemon()
		},
	}
}

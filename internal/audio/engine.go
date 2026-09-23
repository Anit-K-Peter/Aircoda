package audio

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"aircoda/internal/config"
	"aircoda/internal/library"
	"aircoda/internal/logger"
	"aircoda/internal/playlist"
)

// Engine manages the continuous FFmpeg audio encoding worker.
type Engine struct {
	Config        *config.Config
	Playlist      *playlist.Manager
	Icecast       *IcecastProcess
	Hub           *StreamHub
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	running       bool
	useEmbedded   bool
	masterCmd     *exec.Cmd
	encoderWriter io.WriteCloser
}

// BuildStreamURL generates public stream URL from configuration.
func BuildStreamURL(cfg *config.Config) string {
	mount := cfg.Streaming.Mount
	if !strings.HasPrefix(mount, "/") {
		mount = "/" + mount
	}
	return fmt.Sprintf("http://%s:%d%s", cfg.Streaming.Host, cfg.Streaming.Port, mount)
}

// NewEngine constructs an Audio Engine instance.
func NewEngine(cfg *config.Config, pl *playlist.Manager) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		Config:   cfg,
		Playlist: pl,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins continuous audio streaming to Icecast or embedded stream server.
func (e *Engine) Start() error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("station audio engine is already running")
	}
	e.running = true
	e.mu.Unlock()

	// 1. Check if port is available
	if err := CheckPortAvailable(e.Config.Streaming.Host, e.Config.Streaming.Port); err != nil {
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
		return err
	}

	// 2. Verify playlist has available tracks
	currentTrack, err := e.Playlist.Current()
	if err != nil {
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
		return fmt.Errorf("cannot start station: %w", err)
	}

	// 3. Detect Icecast binary or initialize Embedded StreamHub
	_, errIcecast := exec.LookPath("icecast")
	if errIcecast != nil {
		_, errIcecast = exec.LookPath("icecast2")
	}

	if errIcecast == nil {
		// Icecast binary exists -> start Icecast server process
		icecastProc, err := StartIcecast(e.Config)
		if err != nil {
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			return fmt.Errorf("failed to start Icecast: %w", err)
		}
		e.Icecast = icecastProc
		e.useEmbedded = false
	} else {
		// Icecast binary missing -> use embedded HTTP stream server fallback
		hub := NewStreamHub(e.Config.Streaming.Mount)
		if err := hub.Start(e.Config.Streaming.Host, e.Config.Streaming.Port); err != nil {
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			return fmt.Errorf("failed to start streaming server: %w", err)
		}
		e.Hub = hub
		e.useEmbedded = true
	}

	// 4. Start long-running Master FFmpeg Encoder
	if err := e.startMasterEncoder(); err != nil {
		e.Stop()
		return err
	}

	streamURL := BuildStreamURL(e.Config)
	logger.Info("Audio engine starting for station %s. Stream URL: %s", e.Config.Station.Name, streamURL)

	// Launch background streaming loop
	go e.runWorkerLoop(currentTrack)

	return nil
}

func (e *Engine) startMasterEncoder() error {
	ffmpegBin, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg binary not found in PATH")
	}

	bitrateArg := fmt.Sprintf("%dk", e.Config.Streaming.Bitrate)

	var masterCmd *exec.Cmd
	if !e.useEmbedded {
		mount := e.Config.Streaming.Mount
		if !strings.HasPrefix(mount, "/") {
			mount = "/" + mount
		}
		icecastDestination := fmt.Sprintf("icecast://source:%s@%s:%d%s",
			e.Config.Streaming.SourcePassword,
			e.Config.Streaming.Host,
			e.Config.Streaming.Port,
			mount,
		)

		masterCmd = exec.CommandContext(e.ctx, ffmpegBin,
			"-y",
			"-re",
			"-f", "s16le",
			"-ar", "44100",
			"-ac", "2",
			"-i", "pipe:0",
			"-content_type", "audio/mpeg",
			"-acodec", "libmp3lame",
			"-b:a", bitrateArg,
			"-ac", "2",
			"-ar", "44100",
			"-f", "mp3",
			icecastDestination,
		)
		masterCmd.Stdout = nil
		masterCmd.Stderr = nil
	} else {
		masterCmd = exec.CommandContext(e.ctx, ffmpegBin,
			"-y",
			"-re",
			"-f", "s16le",
			"-ar", "44100",
			"-ac", "2",
			"-i", "pipe:0",
			"-acodec", "libmp3lame",
			"-b:a", bitrateArg,
			"-ac", "2",
			"-ar", "44100",
			"-f", "mp3",
			"pipe:1",
		)
		masterCmd.Stdout = &StreamPipeWriter{Hub: e.Hub}
		masterCmd.Stderr = nil
	}

	encoderWriter, err := masterCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open master encoder stdin pipe: %w", err)
	}

	if err := masterCmd.Start(); err != nil {
		return fmt.Errorf("failed to start master FFmpeg streaming encoder: %w", err)
	}

	e.masterCmd = masterCmd
	e.encoderWriter = encoderWriter
	return nil
}

// StartBlocking starts the audio engine and blocks until context is cancelled or interrupted.
func (e *Engine) StartBlocking() error {
	if err := e.Start(); err != nil {
		return err
	}
	<-e.ctx.Done()
	return nil
}

// runWorkerLoop continuously decodes tracks sequentially to master encoder.
func (e *Engine) runWorkerLoop(initialTrack *library.Track) {
	currentTrack := initialTrack
	streamURL := BuildStreamURL(e.Config)
	ffmpegBin, _ := exec.LookPath("ffmpeg")
	startTime := time.Now()
	tracksPlayedCount := 0

	for {
		select {
		case <-e.ctx.Done():
			logger.Info("Audio engine loop shutting down...")
			return
		default:
		}

		// Verify Master Encoder is alive; auto-recover if process died
		if e.masterCmd == nil || e.masterCmd.Process == nil || !IsProcessRunning(e.masterCmd.Process.Pid) {
			logger.Info("Master encoder process not active. Auto-recovering master encoder...")
			if err := e.startMasterEncoder(); err != nil {
				logger.Error("Failed to auto-recover master encoder: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
		}

		if currentTrack == nil {
			nextTr, err := e.Playlist.Next()
			if err != nil {
				logger.Error("Audio engine: No available tracks in playlist: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			currentTrack = nextTr
		}

		// Check Station ID jingle interlacing drop
		if e.Config.Playlist.StationIDEnabled && tracksPlayedCount > 0 && e.Config.Playlist.StationIDInterval > 0 && tracksPlayedCount%e.Config.Playlist.StationIDInterval == 0 {
			jinglePath := resolveJinglePath(e.Config.Playlist.StationIDPath)
			if jinglePath != "" {
				logger.Info("Playing Station ID drop: %s", filepath.Base(jinglePath))
				jingleCmd := exec.CommandContext(e.ctx, ffmpegBin,
					"-y",
					"-i", jinglePath,
					"-f", "s16le",
					"-ar", "44100",
					"-ac", "2",
					"pipe:1",
				)
				jingleCmd.Stdout = e.encoderWriter
				_ = jingleCmd.Run()
			}
		}

		// Peek next track for status display
		nextTrackName := "Unknown"
		if qs, _, err := e.Playlist.GetOrInitQueue(); err == nil {
			if nextTrack, ok := qs.Current(); ok {
				nextTrackName = nextTrack.Filename
			}
		}

		// Save current server state
		icecastPID := 0
		if e.Icecast != nil {
			icecastPID = e.Icecast.PID()
		}
		masterPID := 0
		if e.masterCmd != nil && e.masterCmd.Process != nil {
			masterPID = e.masterCmd.Process.Pid
		}

		state := &ServerState{
			IsLive:       true,
			DaemonPID:    os.Getpid(),
			IcecastPID:   icecastPID,
			FFmpegPID:    masterPID,
			StartTime:    startTime,
			CurrentTrack: currentTrack.Filename,
			NextTrack:    nextTrackName,
			StreamURL:    streamURL,
			Bitrate:      e.Config.Streaming.Bitrate,
			Listeners:    "unavailable",
		}
		_ = SaveServerState(state)

		logger.Info("Playing track: %s", currentTrack.Filename)

		// Decode audio file to PCM and pipe continuously to master encoder
		decoderCmd := exec.CommandContext(e.ctx, ffmpegBin,
			"-y",
			"-i", currentTrack.Path,
			"-f", "s16le",
			"-ar", "44100",
			"-ac", "2",
			"pipe:1",
		)
		decoderCmd.Stdout = e.encoderWriter
		decoderCmd.Stderr = nil

		if err := decoderCmd.Start(); err != nil {
			logger.Error("Track decoder error for %s: %v", currentTrack.Filename, err)
			time.Sleep(1 * time.Second)
		} else {
			_ = decoderCmd.Wait()
		}

		tracksPlayedCount++

		// Advance to next track in playlist queue
		nextTrack, err := e.Playlist.Next()
		if err != nil {
			logger.Info("Playlist finished or reached end: %v", err)
			currentTrack = nil
		} else {
			currentTrack = nextTrack
		}
	}
}

func resolveJinglePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	if !info.IsDir() {
		return path
	}
	tracks, err := library.ScanDirectory(path, false)
	if err == nil && len(tracks) > 0 {
		return tracks[0].Path
	}
	return ""
}

// Stop cleanly terminates the streaming engine, hub, master encoder, and Icecast.
func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	e.mu.Unlock()

	e.cancel()

	if e.encoderWriter != nil {
		_ = e.encoderWriter.Close()
	}

	if e.masterCmd != nil && e.masterCmd.Process != nil {
		_ = e.masterCmd.Process.Kill()
	}

	if e.Icecast != nil {
		e.Icecast.Stop()
	}

	if e.Hub != nil {
		e.Hub.Stop()
	}

	ClearServerState()
	logger.Info("Station audio engine stopped successfully.")
}

// KillExistingProcesses reads saved server state and kills orphaned Icecast/FFmpeg processes.
func KillExistingProcesses() {
	state, err := LoadServerState()
	if err == nil && state != nil {
		if state.DaemonPID > 0 && state.DaemonPID != os.Getpid() && IsProcessRunning(state.DaemonPID) {
			if proc, err := os.FindProcess(state.DaemonPID); err == nil {
				_ = proc.Kill()
			}
		}
		if state.FFmpegPID > 0 && IsProcessRunning(state.FFmpegPID) {
			if proc, err := os.FindProcess(state.FFmpegPID); err == nil {
				_ = proc.Kill()
			}
		}
		if state.IcecastPID > 0 && IsProcessRunning(state.IcecastPID) {
			if proc, err := os.FindProcess(state.IcecastPID); err == nil {
				_ = proc.Kill()
			}
		}
	}
	ClearServerState()

	// Terminate any leftover orphaned icecast or radio daemon processes
	_ = exec.Command("pkill", "-9", "-f", "icecast.xml").Run()
	if os.Getpid() > 0 {
		// Kill radio daemon processes that are NOT current process
		_ = exec.Command("pkill", "-9", "-f", "radio daemon").Run()
	}
	time.Sleep(300 * time.Millisecond)
}

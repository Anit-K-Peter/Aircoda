package cli

import (
	"fmt"
	"strings"

	"aircoda/internal/audio"
	"aircoda/internal/config"
	"aircoda/internal/playlist"
	"aircoda/internal/public"
	"aircoda/internal/service"
	"aircoda/internal/system"
	"aircoda/internal/tunnel"
)

// RunStatus displays current station, OS platform, and background service status.
func RunStatus() error {
	cfg, _, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("Aircoda is not configured yet. Run 'radio setup' to initialize")
		}
		return fmt.Errorf("failed to load status: %w", err)
	}

	state, stateErr := audio.LoadServerState()
	isLive := false
	pidStr := "-"
	icecastState := "● STOPPED"
	ffmpegState := "● STOPPED"
	currentTrack := "None"
	nextTrack := "None"
	localStreamURL := audio.BuildStreamURL(cfg)

	if stateErr == nil && state != nil && state.IsLive {
		if audio.IsProcessRunning(state.DaemonPID) || audio.IsProcessRunning(state.IcecastPID) || audio.IsProcessRunning(state.FFmpegPID) {
			isLive = true
			if state.DaemonPID > 0 {
				pidStr = fmt.Sprintf("%d", state.DaemonPID)
			}
			if state.IcecastPID > 0 && audio.IsProcessRunning(state.IcecastPID) {
				icecastState = "● RUNNING"
			}
			if state.FFmpegPID > 0 && audio.IsProcessRunning(state.FFmpegPID) {
				ffmpegState = "● RUNNING"
			}
			if state.CurrentTrack != "" {
				currentTrack = state.CurrentTrack
			}
			if state.NextTrack != "" {
				nextTrack = state.NextTrack
			}
			if state.StreamURL != "" {
				localStreamURL = state.StreamURL
			}
		}
	}

	// Service Status
	svcStateStr := "● STOPPED"
	if svcMgr, err := service.NewServiceManager(); err == nil {
		if svcMgr.Status() == "RUNNING" {
			svcStateStr = "● RUNNING"
		}
	}
	if isLive {
		svcStateStr = "● RUNNING"
	}

	mgr := playlist.NewManager(cfg)
	_, tracks, scanErr := mgr.GetOrInitQueue()
	trackCountStr := "0 tracks"
	if scanErr == nil {
		trackCountStr = fmt.Sprintf("%d tracks", len(tracks))
	} else {
		trackCountStr = "Directory missing"
	}

	modeLabel := "Shuffle"
	if strings.ToLower(cfg.Playlist.Mode) == "sequential" {
		modeLabel = "Sequential"
	}

	repeatLabel := "No"
	if cfg.Playlist.Repeat {
		repeatLabel = "Yes"
	}

	// Public access status
	pubModeUpper := strings.ToUpper(cfg.Public.Mode)
	if pubModeUpper == "" {
		pubModeUpper = "LOCAL"
	}
	pubAccessStr := "● LOCAL"
	if pubModeUpper == "DIRECT" {
		pubAccessStr = "● DIRECT"
	} else if pubModeUpper == "TUNNEL" {
		if tunSt, _ := tunnel.GetStatus(cfg); tunSt != nil && tunSt.IsRunning {
			pubAccessStr = "● CONNECTED"
		} else {
			pubAccessStr = "● TUNNEL (OFFLINE)"
		}
	}

	publicURL := public.BuildPublicStreamURL(cfg)

	fmt.Println("╭────────────────────────────────────────────╮")
	fmt.Println("│                  AIRCODA                   │")
	fmt.Printf("│                  %-25s │\n", system.GetVersion())
	fmt.Println("╰────────────────────────────────────────────╯")
	fmt.Println()
	fmt.Printf("Station      %s\n", cfg.Station.Name)
	fmt.Printf("Service      %s\n", svcStateStr)
	fmt.Printf("Platform     %s (%s)\n", system.GetPlatformName(), system.GetArch())
	fmt.Printf("PID          %s\n", pidStr)
	fmt.Println()
	fmt.Printf("Playlist     %s\n", trackCountStr)
	fmt.Printf("Mode         %s\n", modeLabel)
	fmt.Printf("Repeat       %s\n", repeatLabel)
	fmt.Println()
	fmt.Printf("Now Playing  %s\n", currentTrack)
	fmt.Printf("Next         %s\n", nextTrack)
	fmt.Println()
	fmt.Printf("Icecast      %s\n", icecastState)
	fmt.Printf("FFmpeg       %s\n", ffmpegState)
	fmt.Println()
	fmt.Printf("Local Stream %s\n", localStreamURL)
	fmt.Printf("Public Access %s\n", pubAccessStr)
	if pubModeUpper != "LOCAL" {
		fmt.Printf("Public Stream %s\n", publicURL)
	}

	return nil
}

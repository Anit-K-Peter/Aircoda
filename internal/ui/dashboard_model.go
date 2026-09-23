package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"aircoda/internal/audio"
	"aircoda/internal/config"
	"aircoda/internal/deps"
	"aircoda/internal/library"
	"aircoda/internal/picker"
	"aircoda/internal/playlist"
	"aircoda/internal/public"
	"aircoda/internal/system"
	"aircoda/internal/tunnel"
)

type DashboardView int

const (
	ViewMainDashboard DashboardView = iota
	ViewNowPlaying
	ViewPlaylist
	ViewRadioMenu
	ViewStationMenu
	ViewSystemMenu
	ViewPublicMenu
	ViewLogs
	ViewExitPrompt
	ViewEditFolderPicker
	ViewEditFolderPathInput
	ViewFolderScanConfirm
	ViewNotice
	ViewYouTubeInput
	ViewProgress
	ViewDiagnostics
	ViewServiceStatus
	ViewResetConfirm
)

type actionCompleteMsg struct {
	notice string
}

type DashboardModel struct {
	Cfg          *config.Config
	CurrentView  DashboardView
	FolderPicker picker.FolderPicker

	// Active status
	IsLive       bool
	CurrentTrack string
	NextTrack    string
	TrackCount   int
	Tracks       []library.Track
	State        *audio.ServerState
	StatusText   string

	// Menu cursor positions
	MainMenuCursor   int
	SubMenuCursor    int
	NowPlayingCursor int
	PlaylistCursor   int
	SystemMenuCursor int
	ExitCursor       int
	ResetCursor      int

	// Inputs & Buffers
	FolderScanCount int
	NewFolder       string
	ManualPathBuf   string
	YouTubeURLBuf   string
	NoticeMessage   string
	ProgressTitle   string
	ProgressMessage string

	Width  int
	Height int

	// Log lines
	LogLines []string

	// Flag to exit program
	ShouldQuit bool
}

func NewDashboardModel(cfg *config.Config) *DashboardModel {
	m := &DashboardModel{
		Cfg:          cfg,
		CurrentView:  ViewMainDashboard,
		FolderPicker: picker.NewSystemFolderPicker(),
		ExitCursor:   1, // Default to "No" for safe exit
		ResetCursor:  0, // Default to "No, Cancel"
	}
	m.RefreshState()
	return m
}

func (m *DashboardModel) RefreshState() {
	if cfg, _, err := config.Load(); err == nil {
		m.Cfg = cfg
	}

	mgr := playlist.NewManager(m.Cfg)
	_, tracks, err := mgr.GetOrInitQueue()
	if err == nil {
		m.Tracks = tracks
		m.TrackCount = len(tracks)
	}

	state, err := audio.LoadServerState()
	if err == nil && state.IsLive {
		if audio.IsProcessRunning(state.DaemonPID) || audio.IsProcessRunning(state.IcecastPID) || audio.IsProcessRunning(state.FFmpegPID) {
			m.IsLive = true
			m.State = state
			m.CurrentTrack = state.CurrentTrack
			m.NextTrack = state.NextTrack
			m.StatusText = "● LIVE"
		} else {
			m.IsLive = false
			m.StatusText = "● STOPPED"
		}
	} else {
		m.IsLive = false
		if err != nil || m.TrackCount == 0 {
			m.StatusText = "● NO AUDIO FILES"
		} else {
			m.StatusText = "● STOPPED"
		}
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case actionCompleteMsg:
		m.RefreshState()
		mgr := playlist.NewManager(m.Cfg)
		_, tracks, err := mgr.GetOrInitQueue()
		if err == nil {
			m.Tracks = tracks
			m.TrackCount = len(tracks)
		}
		m.NoticeMessage = msg.notice
		m.CurrentView = ViewNotice
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.ShouldQuit = true
			return m, tea.Quit
		}

		switch m.CurrentView {
		case ViewMainDashboard:
			return m.updateMainDashboard(msg)
		case ViewNowPlaying:
			return m.updateNowPlaying(msg)
		case ViewPlaylist:
			return m.updatePlaylist(msg)
		case ViewRadioMenu:
			return m.updateRadioMenu(msg)
		case ViewStationMenu:
			return m.updateStationMenu(msg)
		case ViewSystemMenu:
			return m.updateSystemMenu(msg)
		case ViewPublicMenu:
			return m.updatePublicMenu(msg)
		case ViewLogs:
			return m.updateLogs(msg)
		case ViewExitPrompt:
			return m.updateExitPrompt(msg)
		case ViewEditFolderPicker:
			return m.updateEditFolderPicker(msg)
		case ViewEditFolderPathInput:
			return m.updateEditFolderPathInput(msg)
		case ViewFolderScanConfirm:
			return m.updateFolderScanConfirm(msg)
		case ViewYouTubeInput:
			return m.updateYouTubeInput(msg)
		case ViewNotice:
			return m.updateNotice(msg)
		case ViewDiagnostics:
			return m.updateDiagnostics(msg)
		case ViewServiceStatus:
			return m.updateServiceStatus(msg)
		case ViewResetConfirm:
			return m.updateResetConfirm(msg)
		case ViewProgress:
			return m, nil
		}
	}

	return m, nil
}

func (m *DashboardModel) updateMainDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.MainMenuCursor > 0 {
			m.MainMenuCursor--
		}
	} else if isDownKey(msg) {
		if m.MainMenuCursor < 4 {
			m.MainMenuCursor++
		}
	} else if isEnterKey(msg) {
		m.selectMainMenuItem(m.MainMenuCursor)
	} else if msg.String() >= "1" && msg.String() <= "5" {
		idx := int(msg.String()[0] - '1')
		m.MainMenuCursor = idx
		m.selectMainMenuItem(idx)
	} else if msg.String() == "q" {
		m.ExitCursor = 1
		m.CurrentView = ViewExitPrompt
	}
	return m, nil
}

func (m *DashboardModel) selectMainMenuItem(index int) {
	m.SubMenuCursor = 0
	m.NowPlayingCursor = 0
	m.PlaylistCursor = 0
	m.SystemMenuCursor = 0
	switch index {
	case 0: // Now Playing
		m.RefreshState()
		m.CurrentView = ViewNowPlaying
	case 1: // Playlist
		m.RefreshState()
		m.CurrentView = ViewPlaylist
	case 2: // Radio
		m.RefreshState()
		m.CurrentView = ViewRadioMenu
	case 3: // System
		m.RefreshState()
		m.CurrentView = ViewSystemMenu
	case 4: // Exit
		m.ExitCursor = 1 // Default to "No"
		m.CurrentView = ViewExitPrompt
	}
}

func (m *DashboardModel) updateNowPlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.NowPlayingCursor > 0 {
			m.NowPlayingCursor--
		}
	} else if isDownKey(msg) {
		if m.NowPlayingCursor < 2 {
			m.NowPlayingCursor++
		}
	} else if msg.String() >= "1" && msg.String() <= "3" {
		m.NowPlayingCursor = int(msg.String()[0] - '1')
		return m.executeNowPlayingAction()
	} else if isEnterKey(msg) || msg.String() == "s" {
		return m.executeNowPlayingAction()
	} else if isEscKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewMainDashboard
	} else if msg.String() == "r" {
		m.RefreshState()
	}
	return m, nil
}

func (m *DashboardModel) executeNowPlayingAction() (tea.Model, tea.Cmd) {
	if m.NowPlayingCursor == 0 {
		if m.IsLive {
			m.ProgressTitle = "STOPPING RADIO STATION"
			m.ProgressMessage = "[1/2] Terminating audio engine process...\n[2/2] Clearing runtime state..."
			m.CurrentView = ViewProgress
			return m, stopRadioCmd()
		} else {
			m.ProgressTitle = "STARTING RADIO STATION"
			m.ProgressMessage = "[1/2] Terminating old processes...\n[2/2] Launching background streaming daemon..."
			m.CurrentView = ViewProgress
			return m, startRadioCmd()
		}
	} else if m.NowPlayingCursor == 1 {
		m.RefreshState()
	} else if m.NowPlayingCursor == 2 {
		m.CurrentView = ViewMainDashboard
	}
	return m, nil
}

func (m *DashboardModel) updatePlaylist(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.PlaylistCursor > 0 {
			m.PlaylistCursor--
		}
	} else if isDownKey(msg) {
		if m.PlaylistCursor < 3 {
			m.PlaylistCursor++
		}
	} else if isEnterKey(msg) {
		switch m.PlaylistCursor {
		case 0: // Add YouTube Audio Link
			m.YouTubeURLBuf = ""
			m.CurrentView = ViewYouTubeInput
		case 1: // Change Music Folder
			folder, err := m.FolderPicker.PickFolder()
			if err != nil || folder == "" {
				m.ManualPathBuf = m.Cfg.Audio.Directory
				m.CurrentView = ViewEditFolderPathInput
				return m, nil
			}
			m.NewFolder = folder
			m.scanNewFolder()
		case 2: // Rescan Library
			mgr := playlist.NewManager(m.Cfg)
			_, tracks, err := mgr.Rescan()
			if err == nil {
				m.Tracks = tracks
				m.TrackCount = len(tracks)
			}
			m.NoticeMessage = fmt.Sprintf("✓ Playlist queue rescanned. Total tracks: %d", m.TrackCount)
			m.CurrentView = ViewNotice
		case 3: // Back
			m.CurrentView = ViewMainDashboard
		}
	} else if isEscKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewMainDashboard
	} else if msg.String() == "r" {
		mgr := playlist.NewManager(m.Cfg)
		_, tracks, _ := mgr.Rescan()
		m.Tracks = tracks
		m.TrackCount = len(tracks)
	}
	return m, nil
}

func (m *DashboardModel) updateYouTubeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		url := strings.TrimSpace(m.YouTubeURLBuf)
		if url == "" {
			return m, nil
		}
		m.ProgressTitle = "DOWNLOADING YOUTUBE AUDIO"
		m.ProgressMessage = fmt.Sprintf("Downloading audio track using yt-dlp to:\n%s\n\nPlease wait...", m.Cfg.Audio.Directory)
		m.CurrentView = ViewProgress
		return m, downloadYouTubeCmd(url, m.Cfg.Audio.Directory)
	}

	if isBackspaceKey(msg) {
		if len(m.YouTubeURLBuf) > 0 {
			m.YouTubeURLBuf = m.YouTubeURLBuf[:len(m.YouTubeURLBuf)-1]
		}
		return m, nil
	}

	if isEscKey(msg) {
		m.CurrentView = ViewPlaylist
		return m, nil
	}

	if len(msg.String()) == 1 {
		m.YouTubeURLBuf += msg.String()
	}

	return m, nil
}

func (m *DashboardModel) updateRadioMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.SubMenuCursor > 0 {
			m.SubMenuCursor--
		}
	} else if isDownKey(msg) {
		if m.SubMenuCursor < 3 {
			m.SubMenuCursor++
		}
	} else if isEnterKey(msg) {
		switch m.SubMenuCursor {
		case 0: // Start Radio
			m.ProgressTitle = "STARTING RADIO STATION"
			m.ProgressMessage = "[1/2] Terminating old processes...\n[2/2] Launching background streaming daemon..."
			m.CurrentView = ViewProgress
			return m, startRadioCmd()

		case 1: // Stop Radio
			m.ProgressTitle = "STOPPING RADIO STATION"
			m.ProgressMessage = "[1/2] Terminating audio engine process...\n[2/2] Clearing runtime state..."
			m.CurrentView = ViewProgress
			return m, stopRadioCmd()

		case 2: // Restart Radio
			m.ProgressTitle = "RESTARTING RADIO STATION"
			m.ProgressMessage = "[1/2] Stopping existing radio processes...\n[2/2] Relaunching audio engine daemon..."
			m.CurrentView = ViewProgress
			return m, restartRadioCmd()

		case 3: // Status
			m.RefreshState()
		}
	} else if isEscKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewMainDashboard
	}
	return m, nil
}

func (m *DashboardModel) updateStationMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.SubMenuCursor > 0 {
			m.SubMenuCursor--
		}
	} else if isDownKey(msg) {
		if m.SubMenuCursor < 6 {
			m.SubMenuCursor++
		}
	} else if isEnterKey(msg) {
		switch m.SubMenuCursor {
		case 0, 1: // Station Name / ID
			m.NoticeMessage = "Station Name/ID can be modified in your station configuration file or setup wizard."
			m.CurrentView = ViewNotice

		case 2: // Music Folder
			folder, err := m.FolderPicker.PickFolder()
			if err != nil || folder == "" {
				m.ManualPathBuf = m.Cfg.Audio.Directory
				m.CurrentView = ViewEditFolderPathInput
				return m, nil
			}
			m.NewFolder = folder
			m.scanNewFolder()

		case 3: // Playback (Toggle Shuffle / Sequential)
			if m.Cfg.Playlist.Mode == "shuffle" {
				m.Cfg.Playlist.Mode = "sequential"
			} else {
				m.Cfg.Playlist.Mode = "shuffle"
			}
			_, _ = config.Save(m.Cfg, "")
			m.RefreshState()

		case 4: // Station ID Jingle Drops (Toggle On/Off)
			m.Cfg.Playlist.StationIDEnabled = !m.Cfg.Playlist.StationIDEnabled
			_, _ = config.Save(m.Cfg, "")
			m.RefreshState()

		case 5: // Daypart Schedule Mode (Toggle On/Off)
			m.Cfg.Schedule.Enabled = !m.Cfg.Schedule.Enabled
			_, _ = config.Save(m.Cfg, "")
			m.RefreshState()

		case 6: // Startup (Toggle Start on boot)
			m.Cfg.System.StartOnBoot = !m.Cfg.System.StartOnBoot
			_, _ = config.Save(m.Cfg, "")
			m.RefreshState()
		}
	} else if isEscKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewMainDashboard
	}
	return m, nil
}

func (m *DashboardModel) scanNewFolder() {
	tracks, err := library.ScanDirectory(m.NewFolder, true)
	if err != nil {
		m.NoticeMessage = fmt.Sprintf("Error scanning folder: %v", err)
		m.CurrentView = ViewNotice
		return
	}
	m.FolderScanCount = len(tracks)
	m.CurrentView = ViewFolderScanConfirm
}

func (m *DashboardModel) updateEditFolderPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		folder, err := m.FolderPicker.PickFolder()
		if err != nil || folder == "" {
			m.ManualPathBuf = m.Cfg.Audio.Directory
			m.CurrentView = ViewEditFolderPathInput
			return m, nil
		}
		m.NewFolder = folder
		m.scanNewFolder()
	} else if isEscKey(msg) {
		m.CurrentView = ViewStationMenu
	}
	return m, nil
}

func (m *DashboardModel) updateEditFolderPathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		val := strings.TrimSpace(m.ManualPathBuf)
		if val != "" {
			m.NewFolder = val
			m.scanNewFolder()
		}
		return m, nil
	}

	if isBackspaceKey(msg) {
		if len(m.ManualPathBuf) > 0 {
			m.ManualPathBuf = m.ManualPathBuf[:len(m.ManualPathBuf)-1]
		}
		return m, nil
	}

	if isEscKey(msg) {
		m.CurrentView = ViewStationMenu
		return m, nil
	}

	if len(msg.String()) == 1 {
		m.ManualPathBuf += msg.String()
	}

	return m, nil
}

func (m *DashboardModel) updateFolderScanConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		m.Cfg.Audio.Directory = m.NewFolder
		_, _ = config.Save(m.Cfg, "")
		mgr := playlist.NewManager(m.Cfg)
		_, _, _ = mgr.Rescan()
		m.RefreshState()
		m.NoticeMessage = "Station music folder updated successfully."
		m.CurrentView = ViewNotice
	} else if isEscKey(msg) {
		m.CurrentView = ViewStationMenu
	}
	return m, nil
}

func (m *DashboardModel) updateSystemMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.SystemMenuCursor > 0 {
			m.SystemMenuCursor--
		}
	} else if isDownKey(msg) {
		if m.SystemMenuCursor < 8 {
			m.SystemMenuCursor++
		}
	} else if isEnterKey(msg) {
		switch m.SystemMenuCursor {
		case 0: // Station Settings
			m.SubMenuCursor = 0
			m.CurrentView = ViewStationMenu
		case 1: // Public Access Settings
			m.SubMenuCursor = 0
			m.CurrentView = ViewPublicMenu
		case 2: // View Recent Logs
			m.loadLogs()
			m.CurrentView = ViewLogs
		case 3: // System Status & Live Diagnostics
			m.CurrentView = ViewDiagnostics
		case 4: // Self-Repair Dependencies
			m.ProgressTitle = "SELF-REPAIRING DEPENDENCIES"
			m.ProgressMessage = "Scanning and auto-repairing missing binaries (yt-dlp, cloudflared, ffmpeg, icecast)...\n\nPlease wait..."
			m.CurrentView = ViewProgress
			return m, selfRepairCmd()
		case 5: // Background Service Status
			m.CurrentView = ViewServiceStatus
		case 6: // Station Configuration Info
			m.NoticeMessage = fmt.Sprintf("Station Name: %s\nStation ID: %s\nConfig Path: %s\nAudio Directory: %s\nStream Port: %d (Format: %s, Bitrate: %dkbps)",
				m.Cfg.Station.Name,
				m.Cfg.Station.ID,
				config.ResolveConfigPath(),
				m.Cfg.Audio.Directory,
				m.Cfg.Streaming.Port,
				m.Cfg.Streaming.Format,
				m.Cfg.Streaming.Bitrate,
			)
			m.CurrentView = ViewNotice
		case 7: // Reset Radio
			m.ResetCursor = 0 // Default to Cancel
			m.CurrentView = ViewResetConfirm
		case 8: // Back
			m.CurrentView = ViewMainDashboard
		}
	} else if isEscKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewMainDashboard
	}
	return m, nil
}

func (m *DashboardModel) updateDiagnostics(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEscKey(msg) || isEnterKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewSystemMenu
	}
	return m, nil
}

func (m *DashboardModel) updateServiceStatus(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEscKey(msg) || isEnterKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewSystemMenu
	}
	return m, nil
}

func (m *DashboardModel) updateResetConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) || isDownKey(msg) {
		if m.ResetCursor == 0 {
			m.ResetCursor = 1
		} else {
			m.ResetCursor = 0
		}
	} else if isEnterKey(msg) {
		if m.ResetCursor == 1 { // Yes, Reset Everything
			audio.KillExistingProcesses()
			audio.ClearServerState()
			_ = os.Remove(config.ResolveConfigPath())
			m.RefreshState()
			m.NoticeMessage = "✓ Aircoda radio reset completed successfully.\nRun 'radio setup' to re-initialize your radio station."
			m.CurrentView = ViewNotice
			return m, nil
		}
		m.CurrentView = ViewSystemMenu // Cancel
	} else if isEscKey(msg) {
		m.CurrentView = ViewSystemMenu
	}
	return m, nil
}

func (m *DashboardModel) updatePublicMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.SubMenuCursor > 0 {
			m.SubMenuCursor--
		}
	} else if isDownKey(msg) {
		if m.SubMenuCursor < 3 {
			m.SubMenuCursor++
		}
	} else if isEnterKey(msg) {
		switch m.SubMenuCursor {
		case 0: // One-Click Public Broadcast (Cloudflare Tunnel)
			m.ProgressTitle = "SETTING UP PUBLIC BROADCAST"
			m.ProgressMessage = "[1/3] Ensuring radio engine is active...\n[2/3] Verifying/Downloading cloudflared binary...\n[3/3] Establishing Cloudflare Tunnel..."
			m.CurrentView = ViewProgress
			return m, startCloudflareCmd(m.Cfg)

		case 1: // Access Status
			pubURL := public.BuildPublicStreamURL(m.Cfg)
			m.NoticeMessage = fmt.Sprintf("Public Mode: %s\nStream URL: %s", strings.ToUpper(m.Cfg.Public.Mode), pubURL)
			m.CurrentView = ViewNotice

		case 2: // Direct Access
			m.Cfg.Public.Mode = "direct"
			_, _ = config.Save(m.Cfg, "")
			pubURL := public.BuildPublicStreamURL(m.Cfg)
			m.NoticeMessage = fmt.Sprintf("Public mode set to Direct Access.\nStream URL: %s", pubURL)
			m.CurrentView = ViewNotice

		case 3: // Cloudflare Tunnel Settings
			m.Cfg.Public.Mode = "tunnel"
			_, _ = config.Save(m.Cfg, "")
			m.NoticeMessage = "Public mode set to Cloudflare Tunnel. Run 'radio tunnel setup' for custom domain settings."
			m.CurrentView = ViewNotice
		}
	} else if isEscKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewMainDashboard
	}
	return m, nil
}

func (m *DashboardModel) loadLogs() {
	logPath := config.ResolveLogPath()
	file, err := os.Open(logPath)
	if err != nil {
		m.LogLines = []string{"No logs found at " + logPath}
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}
	m.LogLines = lines
}

func (m *DashboardModel) updateLogs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEscKey(msg) || msg.String() == "q" || isEnterKey(msg) {
		m.CurrentView = ViewMainDashboard
	} else if msg.String() == "r" {
		m.loadLogs()
	}
	return m, nil
}

func (m *DashboardModel) updateNotice(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) || isEscKey(msg) || msg.String() == "q" {
		m.CurrentView = ViewMainDashboard
	}
	return m, nil
}

func (m *DashboardModel) updateExitPrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) || isDownKey(msg) {
		if m.ExitCursor == 0 {
			m.ExitCursor = 1
		} else {
			m.ExitCursor = 0
		}
	} else if isEnterKey(msg) {
		if m.ExitCursor == 0 { // Yes, exit TUI
			m.ShouldQuit = true
			return m, tea.Quit
		}
		m.CurrentView = ViewMainDashboard // No, remain in dashboard
	} else if isEscKey(msg) {
		m.CurrentView = ViewMainDashboard
	}
	return m, nil
}

func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%02dm %02ds", m, s)
	}
	return fmt.Sprintf("%02ds", s)
}

func (m *DashboardModel) View() string {
	var b strings.Builder

	// Header Box with ASCII Banner
	b.WriteString(RenderAsciiBanner(system.GetVersion()) + "\n\n")

	switch m.CurrentView {
	case ViewMainDashboard:
		b.WriteString(StyleTitle.Render(m.Cfg.Station.Name) + "\n\n")

		statusBadge := StyleBadgeOffline.Render("● OFFLINE")
		uptimeStr := "Not live"
		if m.IsLive {
			statusBadge = StyleBadgeLive.Render("● LIVE")
			if m.State != nil && !m.State.StartTime.IsZero() {
				uptimeStr = FormatDuration(time.Since(m.State.StartTime))
			}
		}

		nowPlaying := m.CurrentTrack
		if nowPlaying == "" {
			nowPlaying = "None"
		}
		nextTrack := m.NextTrack
		if nextTrack == "" {
			nextTrack = "None"
		}

		b.WriteString(fmt.Sprintf("%-14s %s\n", "Status", statusBadge))
		if m.IsLive {
			b.WriteString(fmt.Sprintf("%-14s %s\n", "Streaming for", uptimeStr))
		}
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Now Playing", nowPlaying))
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Next", nextTrack))
		b.WriteString(fmt.Sprintf("%-14s %d\n", "Tracks", m.TrackCount))
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Platform", system.GetPlatformName()))

		b.WriteString("\n──────────────────────────────────────────────\n\n")

		menuItems := []string{
			"Now Playing",
			"Playlist",
			"Radio",
			"System",
			"Exit",
		}

		for i, item := range menuItems {
			if i == m.MainMenuCursor {
				b.WriteString("> " + StyleSelectedOption.Render(item) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(item) + "\n")
			}
		}

		b.WriteString("\n──────────────────────────────────────────────\n")
		b.WriteString(StyleHelp.Render("↑↓ or j/k Navigate   Enter/1-5 Select   Esc Back"))

	case ViewNowPlaying:
		b.WriteString(StyleTitle.Render("NOW PLAYING DETAILS & BROADCAST CONTROL") + "\n\n")

		toggleAction := "[START RADIO STREAM]"
		if m.IsLive {
			toggleAction = "[STOP RADIO STREAM]"
		}

		nowPlayingOpts := []string{
			toggleAction,
			"Refresh Status",
			"Back to Main Dashboard",
		}

		for i, opt := range nowPlayingOpts {
			if i == m.NowPlayingCursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n──────────────────────────────────────────────\n\n")

		statusBadge := StyleBadgeOffline.Render("● OFFLINE")
		uptimeStr := "Not live"
		if m.IsLive {
			statusBadge = StyleBadgeLive.Render("● LIVE")
			if m.State != nil && !m.State.StartTime.IsZero() {
				uptimeStr = FormatDuration(time.Since(m.State.StartTime))
			}
		}
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Status", statusBadge))
		if m.IsLive {
			b.WriteString(fmt.Sprintf("%-14s %s\n", "Streaming for", uptimeStr))
		}
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Station", m.Cfg.Station.Name))
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Track", m.CurrentTrack))
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Next Track", m.NextTrack))
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Stream URL", public.BuildPublicStreamURL(m.Cfg)))
		b.WriteString(fmt.Sprintf("%-14s %d kbps\n", "Bitrate", m.Cfg.Streaming.Bitrate))
		b.WriteString(fmt.Sprintf("%-14s %s\n", "Format", m.Cfg.Streaming.Format))
		b.WriteString("\n" + StyleHelp.Render("Press S or Enter to toggle Start/Stop, R to refresh, Esc Back."))

	case ViewPlaylist:
		b.WriteString(StyleTitle.Render("STATION PLAYLIST & MEDIA MANAGEMENT") + "\n\n")

		playlistOpts := []string{
			"Add Audio via YouTube Link",
			"Select / Change Music Folder",
			"Rescan Library Queue",
			"Back to Main Dashboard",
		}

		for i, opt := range playlistOpts {
			if i == m.PlaylistCursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n──────────────────────────────────────────────\n\n")

		b.WriteString(fmt.Sprintf("Audio Directory: %s\n", m.Cfg.Audio.Directory))
		b.WriteString(fmt.Sprintf("Total Tracks   : %d\n\n", m.TrackCount))
		b.WriteString("Upcoming Queue:\n")
		limit := 8
		if len(m.Tracks) < limit {
			limit = len(m.Tracks)
		}
		for i := 0; i < limit; i++ {
			b.WriteString(fmt.Sprintf("  %2d. %s\n", i+1, m.Tracks[i].Filename))
		}
		b.WriteString("\n" + StyleHelp.Render("↑↓ Navigate, Enter Select, R Rescan, Esc Back."))

	case ViewYouTubeInput:
		b.WriteString(StyleTitle.Render("ADD AUDIO VIA YOUTUBE LINK") + "\n\n")
		b.WriteString("Enter YouTube Video URL:\n")
		b.WriteString("> " + m.YouTubeURLBuf + "\n\n")
		b.WriteString("Note: Requires yt-dlp installed. Audio will be saved to:\n")
		b.WriteString(m.Cfg.Audio.Directory + "\n\n")
		b.WriteString(StyleHelp.Render("Press Enter to download, Esc to cancel."))

	case ViewRadioMenu:
		b.WriteString(StyleTitle.Render("RADIO MANAGEMENT") + "\n\n")
		opts := []string{
			"Start Radio",
			"Stop Radio",
			"Restart Radio",
			"Status",
		}
		for i, opt := range opts {
			if i == m.SubMenuCursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("↑↓ Navigate, Enter Select, Esc Back"))

	case ViewStationMenu:
		b.WriteString(StyleTitle.Render("STATION SETTINGS") + "\n\n")
		stationIDStatus := "Disabled"
		if m.Cfg.Playlist.StationIDEnabled {
			stationIDStatus = fmt.Sprintf("Every %d tracks", m.Cfg.Playlist.StationIDInterval)
		}
		scheduleStatus := "Disabled"
		if m.Cfg.Schedule.Enabled {
			scheduleStatus = "Enabled (Time-based Dayparts)"
		}

		opts := []string{
			fmt.Sprintf("Station Name (%s)", m.Cfg.Station.Name),
			fmt.Sprintf("Station ID (%s)", m.Cfg.Station.ID),
			fmt.Sprintf("Music Folder (%s)", filepath.Base(m.Cfg.Audio.Directory)),
			fmt.Sprintf("Playback (%s, Repeat: %v)", m.Cfg.Playlist.Mode, m.Cfg.Playlist.Repeat),
			fmt.Sprintf("Station ID Drops (%s)", stationIDStatus),
			fmt.Sprintf("Daypart Schedule (%s)", scheduleStatus),
			fmt.Sprintf("Startup (Start on boot: %v)", m.Cfg.System.StartOnBoot),
		}
		for i, opt := range opts {
			if i == m.SubMenuCursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("↑↓ Navigate, Enter Select/Toggle, Esc Back"))

	case ViewSystemMenu:
		b.WriteString(StyleTitle.Render("SYSTEM MANAGEMENT & DIAGNOSTICS") + "\n\n")
		opts := []string{
			"Station Settings",
			"Public Access Settings",
			"View Recent Logs",
			"Live System Diagnostics",
			"Background Service Status",
			"Station Configuration Info",
			"Reset Radio (Factory Reset)",
			"Back to Main Menu",
		}
		for i, opt := range opts {
			if i == m.SystemMenuCursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("↑↓ Navigate, Enter Select, Esc Back"))

	case ViewDiagnostics:
		b.WriteString(StyleTitle.Render("LIVE SYSTEM DIAGNOSTICS REPORT") + "\n\n")
		b.WriteString(fmt.Sprintf("%-18s %s\n", "Platform", system.GetPlatformName()))
		b.WriteString(fmt.Sprintf("%-18s %s\n", "Go Runtime", system.GetGoVersion()))

		ffmpegStatus := StyleCheck.Render("OK")
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			ffmpegStatus = StyleCross.Render("MISSING")
		}
		b.WriteString(fmt.Sprintf("%-18s %s\n", "FFmpeg", ffmpegStatus))

		ffprobeStatus := StyleCheck.Render("OK")
		if _, err := exec.LookPath("ffprobe"); err != nil {
			ffprobeStatus = StyleCross.Render("MISSING")
		}
		b.WriteString(fmt.Sprintf("%-18s %s\n", "FFprobe", ffprobeStatus))

		icecastStatus := StyleCheck.Render("OK")
		if _, err := exec.LookPath("icecast"); err != nil {
			if _, err2 := exec.LookPath("icecast2"); err2 != nil {
				icecastStatus = StyleBadgeWarning.Render("Embedded Stream Server")
			}
		}
		b.WriteString(fmt.Sprintf("%-18s %s\n", "Icecast Server", icecastStatus))

		cloudflaredStatus := StyleCheck.Render("OK")
		if path, err := tunnel.CheckCloudflared(); err != nil || path == "" {
			cloudflaredStatus = StyleCross.Render("NOT INSTALLED")
		}
		b.WriteString(fmt.Sprintf("%-18s %s\n", "cloudflared", cloudflaredStatus))

		b.WriteString("\n──────────────────────────────────────────────\n\n")
		b.WriteString(fmt.Sprintf("Config File    : %s\n", config.ResolveConfigPath()))
		b.WriteString(fmt.Sprintf("Audio Directory: %s (%d tracks)\n", m.Cfg.Audio.Directory, m.TrackCount))
		b.WriteString(fmt.Sprintf("Stream Port    : %d (Host: %s)\n", m.Cfg.Streaming.Port, m.Cfg.Streaming.Host))
		b.WriteString("\n" + StyleHelp.Render("Press Esc/Enter to return to System Menu."))

	case ViewServiceStatus:
		b.WriteString(StyleTitle.Render("BACKGROUND SERVICE STATUS") + "\n\n")
		serviceSys := system.GetServiceSystemName()
		b.WriteString(fmt.Sprintf("%-18s %s\n", "Service System", serviceSys))
		b.WriteString(fmt.Sprintf("%-18s %s\n", "Service Unit", "aircoda.service"))
		b.WriteString(fmt.Sprintf("%-18s %s\n", "Start on Boot", fmt.Sprintf("%v", m.Cfg.System.StartOnBoot)))
		b.WriteString(fmt.Sprintf("%-18s %d\n", "Stream Port", m.Cfg.Streaming.Port))
		b.WriteString("\nCLI Commands for Background Service:\n")
		b.WriteString("  radio service install   (Create system user service)\n")
		b.WriteString("  radio service start     (Start background service)\n")
		b.WriteString("  radio service stop      (Stop background service)\n")
		b.WriteString("  radio service status    (Check service status)\n")
		b.WriteString("\n" + StyleHelp.Render("Press Esc/Enter to return to System Menu."))

	case ViewResetConfirm:
		b.WriteString(StyleTitle.Render("RESET AIRCODA RADIO") + "\n\n")
		b.WriteString("WARNING: This will stop all running radio processes\n")
		b.WriteString("and remove station configuration and state files.\n\n")
		b.WriteString("Are you sure you want to reset Aircoda?\n\n")
		opts := []string{"No, Cancel", "Yes, Reset Everything"}
		for i, opt := range opts {
			if i == m.ResetCursor {
				b.WriteString("  ● " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  ○ " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("↑↓ Navigate, Enter Select, Esc Cancel"))

	case ViewPublicMenu:
		b.WriteString(StyleTitle.Render("PUBLIC ACCESS SETTINGS") + "\n\n")
		opts := []string{
			"One-Click Public Stream (Cloudflare Tunnel)",
			"Access Status",
			"Direct Access",
			"Cloudflare Tunnel Settings",
		}
		for i, opt := range opts {
			if i == m.SubMenuCursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("↑↓ Navigate, Enter Select, Esc Back"))

	case ViewLogs:
		b.WriteString(StyleTitle.Render("AIRCODA RECENT LOGS") + "\n\n")
		for _, line := range m.LogLines {
			b.WriteString(line + "\n")
		}
		b.WriteString("\n" + StyleHelp.Render("Press R to refresh logs, Esc to return to main menu."))

	case ViewProgress:
		b.WriteString(StyleTitle.Render(m.ProgressTitle) + "\n\n")
		b.WriteString(m.ProgressMessage + "\n\n")
		b.WriteString(StyleHelp.Render("Processing request in background... Please wait."))

	case ViewNotice:
		b.WriteString(StyleTitle.Render("NOTICE") + "\n\n")
		b.WriteString(m.NoticeMessage + "\n\n")
		b.WriteString(StyleHelp.Render("Press Enter/Esc to return to main menu."))

	case ViewFolderScanConfirm:
		b.WriteString(StyleTitle.Render("CHANGE MUSIC FOLDER") + "\n\n")
		b.WriteString("Selected Folder:\n" + m.NewFolder + "\n\n")
		b.WriteString(fmt.Sprintf("%s Found %d audio files.\n\n", StyleCheck.Render("✓"), m.FolderScanCount))
		b.WriteString("Save this folder as station music directory?\n\n")
		b.WriteString("> " + StyleSelectedOption.Render("Save Folder") + "\n")
		b.WriteString("  " + StyleNormalOption.Render("Cancel") + "\n")

	case ViewEditFolderPathInput:
		b.WriteString("Enter the path to your music folder:\n\n")
		b.WriteString("> " + m.ManualPathBuf + "\n\n")
		b.WriteString(StyleHelp.Render("Press Enter to scan path, Esc to cancel."))

	case ViewExitPrompt:
		b.WriteString("Exit Aircoda interface?\n\n")
		opts := []string{"Yes", "No"}
		for i, opt := range opts {
			if i == m.ExitCursor {
				b.WriteString("  ● " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  ○ " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("Exiting the interface will NOT stop the radio broadcast background process."))
	}

	content := b.String()
	width, height := m.Width, m.Height
	if width <= 0 || height <= 0 {
		w, h, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && w > 0 && h > 0 {
			width, height = w, h
		}
	}

	contentHeight := lipgloss.Height(content)
	contentWidth := lipgloss.Width(content)

	w := width
	if w <= 0 {
		w = contentWidth
	}

	if height > contentHeight {
		return lipgloss.Place(w, height, lipgloss.Center, lipgloss.Center, content)
	}
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, content)
}

func startRadioCmd() tea.Cmd {
	return func() tea.Msg {
		audio.KillExistingProcesses()
		time.Sleep(300 * time.Millisecond)
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

		state, err := audio.LoadServerState()
		if err == nil && state.IsLive {
			return actionCompleteMsg{notice: "✓ Station started successfully!\n\nStream URL:\n" + state.StreamURL}
		}
		return actionCompleteMsg{notice: "Station start attempted. Check logs (Option 7) for details."}
	}
}

func stopRadioCmd() tea.Cmd {
	return func() tea.Msg {
		audio.KillExistingProcesses()
		audio.ClearServerState()
		time.Sleep(300 * time.Millisecond)
		return actionCompleteMsg{notice: "✓ Station stopped successfully."}
	}
}

func restartRadioCmd() tea.Cmd {
	return func() tea.Msg {
		audio.KillExistingProcesses()
		time.Sleep(500 * time.Millisecond)
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

		state, err := audio.LoadServerState()
		if err == nil && state.IsLive {
			return actionCompleteMsg{notice: "✓ Station restarted successfully!\n\nStream URL:\n" + state.StreamURL}
		}
		return actionCompleteMsg{notice: "Station restart failed. Check logs (Option 7) for details."}
	}
}

func startCloudflareCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		state, err := audio.LoadServerState()
		if err != nil || !state.IsLive {
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
		}

		err = tunnel.Start(cfg)
		if err != nil {
			return actionCompleteMsg{notice: fmt.Sprintf("Cloudflare setup error: %v", err)}
		}
		pubURL := public.BuildPublicStreamURL(cfg)
		return actionCompleteMsg{notice: fmt.Sprintf("✓ One-Click Cloudflare Setup Completed!\n\n✓ Radio engine active\n✓ Cloudflare binary verified / downloaded\n✓ Cloudflare Tunnel active\n\nPublic Stream URL:\n%s\n\nStatus: ● LIVE PUBLIC", pubURL)}
	}
}

func downloadYouTubeCmd(url string, dir string) tea.Cmd {
	return func() tea.Msg {
		output, err := library.DownloadYouTubeAudio(url, dir)
		if err != nil {
			return actionCompleteMsg{notice: fmt.Sprintf("YouTube download failed:\n%v", err)}
		}
		_ = output
		return actionCompleteMsg{notice: fmt.Sprintf("✓ YouTube audio downloaded successfully!\nTarget directory:\n%s\n\nPlaylist library updated.", dir)}
	}
}

func selfRepairCmd() tea.Cmd {
	return func() tea.Msg {
		results := deps.SelfRepairAll()
		var sb strings.Builder
		sb.WriteString("✓ Self-Repair & Auto-Install Results:\n\n")
		for comp, status := range results {
			sb.WriteString(fmt.Sprintf("  %-14s %s\n", comp, status))
		}
		return actionCompleteMsg{notice: sb.String()}
	}
}

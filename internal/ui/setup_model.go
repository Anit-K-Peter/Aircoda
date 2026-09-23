package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"aircoda/internal/config"
	"aircoda/internal/library"
	"aircoda/internal/picker"
)

type SetupStep int

const (
	StepRadioName SetupStep = iota
	StepRadioID
	StepMediaSource
	StepMediaSourceNotice
	StepFolderSelect
	StepFolderPathInput
	StepFolderScanResults
	StepFolderNoFiles
	StepPlaybackMode
	StepRepeat
	StepStartOnBoot
	StepReview
	StepCompletion
	StepDone
)

type SetupModel struct {
	Step           SetupStep
	Cfg            *config.Config
	FolderPicker   picker.FolderPicker
	StationName    string
	StationID      string
	MediaSource    string // "local"
	SelectedFolder string
	ManualPathBuf  string
	ScanTracks     []library.Track
	ScanCounts     map[string]int
	ScanErr        error
	PlaybackMode   string // "shuffle", "sequential"
	Repeat         bool
	StartOnBoot    bool
	Cursor         int
	Width          int
	Height         int

	// State for text inputs
	TextInputBuf string

	// Navigation & state flags
	IsFinished bool
	Cancelled  bool

	// Notice back target step
	NoticeReturnStep SetupStep
}

func isUpKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyUp || msg.String() == "up" || msg.String() == "k" || msg.String() == "p"
}

func isDownKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyDown || msg.String() == "down" || msg.String() == "j" || msg.String() == "n"
}

func isEnterKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEnter || msg.String() == "enter" || msg.String() == "\r" || msg.String() == "\n"
}

func isEscKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEsc || msg.String() == "esc"
}

func isBackspaceKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyBackspace || msg.String() == "backspace"
}

func NewSetupModel() *SetupModel {
	cfg := config.Default()
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/home/user"
	}
	defaultFolder := filepath.Join(home, "Music", "Aircoda")

	return &SetupModel{
		Step:           StepRadioName,
		Cfg:            cfg,
		FolderPicker:   picker.NewSystemFolderPicker(),
		StationName:    "986 FM",
		StationID:      "986fm",
		MediaSource:    "local",
		SelectedFolder: defaultFolder,
		PlaybackMode:   "shuffle",
		Repeat:         true,
		StartOnBoot:    true,
		TextInputBuf:   "986 FM",
		ScanCounts:     make(map[string]int),
	}
}

func (m *SetupModel) Init() tea.Cmd {
	return nil
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.Cancelled = true
			return m, tea.Quit
		}

		switch m.Step {
		case StepRadioName:
			return m.updateRadioName(msg)
		case StepRadioID:
			return m.updateRadioID(msg)
		case StepMediaSource:
			return m.updateMediaSource(msg)
		case StepMediaSourceNotice:
			return m.updateMediaSourceNotice(msg)
		case StepFolderSelect:
			return m.updateFolderSelect(msg)
		case StepFolderPathInput:
			return m.updateFolderPathInput(msg)
		case StepFolderScanResults:
			return m.updateFolderScanResults(msg)
		case StepFolderNoFiles:
			return m.updateFolderNoFiles(msg)
		case StepPlaybackMode:
			return m.updatePlaybackMode(msg)
		case StepRepeat:
			return m.updateRepeat(msg)
		case StepStartOnBoot:
			return m.updateStartOnBoot(msg)
		case StepReview:
			return m.updateReview(msg)
		case StepCompletion:
			return m.updateCompletion(msg)
		}
	}

	return m, nil
}

func (m *SetupModel) updateRadioName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		val := strings.TrimSpace(m.TextInputBuf)
		if val == "" {
			val = "986 FM"
		}
		m.StationName = val
		m.StationID = config.Slugify(m.StationName)
		m.TextInputBuf = m.StationID
		m.Step = StepRadioID
		return m, nil
	}

	if isBackspaceKey(msg) {
		if len(m.TextInputBuf) > 0 {
			m.TextInputBuf = m.TextInputBuf[:len(m.TextInputBuf)-1]
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes {
		m.TextInputBuf += string(msg.Runes)
	} else if len(msg.String()) == 1 {
		m.TextInputBuf += msg.String()
	}

	return m, nil
}

func (m *SetupModel) updateRadioID(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		val := strings.TrimSpace(m.TextInputBuf)
		if val == "" {
			val = config.Slugify(m.StationName)
		}
		slug := config.Slugify(val)
		if slug != "" {
			m.StationID = slug
			m.Step = StepMediaSource
			m.Cursor = 0
		}
		return m, nil
	}

	if isBackspaceKey(msg) {
		if len(m.TextInputBuf) > 0 {
			m.TextInputBuf = m.TextInputBuf[:len(m.TextInputBuf)-1]
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes {
		m.TextInputBuf += string(msg.Runes)
	} else if len(msg.String()) == 1 {
		m.TextInputBuf += msg.String()
	}

	return m, nil
}

func (m *SetupModel) updateMediaSource(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.Cursor > 0 {
			m.Cursor--
		}
	} else if isDownKey(msg) {
		if m.Cursor < 2 {
			m.Cursor++
		}
	} else if isEnterKey(msg) {
		switch m.Cursor {
		case 0:
			m.MediaSource = "local"
			m.Step = StepFolderSelect
			m.Cursor = 0
		case 1, 2:
			m.NoticeReturnStep = StepMediaSource
			m.Step = StepMediaSourceNotice
			m.Cursor = 0
		}
	} else if msg.String() == "1" {
		m.Cursor = 0
		m.MediaSource = "local"
		m.Step = StepFolderSelect
	} else if msg.String() == "2" || msg.String() == "3" {
		m.Cursor = 1
		m.NoticeReturnStep = StepMediaSource
		m.Step = StepMediaSourceNotice
	}
	return m, nil
}

func (m *SetupModel) updateMediaSourceNotice(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) || isDownKey(msg) {
		if m.Cursor == 0 {
			m.Cursor = 1
		} else {
			m.Cursor = 0
		}
	} else if isEnterKey(msg) {
		if m.Cursor == 0 {
			m.MediaSource = "local"
			m.Step = StepFolderSelect
			m.Cursor = 0
		} else {
			m.Step = StepMediaSource
			m.Cursor = 0
		}
	}
	return m, nil
}

func (m *SetupModel) updateFolderSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		folder, err := m.FolderPicker.PickFolder()
		if err != nil || folder == "" {
			m.ManualPathBuf = m.SelectedFolder
			m.Step = StepFolderPathInput
			return m, nil
		}
		m.SelectedFolder = folder
		m.scanFolder()
	} else if isEscKey(msg) {
		m.Step = StepMediaSource
		m.Cursor = 0
	}
	return m, nil
}

func (m *SetupModel) updateFolderPathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		val := strings.TrimSpace(m.ManualPathBuf)
		if val != "" {
			m.SelectedFolder = val
			m.scanFolder()
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
		m.Step = StepFolderSelect
		return m, nil
	}

	if msg.Type == tea.KeyRunes {
		m.ManualPathBuf += string(msg.Runes)
	} else if len(msg.String()) == 1 {
		m.ManualPathBuf += msg.String()
	}

	return m, nil
}

func (m *SetupModel) scanFolder() {
	_ = os.MkdirAll(m.SelectedFolder, 0755)
	tracks, err := library.ScanDirectory(m.SelectedFolder, true)
	m.ScanErr = err
	m.ScanTracks = tracks

	m.ScanCounts = make(map[string]int)
	for _, t := range tracks {
		ext := strings.TrimPrefix(t.Extension, ".")
		extUpper := strings.ToUpper(ext)
		m.ScanCounts[extUpper]++
	}

	if len(tracks) == 0 {
		m.Step = StepFolderNoFiles
		m.Cursor = 0
	} else {
		m.Step = StepFolderScanResults
	}
}

func (m *SetupModel) updateFolderScanResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		m.Step = StepPlaybackMode
		m.Cursor = 0
	} else if isEscKey(msg) {
		m.Step = StepFolderSelect
	}
	return m, nil
}

func (m *SetupModel) updateFolderNoFiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.Cursor > 0 {
			m.Cursor--
		}
	} else if isDownKey(msg) {
		if m.Cursor < 2 {
			m.Cursor++
		}
	} else if isEnterKey(msg) {
		switch m.Cursor {
		case 0:
			folder, err := m.FolderPicker.PickFolder()
			if err != nil || folder == "" {
				m.ManualPathBuf = m.SelectedFolder
				m.Step = StepFolderPathInput
				return m, nil
			}
			m.SelectedFolder = folder
			m.scanFolder()
		case 1:
			m.ManualPathBuf = m.SelectedFolder
			m.Step = StepFolderPathInput
		case 2:
			m.Step = StepMediaSource
			m.Cursor = 0
		}
	}
	return m, nil
}

func (m *SetupModel) updatePlaybackMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) || isDownKey(msg) {
		if m.Cursor == 0 {
			m.Cursor = 1
		} else {
			m.Cursor = 0
		}
	} else if isEnterKey(msg) {
		if m.Cursor == 0 {
			m.PlaybackMode = "shuffle"
		} else {
			m.PlaybackMode = "sequential"
		}
		m.Step = StepRepeat
		m.Cursor = 0
	} else if isEscKey(msg) {
		m.Step = StepFolderScanResults
	}
	return m, nil
}

func (m *SetupModel) updateRepeat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) || isDownKey(msg) {
		if m.Cursor == 0 {
			m.Cursor = 1
		} else {
			m.Cursor = 0
		}
	} else if isEnterKey(msg) {
		m.Repeat = (m.Cursor == 0)
		m.Step = StepStartOnBoot
		m.Cursor = 0
	} else if isEscKey(msg) {
		m.Step = StepPlaybackMode
		m.Cursor = 0
	}
	return m, nil
}

func (m *SetupModel) updateStartOnBoot(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) || isDownKey(msg) {
		if m.Cursor == 0 {
			m.Cursor = 1
		} else {
			m.Cursor = 0
		}
	} else if isEnterKey(msg) {
		m.StartOnBoot = (m.Cursor == 0)
		m.Step = StepReview
		m.Cursor = 1
	} else if isEscKey(msg) {
		m.Step = StepRepeat
		m.Cursor = 0
	}
	return m, nil
}

func (m *SetupModel) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isUpKey(msg) {
		if m.Cursor > 0 {
			m.Cursor--
		}
	} else if isDownKey(msg) {
		if m.Cursor < 2 {
			m.Cursor++
		}
	} else if isEnterKey(msg) {
		switch m.Cursor {
		case 0:
			m.Step = StepRadioName
			m.TextInputBuf = m.StationName
		case 1:
			m.saveConfig()
			m.Step = StepCompletion
		case 2:
			m.Cancelled = true
			return m, tea.Quit
		}
	} else if isEscKey(msg) {
		m.Step = StepStartOnBoot
		m.Cursor = 0
	}
	return m, nil
}

func (m *SetupModel) saveConfig() {
	m.Cfg.Station.Name = m.StationName
	m.Cfg.Station.ID = m.StationID
	m.Cfg.Streaming.Mount = "/" + m.StationID
	m.Cfg.Audio.Source = m.MediaSource
	m.Cfg.Audio.Directory = m.SelectedFolder
	m.Cfg.Playlist.Mode = m.PlaybackMode
	m.Cfg.Playlist.Repeat = m.Repeat
	m.Cfg.System.StartOnBoot = m.StartOnBoot

	_, _ = config.Save(m.Cfg, "")
}

func (m *SetupModel) updateCompletion(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isEnterKey(msg) {
		m.IsFinished = true
		m.Step = StepDone
		return m, tea.Quit
	}
	return m, nil
}

func (m *SetupModel) View() string {
	var b strings.Builder

	// Header banner with ASCII art
	header := RenderAsciiBanner("Open Source Internet Radio Setup")
	b.WriteString(header + "\n\n")

	switch m.Step {
	case StepRadioName:
		b.WriteString("Let's set up your radio.\n\n")
		b.WriteString("What would you like to name this radio?\n\n")
		b.WriteString("> " + m.TextInputBuf + "\n\n")
		b.WriteString(StyleHelp.Render("Type name and press Enter to continue."))

	case StepRadioID:
		b.WriteString(fmt.Sprintf("Radio Name: %s\n\n", m.StationName))
		b.WriteString("Choose a short ID for your radio\n\n")
		b.WriteString("> " + m.TextInputBuf + "\n\n")
		b.WriteString(StyleHelp.Render("Safe for stream mount (e.g. /" + m.StationID + "). Press Enter to confirm."))

	case StepMediaSource:
		b.WriteString("What would you like to play?\n\n")
		options := []string{
			"Local music files",
			"YouTube (Coming soon)",
			"Both (Coming soon)",
		}
		for i, opt := range options {
			if i == m.Cursor {
				b.WriteString("  ● " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  ○ " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("Use ↑/↓ or j/k to navigate, Enter to select."))

	case StepMediaSourceNotice:
		b.WriteString(StyleBadgeWarning.Render("YouTube sources are not available in this version yet.") + "\n\n")
		b.WriteString("You can continue with local music files.\n\n")
		opts := []string{"Use Local Music", "Go Back"}
		for i, opt := range opts {
			if i == m.Cursor {
				b.WriteString("  ● " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  ○ " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("Use ↑/↓ or j/k to navigate, Enter to select."))

	case StepFolderSelect:
		b.WriteString("Where are your music files?\n\n")
		b.WriteString("  " + StyleSelectedOption.Render("[ Select Folder ]") + "\n\n")
		if m.SelectedFolder != "" {
			b.WriteString("  " + m.SelectedFolder + "\n\n")
		} else {
			b.WriteString("  No folder selected yet.\n\n")
		}
		b.WriteString(StyleHelp.Render("Press Enter to choose a folder using the native picker, or Esc to go back."))

	case StepFolderPathInput:
		b.WriteString("No graphical folder picker is available (or selection was cancelled).\n\n")
		b.WriteString("Enter the path to your music folder:\n\n")
		b.WriteString("> " + m.ManualPathBuf + "\n\n")
		b.WriteString(StyleHelp.Render("Press Enter to scan path, or Esc to retry folder picker."))

	case StepFolderScanResults:
		b.WriteString("Scanning music folder...\n\n")
		b.WriteString(StyleCheck.Render("✓") + " Folder found\n")
		b.WriteString(StyleCheck.Render("✓") + " Searching recursively\n")
		b.WriteString(StyleCheck.Render("✓") + " Reading audio files\n\n")
		b.WriteString("Music folder:\n" + m.SelectedFolder + "\n\n")
		b.WriteString(fmt.Sprintf("Found %d audio files.\n\n", len(m.ScanTracks)))
		b.WriteString("Supported:\n")

		formats := []string{"MP3", "FLAC", "WAV", "OGG", "M4A", "AAC"}
		for _, fmtName := range formats {
			count := m.ScanCounts[fmtName]
			b.WriteString(fmt.Sprintf("  %-5s %d\n", fmtName, count))
		}
		b.WriteString("\n" + StyleHelp.Render("Press Enter to continue."))

	case StepFolderNoFiles:
		b.WriteString(StyleCross.Render("✕") + " No supported audio files were found.\n\n")
		b.WriteString("Supported formats include:\nMP3, WAV, FLAC, OGG, M4A, AAC\n\n")
		b.WriteString("Choose another folder:\n\n")
		opts := []string{"Select Folder", "Enter Path", "Cancel"}
		for i, opt := range opts {
			if i == m.Cursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}

	case StepPlaybackMode:
		b.WriteString("How should your music play?\n\n")
		opts := []string{"Shuffle", "In order"}
		for i, opt := range opts {
			if i == m.Cursor {
				b.WriteString("  ● " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  ○ " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("Use ↑/↓ or j/k to navigate, Enter to select."))

	case StepRepeat:
		b.WriteString("Should the playlist repeat?\n\n")
		opts := []string{"Yes", "No"}
		for i, opt := range opts {
			if i == m.Cursor {
				b.WriteString("  ● " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  ○ " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("Use ↑/↓ or j/k to navigate, Enter to select."))

	case StepStartOnBoot:
		b.WriteString("Start your radio automatically when your computer starts?\n\n")
		opts := []string{"Yes", "No"}
		for i, opt := range opts {
			if i == m.Cursor {
				b.WriteString("  ● " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  ○ " + StyleNormalOption.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + StyleHelp.Render("Use ↑/↓ or j/k to navigate, Enter to select."))

	case StepReview:
		b.WriteString(StyleTitle.Render("RADIO SETUP REVIEW") + "\n\n")

		b.WriteString(fmt.Sprintf("%-16s %s\n", "Name", m.StationName))
		b.WriteString(fmt.Sprintf("%-16s %s\n", "ID", m.StationID))
		b.WriteString(fmt.Sprintf("%-16s %s\n", "Source", m.MediaSource))
		b.WriteString(fmt.Sprintf("%-16s %s\n", "Music folder", m.SelectedFolder))
		b.WriteString(fmt.Sprintf("%-16s %d\n", "Tracks", len(m.ScanTracks)))

		pbStr := "Shuffle"
		if m.PlaybackMode == "sequential" {
			pbStr = "In order"
		}
		b.WriteString(fmt.Sprintf("%-16s %s\n", "Playback", pbStr))

		repStr := "Yes"
		if !m.Repeat {
			repStr = "No"
		}
		b.WriteString(fmt.Sprintf("%-16s %s\n", "Repeat", repStr))

		bootStr := "Yes"
		if !m.StartOnBoot {
			bootStr = "No"
		}
		b.WriteString(fmt.Sprintf("%-16s %s\n", "Start on boot", bootStr))

		b.WriteString("\n──────────────────────────────────────────────\n\n")
		opts := []string{"Edit", "Save & Continue", "Cancel"}
		for i, opt := range opts {
			if i == m.Cursor {
				b.WriteString("> " + StyleSelectedOption.Render(opt) + "\n")
			} else {
				b.WriteString("  " + StyleNormalOption.Render(opt) + "\n")
			}
		}

	case StepCompletion:
		b.WriteString(StyleTitle.Render("YOUR RADIO IS READY") + "\n\n")
		b.WriteString(StyleCheck.Render("✓") + " Configuration saved\n")
		b.WriteString(StyleCheck.Render("✓") + " Music library loaded\n")
		b.WriteString(StyleCheck.Render("✓") + " Playlist ready\n")
		b.WriteString(StyleCheck.Render("✓") + " Background service configured\n\n")
		b.WriteString("Local stream:\n" + StyleTitle.Render("http://127.0.0.1:8000/"+m.StationID) + "\n\n")
		b.WriteString(StyleHelp.Render("Press Enter to start your radio."))
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

package picker

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	ErrNoGUI     = errors.New("no graphical desktop environment available")
	ErrCancelled = errors.New("folder selection cancelled")
)

// FolderPicker defines the interface for opening native folder selection dialogs.
type FolderPicker interface {
	PickFolder() (string, error)
}

// SystemFolderPicker implements FolderPicker using platform-native dialog tools.
type SystemFolderPicker struct{}

// NewSystemFolderPicker returns a new SystemFolderPicker instance.
func NewSystemFolderPicker() FolderPicker {
	return &SystemFolderPicker{}
}

// PickFolder opens a native desktop folder picker dialog based on the host OS.
func (p *SystemFolderPicker) PickFolder() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return pickLinuxFolder()
	case "darwin":
		return pickMacOSFolder()
	case "windows":
		return pickWindowsFolder()
	default:
		return "", ErrNoGUI
	}
}

func pickLinuxFolder() (string, error) {
	// Check if desktop display server is present
	display := os.Getenv("DISPLAY")
	wayland := os.Getenv("WAYLAND_DISPLAY")
	if display == "" && wayland == "" {
		return "", ErrNoGUI
	}

	// 1. Try zenity
	if path, err := exec.LookPath("zenity"); err == nil && path != "" {
		cmd := exec.Command("zenity", "--file-selection", "--directory", "--title=Select your music folder")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			selected := strings.TrimSpace(out.String())
			if selected != "" {
				return selected, nil
			}
		}
	}

	// 2. Try kdialog
	if path, err := exec.LookPath("kdialog"); err == nil && path != "" {
		cmd := exec.Command("kdialog", "--getexistingdirectory", ".", "--title", "Select your music folder")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			selected := strings.TrimSpace(out.String())
			if selected != "" {
				return selected, nil
			}
		}
	}

	// 3. Try python3 tkinter fallback
	if path, err := exec.LookPath("python3"); err == nil && path != "" {
		pyScript := `import tkinter, tkinter.filedialog; root = tkinter.Tk(); root.withdraw(); root.wm_attributes('-topmost', 1); path = tkinter.filedialog.askdirectory(title='Select your music folder'); print(path)`
		cmd := exec.Command("python3", "-c", pyScript)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			selected := strings.TrimSpace(out.String())
			if selected != "" {
				return selected, nil
			}
		}
	}

	return "", ErrNoGUI
}

func pickMacOSFolder() (string, error) {
	appleScript := `POSIX path of (choose folder with prompt "Select your music folder:")`
	cmd := exec.Command("osascript", "-e", appleScript)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", ErrCancelled
	}
	selected := strings.TrimSpace(out.String())
	if selected == "" {
		return "", ErrCancelled
	}
	return selected, nil
}

func pickWindowsFolder() (string, error) {
	psScript := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = 'Select your music folder'; if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Host $f.SelectedPath }`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", ErrCancelled
	}
	selected := strings.TrimSpace(out.String())
	if selected == "" {
		return "", ErrCancelled
	}
	return selected, nil
}

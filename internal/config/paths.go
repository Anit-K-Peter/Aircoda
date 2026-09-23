package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	ConfigFileName = "config.toml"
	LogFileName    = "aircoda.log"
)

// GetSystemConfigDir returns platform system-wide configuration directory.
func GetSystemConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		if programData := os.Getenv("ProgramData"); programData != "" {
			return filepath.Join(programData, "Aircoda")
		}
		return `C:\ProgramData\Aircoda`
	case "darwin":
		return "/Library/Application Support/Aircoda"
	default:
		return "/etc/aircoda"
	}
}

// GetSystemLogDir returns platform system-wide log directory.
func GetSystemLogDir() string {
	switch runtime.GOOS {
	case "windows":
		if programData := os.Getenv("ProgramData"); programData != "" {
			return filepath.Join(programData, "Aircoda", "Logs")
		}
		return `C:\ProgramData\Aircoda\Logs`
	case "darwin":
		return "/Library/Logs/Aircoda"
	default:
		return "/var/log/aircoda"
	}
}

// GetUserConfigDir returns platform user-level configuration directory.
func GetUserConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" && runtime.GOOS == "linux" {
		return filepath.Join(xdg, "aircoda")
	}

	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, errHome := os.UserHomeDir()
		if errHome != nil {
			home = os.TempDir()
		}
		return filepath.Join(home, ".config", "aircoda")
	}

	return filepath.Join(dir, "Aircoda")
}

// GetUserBinDir returns platform user-level local binary directory (~/.config/Aircoda/bin).
func GetUserBinDir() string {
	return filepath.Join(GetUserConfigDir(), "bin")
}

// GetUserStateDir returns platform user-level state and logs directory.
func GetUserStateDir() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" && runtime.GOOS == "linux" {
		return filepath.Join(xdg, "aircoda")
	}

	switch runtime.GOOS {
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "Aircoda", "Logs")
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Logs", "Aircoda")
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".local", "state", "aircoda")
}

// ResolveConfigPath returns existing or preferred path to config file.
func ResolveConfigPath() string {
	sysPath := filepath.Join(GetSystemConfigDir(), ConfigFileName)
	if _, err := os.Stat(sysPath); err == nil {
		return sysPath
	}

	userPath := filepath.Join(GetUserConfigDir(), ConfigFileName)
	if _, err := os.Stat(userPath); err == nil {
		return userPath
	}

	if isDirWritableOrCreatable(GetSystemConfigDir()) {
		return sysPath
	}

	return userPath
}

// ResolveLogPath returns appropriate log file path.
func ResolveLogPath() string {
	sysLog := filepath.Join(GetSystemLogDir(), LogFileName)
	if isDirWritableOrCreatable(GetSystemLogDir()) {
		return sysLog
	}

	userStateDir := GetUserStateDir()
	return filepath.Join(userStateDir, LogFileName)
}

func isDirWritableOrCreatable(dir string) bool {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false
	}
	testFile := filepath.Join(dir, ".perm_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return false
	}
	_ = os.Remove(testFile)
	return true
}

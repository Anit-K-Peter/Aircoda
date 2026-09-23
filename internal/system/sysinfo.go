package system

import (
	"runtime"
	"strings"
)

var (
	Version   = "v0.1.0"
	Commit    = "dev"
	BuildDate = "unknown"
)

// GetVersion returns current Aircoda software version.
func GetVersion() string {
	return Version
}

// GetCommit returns git commit hash if available.
func GetCommit() string {
	return Commit
}

// GetBuildDate returns binary compilation date.
func GetBuildDate() string {
	return BuildDate
}

// GetOS returns current operating system identifier.
func GetOS() string {
	return runtime.GOOS
}

// GetPlatformName returns user-friendly platform title (Linux, Windows, macOS).
func GetPlatformName() string {
	switch runtime.GOOS {
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	default:
		return strings.Title(runtime.GOOS)
	}
}

// GetServiceSystemName returns native OS background service manager name.
func GetServiceSystemName() string {
	switch runtime.GOOS {
	case "linux":
		return "systemd"
	case "windows":
		return "Windows Service"
	case "darwin":
		return "launchd"
	default:
		return "Manual"
	}
}

// GetArch returns current system architecture identifier.
func GetArch() string {
	return runtime.GOARCH
}

// GetGoVersion returns the Go runtime version used for compilation.
func GetGoVersion() string {
	return runtime.Version()
}

package tests

import (
	"runtime"
	"strings"
	"testing"

	"aircoda/internal/config"
)

func TestCrossPlatformPaths(t *testing.T) {
	sysCfg := config.GetSystemConfigDir()
	userCfg := config.GetUserConfigDir()
	sysLog := config.GetSystemLogDir()
	userState := config.GetUserStateDir()

	if sysCfg == "" || userCfg == "" || sysLog == "" || userState == "" {
		t.Fatal("Path resolution returned empty string")
	}

	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(sysCfg, "ProgramData") && !strings.Contains(sysCfg, "Aircoda") {
			t.Errorf("Unexpected Windows system config path: %s", sysCfg)
		}
	case "darwin":
		if !strings.Contains(sysCfg, "Library") {
			t.Errorf("Unexpected macOS system config path: %s", sysCfg)
		}
	case "linux":
		if sysCfg != "/etc/aircoda" {
			t.Errorf("Unexpected Linux system config path: %s", sysCfg)
		}
	}
}

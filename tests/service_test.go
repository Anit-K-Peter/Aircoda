package tests

import (
	"runtime"
	"testing"

	"aircoda/internal/service"
	"aircoda/internal/system"
)

func TestServiceSupport(t *testing.T) {
	supported := service.IsServiceSupported()
	if runtime.GOOS == "linux" || runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		if !supported {
			t.Errorf("Expected service support to be true on %s", runtime.GOOS)
		}
	}

	platformName := system.GetPlatformName()
	if platformName == "" {
		t.Error("Platform name returned empty string")
	}

	sysName := system.GetServiceSystemName()
	if sysName == "" {
		t.Error("Service system name returned empty string")
	}
}

package tests

import (
	"testing"

	"aircoda/internal/config"
	"aircoda/internal/tunnel"
)

func TestTunnelStatusDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := config.Default()
	st, err := tunnel.GetStatus(cfg)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if st == nil {
		t.Fatal("Expected non-nil tunnel state")
	}
	if st.IsRunning {
		t.Error("Expected default tunnel status to be not running")
	}
}

func TestCheckCloudflared(t *testing.T) {
	_, err := tunnel.CheckCloudflared()
	// Should return either valid path or descriptive error, not panic
	_ = err
}

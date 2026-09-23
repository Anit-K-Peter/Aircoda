package tests

import (
	"strings"
	"testing"

	"aircoda/internal/config"
	"aircoda/internal/public"
)

func TestBuildPublicStreamURL(t *testing.T) {
	cfg := config.Default()
	cfg.Streaming.Host = "127.0.0.1"
	cfg.Streaming.Port = 8000
	cfg.Streaming.Mount = "/986fm"

	// Local mode
	cfg.Public.Mode = "local"
	urlLocal := public.BuildPublicStreamURL(cfg)
	expectedLocal := "http://127.0.0.1:8000/986fm"
	if urlLocal != expectedLocal {
		t.Errorf("Expected local URL %q, got %q", expectedLocal, urlLocal)
	}

	// Direct mode with hostname
	cfg.Public.Mode = "direct"
	cfg.Public.Hostname = "203.0.113.20"
	urlDirect := public.BuildPublicStreamURL(cfg)
	expectedDirect := "http://203.0.113.20:8000/986fm"
	if urlDirect != expectedDirect {
		t.Errorf("Expected direct URL %q, got %q", expectedDirect, urlDirect)
	}

	// Tunnel mode with domain
	cfg.Public.Mode = "tunnel"
	cfg.Tunnel.Hostname = "radio.example.com"
	urlTunnel := public.BuildPublicStreamURL(cfg)
	expectedTunnel := "https://radio.example.com/986fm"
	if urlTunnel != expectedTunnel {
		t.Errorf("Expected tunnel URL %q, got %q", expectedTunnel, urlTunnel)
	}
}

func TestDetectPublicIP(t *testing.T) {
	ip := public.DetectPublicIP()
	if ip == "" {
		t.Error("Expected non-empty IP string from DetectPublicIP")
	}
	if !strings.Contains(ip, ".") && !strings.Contains(ip, ":") {
		t.Errorf("Unexpected IP format: %s", ip)
	}
}

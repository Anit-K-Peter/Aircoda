package tests

import (
	"net"
	"strings"
	"testing"

	"aircoda/internal/audio"
	"aircoda/internal/config"
)

func TestBuildStreamURL(t *testing.T) {
	cfg := config.Default()
	cfg.Streaming.Host = "127.0.0.1"
	cfg.Streaming.Port = 8000
	cfg.Streaming.Mount = "/986fm"

	url := audio.BuildStreamURL(cfg)
	expected := "http://127.0.0.1:8000/986fm"
	if url != expected {
		t.Errorf("BuildStreamURL() = %q, expected %q", url, expected)
	}

	// Without leading slash in mount
	cfg.Streaming.Mount = "testmount"
	url2 := audio.BuildStreamURL(cfg)
	expected2 := "http://127.0.0.1:8000/testmount"
	if url2 != expected2 {
		t.Errorf("BuildStreamURL() = %q, expected %q", url2, expected2)
	}
}

func TestGenerateIcecastConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Station.Name = "Jazz Station"
	cfg.Streaming.Port = 8000
	cfg.Streaming.Mount = "/jazz"

	xml := audio.GenerateIcecastConfig(cfg, "/tmp/icecast_logs")

	if !strings.Contains(xml, "<port>8000</port>") {
		t.Error("Generated XML missing port 8000")
	}
	if !strings.Contains(xml, "<mount-name>/jazz</mount-name>") {
		t.Error("Generated XML missing mount /jazz")
	}
	if !strings.Contains(xml, "<stream-name>Jazz Station</stream-name>") {
		t.Error("Generated XML missing station name")
	}
}

func TestCheckPortAvailable(t *testing.T) {
	// Free random port check
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to bind random port: %v", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	port := addr.Port

	// Currently open port should report occupied
	if err := audio.CheckPortAvailable("127.0.0.1", port); err == nil {
		t.Errorf("Expected port conflict error for open port %d, got nil", port)
	}

	// Close listener
	l.Close()

	// Now port should be available
	if err := audio.CheckPortAvailable("127.0.0.1", port); err != nil {
		t.Errorf("Expected port %d to be available after close, got error: %v", port, err)
	}
}

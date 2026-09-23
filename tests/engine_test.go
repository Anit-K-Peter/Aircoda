package tests

import (
	"os"
	"testing"
	"time"

	"aircoda/internal/audio"
)

func TestServerStatePersistence(t *testing.T) {
	// Ensure clean start
	audio.ClearServerState()

	state := &ServerStateTestStruct{
		IsLive:       true,
		IcecastPID:   1234,
		FFmpegPID:    5678,
		StartTime:    time.Now(),
		CurrentTrack: "song1.mp3",
		NextTrack:    "song2.wav",
		StreamURL:    "http://127.0.0.1:8000/986fm",
		Bitrate:      128,
		Listeners:    "unavailable",
	}

	st := &audio.ServerState{
		IsLive:       state.IsLive,
		IcecastPID:   state.IcecastPID,
		FFmpegPID:    state.FFmpegPID,
		StartTime:    state.StartTime,
		CurrentTrack: state.CurrentTrack,
		NextTrack:    state.NextTrack,
		StreamURL:    state.StreamURL,
		Bitrate:      state.Bitrate,
		Listeners:    state.Listeners,
	}

	if err := audio.SaveServerState(st); err != nil {
		t.Fatalf("Failed to save server state: %v", err)
	}

	loaded, err := audio.LoadServerState()
	if err != nil {
		t.Fatalf("Failed to load server state: %v", err)
	}

	if !loaded.IsLive || loaded.CurrentTrack != "song1.mp3" || loaded.StreamURL != "http://127.0.0.1:8000/986fm" {
		t.Errorf("Loaded state data mismatch: %+v", loaded)
	}

	audio.ClearServerState()
	if _, err := audio.LoadServerState(); err == nil {
		t.Error("Expected error loading cleared state, got nil")
	}
}

type ServerStateTestStruct struct {
	IsLive       bool
	IcecastPID   int
	FFmpegPID    int
	StartTime    time.Time
	CurrentTrack string
	NextTrack    string
	StreamURL    string
	Bitrate      int
	Listeners    string
}

func TestIsProcessRunning(t *testing.T) {
	currentPID := os.Getpid()
	if !audio.IsProcessRunning(currentPID) {
		t.Errorf("Expected current process PID %d to be running", currentPID)
	}

	invalidPID := 999999
	if audio.IsProcessRunning(invalidPID) {
		t.Errorf("Expected invalid PID %d to return false", invalidPID)
	}
}

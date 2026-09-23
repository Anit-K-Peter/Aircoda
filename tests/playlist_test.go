package tests

import (
	"testing"

	"aircoda/internal/library"
	"aircoda/internal/playlist"
)

func createDummyTracks() []library.Track {
	return []library.Track{
		{Path: "/music/track1.mp3", Filename: "track1.mp3", Extension: ".mp3"},
		{Path: "/music/track2.wav", Filename: "track2.wav", Extension: ".wav"},
		{Path: "/music/track3.flac", Filename: "track3.flac", Extension: ".flac"},
	}
}

func TestSequentialQueue(t *testing.T) {
	tracks := createDummyTracks()
	qs := playlist.NewQueueState(tracks, playlist.ModeSequential, true)

	// Current track should be track 1
	curr, ok := qs.Current()
	if !ok || curr.Filename != "track1.mp3" {
		t.Fatalf("Expected track1.mp3 as current, got %v", curr)
	}

	// Advance to track 2
	next1, ok := qs.Next()
	if !ok || next1.Filename != "track2.wav" {
		t.Fatalf("Expected track2.wav as next1, got %v", next1)
	}

	// Advance to track 3
	next2, ok := qs.Next()
	if !ok || next2.Filename != "track3.flac" {
		t.Fatalf("Expected track3.flac as next2, got %v", next2)
	}

	// Wrap around (repeat = true)
	next3, ok := qs.Next()
	if !ok || next3.Filename != "track1.mp3" {
		t.Fatalf("Expected wrap around to track1.mp3, got %v", next3)
	}
}

func TestSequentialQueueNoRepeat(t *testing.T) {
	tracks := createDummyTracks()
	qs := playlist.NewQueueState(tracks, playlist.ModeSequential, false)

	qs.Next() // move to track 2
	qs.Next() // move to track 3

	// Next when repeat is false should return false
	_, ok := qs.Next()
	if ok {
		t.Error("Expected end of queue (ok=false) when repeat=false, but got true")
	}
}

func TestShuffleQueue(t *testing.T) {
	tracks := createDummyTracks()
	qs := playlist.NewQueueState(tracks, playlist.ModeShuffle, true)

	if len(qs.PlayOrder) != len(tracks) {
		t.Fatalf("Play order length %d != tracks length %d", len(qs.PlayOrder), len(tracks))
	}

	// Verify all indices exist in play order
	seen := make(map[int]bool)
	for _, idx := range qs.PlayOrder {
		seen[idx] = true
	}
	if len(seen) != len(tracks) {
		t.Errorf("Shuffle order missing track indices: %v", qs.PlayOrder)
	}
}

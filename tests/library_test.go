package tests

import (
	"os"
	"path/filepath"
	"testing"

	"aircoda/internal/library"
)

func TestSupportedAudioFormatDetection(t *testing.T) {
	supported := []string{"track.mp3", "audio.WAV", "music.ogg", "song.flac", "file.m4a", "clip.aac"}
	unsupported := []string{"document.txt", "photo.jpg", "video.mp4", "script.sh", ".hidden.mp3"}

	for _, name := range supported {
		if !library.IsSupportedAudio(name) {
			t.Errorf("Expected %s to be supported audio, but returned false", name)
		}
	}

	for _, name := range unsupported {
		if library.IsSupportedAudio(name) && name[0] != '.' {
			t.Errorf("Expected %s to be unsupported audio, but returned true", name)
		}
	}
}

func TestScanDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create test directory structure:
	// tempDir/
	// ├── song1.mp3
	// ├── song2.wav
	// ├── text.txt
	// ├── .hidden.mp3
	// └── sub/
	//     ├── song3.flac
	//     └── .hidden_sub/
	//         └── ignored.ogg

	filesToCreate := []string{
		"song1.mp3",
		"song2.wav",
		"text.txt",
		".hidden.mp3",
		filepath.Join("sub", "song3.flac"),
		filepath.Join("sub", ".hidden_sub", "ignored.ogg"),
	}

	for _, rel := range filesToCreate {
		full := filepath.Join(tempDir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(full, []byte("dummy audio content"), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", full, err)
		}
	}

	// 1. Recursive scan
	tracks, err := library.ScanDirectory(tempDir, true)
	if err != nil {
		t.Fatalf("ScanDirectory(recursive=true) failed: %v", err)
	}

	if len(tracks) != 3 {
		t.Errorf("Expected 3 tracks (song1.mp3, song2.wav, sub/song3.flac), got %d", len(tracks))
		for _, tr := range tracks {
			t.Logf("Found track: %s", tr.Path)
		}
	}

	// 2. Non-recursive scan
	nonRecTracks, err := library.ScanDirectory(tempDir, false)
	if err != nil {
		t.Fatalf("ScanDirectory(recursive=false) failed: %v", err)
	}

	if len(nonRecTracks) != 2 {
		t.Errorf("Expected 2 root tracks (song1.mp3, song2.wav), got %d", len(nonRecTracks))
	}
}

func TestScanMissingAndEmptyDirectory(t *testing.T) {
	// Missing directory
	missingDir := filepath.Join(t.TempDir(), "non_existent_folder")
	_, err := library.ScanDirectory(missingDir, true)
	if err == nil {
		t.Error("Expected error scanning missing directory, got nil")
	}

	// Empty directory
	emptyDir := t.TempDir()
	tracks, err := library.ScanDirectory(emptyDir, true)
	if err != nil {
		t.Fatalf("Unexpected error scanning empty directory: %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("Expected 0 tracks in empty directory, got %d", len(tracks))
	}
}

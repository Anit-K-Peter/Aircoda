package library

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var SupportedExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".ogg":  true,
	".flac": true,
	".m4a":  true,
	".aac":  true,
}

// IsSupportedAudio returns true if the filename extension is a supported audio format.
func IsSupportedAudio(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return SupportedExtensions[ext]
}

// ScanDirectory scans the given directory path for supported audio files.
func ScanDirectory(dirPath string, recursive bool) ([]Track, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(dirPath))
	if cleanPath == "" {
		return nil, fmt.Errorf("audio directory is not set")
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("audio directory does not exist: %s", cleanPath)
		}
		return nil, fmt.Errorf("cannot access audio directory (%s): %w", cleanPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("audio path is not a directory: %s", cleanPath)
	}

	var tracks []Track

	if recursive {
		err = filepath.WalkDir(cleanPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip inaccessible subdirectories
			}

			name := d.Name()
			// Skip hidden files and hidden directories
			if strings.HasPrefix(name, ".") && path != cleanPath {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if !d.IsDir() && IsSupportedAudio(name) {
				fileInfo, err := d.Info()
				size := int64(0)
				if err == nil {
					size = fileInfo.Size()
				}

				ext := strings.ToLower(filepath.Ext(name))
				tracks = append(tracks, Track{
					Path:      path,
					Filename:  name,
					Extension: ext,
					SizeBytes: size,
				})
			}
			return nil
		})
	} else {
		entries, readErr := os.ReadDir(cleanPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read audio directory: %w", readErr)
		}

		for _, d := range entries {
			name := d.Name()
			if strings.HasPrefix(name, ".") || d.IsDir() {
				continue
			}

			if IsSupportedAudio(name) {
				fileInfo, err := d.Info()
				size := int64(0)
				if err == nil {
					size = fileInfo.Size()
				}

				ext := strings.ToLower(filepath.Ext(name))
				tracks = append(tracks, Track{
					Path:      filepath.Join(cleanPath, name),
					Filename:  name,
					Extension: ext,
					SizeBytes: size,
				})
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error scanning audio directory: %w", err)
	}

	// Sort tracks deterministically by filepath
	sort.Slice(tracks, func(i, j int) bool {
		return tracks[i].Path < tracks[j].Path
	})

	return tracks, nil
}

package library

import "time"

// Track represents an audio track discovered in the local audio library.
type Track struct {
	Path      string        `json:"path"`
	Filename  string        `json:"filename"`
	Extension string        `json:"extension"`
	SizeBytes int64         `json:"size_bytes"`
	Duration  time.Duration `json:"duration,omitempty"`
}

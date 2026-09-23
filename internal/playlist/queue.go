package playlist

import (
	"math/rand"
	"time"

	"aircoda/internal/library"
)

const (
	ModeSequential = "sequential"
	ModeShuffle    = "shuffle"
)

// QueueState encapsulates playback queue items and cursor positions.
type QueueState struct {
	Tracks       []library.Track `json:"tracks"`
	PlayOrder    []int           `json:"play_order"`
	CurrentIndex int             `json:"current_index"`
	Mode         string          `json:"mode"`
	Repeat       bool            `json:"repeat"`
}

// NewQueueState creates a fresh queue state from track list.
func NewQueueState(tracks []library.Track, mode string, repeat bool) *QueueState {
	if mode == "" {
		mode = ModeShuffle
	}
	qs := &QueueState{
		Tracks:       tracks,
		CurrentIndex: 0,
		Mode:         mode,
		Repeat:       repeat,
	}
	qs.RebuildOrder()
	return qs
}

// RebuildOrder generates the play index sequence based on mode.
func (qs *QueueState) RebuildOrder() {
	n := len(qs.Tracks)
	qs.PlayOrder = make([]int, n)
	for i := 0; i < n; i++ {
		qs.PlayOrder[i] = i
	}

	if qs.Mode == ModeShuffle && n > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(n, func(i, j int) {
			qs.PlayOrder[i], qs.PlayOrder[j] = qs.PlayOrder[j], qs.PlayOrder[i]
		})
	}

	if qs.CurrentIndex >= n {
		qs.CurrentIndex = 0
	}
}

// Current returns the current track without advancing the queue cursor.
func (qs *QueueState) Current() (*library.Track, bool) {
	if len(qs.Tracks) == 0 {
		return nil, false
	}
	if qs.CurrentIndex < 0 || qs.CurrentIndex >= len(qs.PlayOrder) {
		qs.CurrentIndex = 0
	}
	trackIdx := qs.PlayOrder[qs.CurrentIndex]
	return &qs.Tracks[trackIdx], true
}

// Next advances the queue cursor and returns the next track.
func (qs *QueueState) Next() (*library.Track, bool) {
	if len(qs.Tracks) == 0 {
		return nil, false
	}

	nextIndex := qs.CurrentIndex + 1
	if nextIndex >= len(qs.PlayOrder) {
		if !qs.Repeat {
			return nil, false
		}
		if qs.Mode == ModeShuffle && len(qs.Tracks) > 1 {
			qs.RebuildOrder()
		}
		nextIndex = 0
	}

	qs.CurrentIndex = nextIndex
	trackIdx := qs.PlayOrder[qs.CurrentIndex]
	return &qs.Tracks[trackIdx], true
}

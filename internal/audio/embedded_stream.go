package audio

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"aircoda/internal/logger"
)

// StreamHub broadcasts audio byte chunks to multiple HTTP listeners.
type StreamHub struct {
	mu        sync.RWMutex
	listeners map[chan []byte]bool
	server    *http.Server
	listener  net.Listener
	mount     string
}

// NewStreamHub creates a new HTTP streaming server.
func NewStreamHub(mount string) *StreamHub {
	if !strings.HasPrefix(mount, "/") {
		mount = "/" + mount
	}
	return &StreamHub{
		listeners: make(map[chan []byte]bool),
		mount:     mount,
	}
}

// Start launches the HTTP audio stream server.
func (h *StreamHub) Start(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind HTTP stream server to %s: %w", addr, err)
	}
	h.listener = ln

	mux := http.NewServeMux()
	handler := http.HandlerFunc(h.handleStream)
	mux.Handle(h.mount, handler)
	mux.Handle("/", handler)

	h.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // Continuous streaming
	}

	go func() {
		_ = h.server.Serve(ln)
	}()

	logger.Info("Embedded HTTP audio stream server listening at http://%s%s", addr, h.mount)
	return nil
}

// handleStream serves live MP3 stream to HTTP clients.
func (h *StreamHub) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Server", "Aircoda/v0.1.0")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan []byte, 100)

	h.mu.Lock()
	h.listeners[ch] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.listeners, ch)
		h.mu.Unlock()
		close(ch)
	}()

	// Send initial HTTP headers to client immediately
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case data, open := <-ch:
			if !open {
				return
			}
			if _, err := w.Write(data); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// Write broadcasts a chunk of encoded audio bytes to all active listeners.
func (h *StreamHub) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	buf := make([]byte, len(p))
	copy(buf, p)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.listeners {
		select {
		case ch <- buf:
		default:
			// Drop chunk if listener buffer full to avoid blocking audio pipeline
		}
	}

	return len(p), nil
}

// Stop shuts down the embedded streaming server.
func (h *StreamHub) Stop() {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = h.server.Shutdown(ctx)
	}
	if h.listener != nil {
		_ = h.listener.Close()
	}
}

// StreamPipeWriter wraps io.Writer for FFmpeg stdout.
type StreamPipeWriter struct {
	Hub *StreamHub
}

func (w *StreamPipeWriter) Write(p []byte) (n int, err error) {
	if w.Hub == nil {
		return len(p), nil
	}
	return w.Hub.Write(p)
}

var _ io.Writer = (*StreamPipeWriter)(nil)

package audio

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"aircoda/internal/logger"
)

// StandaloneIcecastServer provides a pure Go Icecast 2.0 compatible server on 127.0.0.1:8000.
type StandaloneIcecastServer struct {
	listener  net.Listener
	mu        sync.RWMutex
	listeners map[chan []byte]bool
	mount     string
}

// NewStandaloneIcecastServer constructs an Icecast-compatible listener server.
func NewStandaloneIcecastServer(mount string) *StandaloneIcecastServer {
	if !strings.HasPrefix(mount, "/") {
		mount = "/" + mount
	}
	return &StandaloneIcecastServer{
		listeners: make(map[chan []byte]bool),
		mount:     mount,
	}
}

// ListenAndServe starts TCP server handling both Icecast SOURCE streams and HTTP GET listeners.
func (s *StandaloneIcecastServer) ListenAndServe(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind Icecast server to %s: %w", addr, err)
	}
	s.listener = ln

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go s.handleConnection(conn)
		}
	}()

	logger.Info("Icecast server listening on http://%s%s", addr, s.mount)
	return nil
}

func (s *StandaloneIcecastServer) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	reqLine, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return
	}

	parts := strings.Fields(reqLine)
	if len(parts) < 2 {
		conn.Close()
		return
	}

	method := strings.ToUpper(parts[0])
	_ = parts[1] // request path

	// Read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(line) == "" {
			break
		}
	}

	if method == "SOURCE" || method == "PUT" {
		// FFmpeg streaming input connection
		conn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
		buf := make([]byte, 8192)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				s.broadcast(buf[:n])
			}
			if err != nil {
				break
			}
		}
		conn.Close()
	} else if method == "GET" {
		// Listener client (mpv, ffplay, browser, VLC)
		header := "HTTP/1.0 200 OK\r\n" +
			"Content-Type: audio/mpeg\r\n" +
			"Cache-Control: no-cache, no-store\r\n" +
			"Pragma: no-cache\r\n" +
			"Connection: close\r\n" +
			"Server: Icecast 2.4.4 (Aircoda)\r\n\r\n"
		conn.Write([]byte(header))

		ch := make(chan []byte, 100)
		s.mu.Lock()
		s.listeners[ch] = true
		s.mu.Unlock()

		defer func() {
			s.mu.Lock()
			delete(s.listeners, ch)
			s.mu.Unlock()
			conn.Close()
		}()

		for data := range ch {
			if _, err := conn.Write(data); err != nil {
				return
			}
		}
	} else {
		conn.Write([]byte("HTTP/1.0 404 Not Found\r\n\r\n"))
		conn.Close()
	}
}

func (s *StandaloneIcecastServer) broadcast(chunk []byte) {
	if len(chunk) == 0 {
		return
	}
	buf := make([]byte, len(chunk))
	copy(buf, chunk)

	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.listeners {
		select {
		case ch <- buf:
		default:
		}
	}
}

// Stop closes the server listener.
func (s *StandaloneIcecastServer) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
	}
}

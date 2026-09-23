package public

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aircoda/internal/config"
	"aircoda/internal/tunnel"
)

// BuildPublicStreamURL constructs the public listener URL based on configuration.
func BuildPublicStreamURL(cfg *config.Config) string {
	mount := cfg.Streaming.Mount
	if !strings.HasPrefix(mount, "/") {
		mount = "/" + mount
	}

	mode := strings.ToLower(strings.TrimSpace(cfg.Public.Mode))
	if mode == "" {
		mode = "local"
	}

	switch mode {
	case "direct":
		host := strings.TrimSpace(cfg.Public.Hostname)
		if host == "" {
			host = DetectPublicIP()
		}
		if host == "" {
			host = cfg.Streaming.Host
		}
		if host == "0.0.0.0" {
			host = "127.0.0.1"
		}
		return fmt.Sprintf("http://%s:%d%s", host, cfg.Streaming.Port, mount)

	case "tunnel":
		host := strings.TrimSpace(cfg.Tunnel.Hostname)
		if host == "" {
			host = strings.TrimSpace(cfg.Public.Hostname)
		}
		if host == "" {
			if st, err := tunnel.LoadState(); err == nil && st != nil && st.PublicURL != "" {
				host = st.PublicURL
			}
		}
		if host == "" {
			return fmt.Sprintf("http://127.0.0.1:%d%s", cfg.Streaming.Port, mount)
		}
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "https://" + host
		}
		return fmt.Sprintf("%s%s", strings.TrimSuffix(host, "/"), mount)

	case "local":
		fallthrough
	default:
		host := cfg.Streaming.Host
		if host == "0.0.0.0" {
			host = "127.0.0.1"
		}
		return fmt.Sprintf("http://%s:%d%s", host, cfg.Streaming.Port, mount)
	}
}

// DetectPublicIP attempts to discover the host's external WAN IP address.
func DetectPublicIP() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	endpoints := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	client := &http.Client{Timeout: 2 * time.Second}

	for _, endpoint := range endpoints {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				ip := strings.TrimSpace(string(body))
				if ip != "" {
					return ip
				}
			}
		}
	}

	return "127.0.0.1"
}

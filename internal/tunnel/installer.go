package tunnel

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"aircoda/internal/config"
	"aircoda/internal/logger"
)

// GetLocalBinPath returns directory for Aircoda local binaries (~/.config/Aircoda/bin).
func GetLocalBinPath() string {
	configDir := filepath.Dir(config.ResolveConfigPath())
	return filepath.Join(configDir, "bin")
}

// EnsureCloudflared checks if cloudflared binary is installed, auto-downloading it if missing.
func EnsureCloudflared() (string, error) {
	// 1. Check system PATH
	if path, err := exec.LookPath("cloudflared"); err == nil && path != "" {
		return path, nil
	}

	// 2. Check local Aircoda bin directory
	localBinDir := GetLocalBinPath()
	binName := "cloudflared"
	if runtime.GOOS == "windows" {
		binName = "cloudflared.exe"
	}
	targetPath := filepath.Join(localBinDir, binName)

	if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
		return targetPath, nil
	}

	// 3. Auto-download cloudflared binary from official Cloudflare releases
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create local bin directory (%s): %w", localBinDir, err)
	}

	downloadURL := getCloudflaredDownloadURL()
	if downloadURL == "" {
		return "", fmt.Errorf("unsupported OS/architecture (%s/%s) for automatic cloudflared download", runtime.GOOS, runtime.GOARCH)
	}

	logger.Info("Downloading cloudflared binary from %s...", downloadURL)
	tmpPath := targetPath + ".download"
	if err := downloadFile(downloadURL, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to download cloudflared: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to set execution permission on cloudflared: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to save cloudflared binary: %w", err)
	}

	logger.Info("cloudflared installed successfully at %s", targetPath)
	return targetPath, nil
}

func getCloudflaredDownloadURL() string {
	baseURL := "https://github.com/cloudflare/cloudflared/releases/latest/download/"

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return baseURL + "cloudflared-linux-amd64"
		case "arm64":
			return baseURL + "cloudflared-linux-arm64"
		case "386":
			return baseURL + "cloudflared-linux-386"
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return baseURL + "cloudflared-darwin-amd64.tgz"
		case "arm64":
			return baseURL + "cloudflared-darwin-amd64.tgz"
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return baseURL + "cloudflared-windows-amd64.exe"
		case "386":
			return baseURL + "cloudflared-windows-386.exe"
		}
	}
	return ""
}

func downloadFile(url string, destPath string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error %d (%s)", resp.StatusCode, resp.Status)
	}

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

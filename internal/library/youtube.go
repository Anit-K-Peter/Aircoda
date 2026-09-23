package library

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"aircoda/internal/deps"
)

// CheckYtDlp verifies whether yt-dlp is installed on the host system or local bin directory.
func CheckYtDlp() bool {
	ok, _ := deps.FindBinary("yt-dlp")
	return ok
}

// DownloadYouTubeAudio downloads audio from a YouTube link using yt-dlp and converts it to MP3 in targetDir.
func DownloadYouTubeAudio(url string, targetDir string) (string, error) {
	ytPath, err := deps.EnsureYtDlp()
	if err != nil {
		return "", fmt.Errorf("yt-dlp auto-installation failed: %w", err)
	}

	if strings.TrimSpace(url) == "" {
		return "", fmt.Errorf("YouTube URL cannot be empty")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to access target directory: %w", err)
	}

	outputTemplate := filepath.Join(targetDir, "%(title)s [%(id)s].%(ext)s")

	cmd := exec.Command(ytPath,
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", outputTemplate,
		"--no-playlist",
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %s", string(output))
	}

	return string(output), nil
}

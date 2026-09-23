package deps

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"aircoda/internal/config"
	"aircoda/internal/logger"
	"aircoda/internal/tunnel"
)

// CheckResult contains detailed dependency check status.
type CheckResult struct {
	FFmpeg      bool
	FFmpegPath  string
	FFprobe     bool
	FFprobePath string
	Icecast     bool
	IcecastPath string
	YtDlp       bool
	YtDlpPath   string
	Cloudflared bool
	CloudPath   string
}

// FindBinary searches system PATH, Aircoda bin directory, and ~/.local/bin for executable.
func FindBinary(name string) (bool, string) {
	// 1. Check system PATH
	if path, err := exec.LookPath(name); err == nil && path != "" {
		return true, path
	}

	binName := name
	if runtime.GOOS == "windows" && !strings.HasSuffix(name, ".exe") {
		binName = name + ".exe"
	}

	// 2. Check Aircoda UserBinDir
	localBin := filepath.Join(config.GetUserBinDir(), binName)
	if info, err := os.Stat(localBin); err == nil && !info.IsDir() {
		EnsureBinInPATH(config.GetUserBinDir())
		return true, localBin
	}

	// 3. Check ~/.local/bin
	if home, err := os.UserHomeDir(); err == nil {
		userLocalBin := filepath.Join(home, ".local", "bin", binName)
		if info, err := os.Stat(userLocalBin); err == nil && !info.IsDir() {
			EnsureBinInPATH(filepath.Dir(userLocalBin))
			return true, userLocalBin
		}
	}

	return false, ""
}

// EnsureBinInPATH prepends a directory to current process PATH environment variable.
func EnsureBinInPATH(dir string) {
	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, dir) {
		os.Setenv("PATH", dir+string(os.PathListSeparator)+currentPath)
	}
}

// CheckDependencies verifies if ffmpeg, ffprobe, icecast, yt-dlp, and cloudflared exist.
func CheckDependencies() (*CheckResult, error) {
	res := &CheckResult{}

	res.FFmpeg, res.FFmpegPath = FindBinary("ffmpeg")
	res.FFprobe, res.FFprobePath = FindBinary("ffprobe")
	res.Icecast, res.IcecastPath = FindBinary("icecast")
	if !res.Icecast {
		res.Icecast, res.IcecastPath = FindBinary("icecast2")
	}
	res.YtDlp, res.YtDlpPath = FindBinary("yt-dlp")
	res.Cloudflared, res.CloudPath = FindBinary("cloudflared")

	if !res.FFmpeg {
		return res, fmt.Errorf("FFmpeg is required but missing. Run 'radio repair' to auto-install dependencies.")
	}

	return res, nil
}

// EnsureYtDlp guarantees yt-dlp binary presence by downloading it automatically if missing.
func EnsureYtDlp() (string, error) {
	if ok, path := FindBinary("yt-dlp"); ok {
		return path, nil
	}

	binDir := config.GetUserBinDir()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create local bin directory (%s): %w", binDir, err)
	}

	binName := "yt-dlp"
	downloadURL := "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	if runtime.GOOS == "windows" {
		binName = "yt-dlp.exe"
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	}

	targetPath := filepath.Join(binDir, binName)
	logger.Info("[Self-Repair] Auto-downloading yt-dlp binary to %s...", targetPath)

	tmpPath := targetPath + ".download"
	if err := downloadFile(downloadURL, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to download yt-dlp: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to set execution permissions on yt-dlp: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed to save yt-dlp executable: %w", err)
	}

	EnsureBinInPATH(binDir)
	logger.Info("[Self-Repair] yt-dlp installed successfully at %s", targetPath)
	return targetPath, nil
}

// EnsureCloudflared auto-installs cloudflared static binary.
func EnsureCloudflared() (string, error) {
	return tunnel.EnsureCloudflared()
}

// SelfRepairAll attempts to auto-repair missing dependencies without stopping the system.
func SelfRepairAll() map[string]string {
	results := make(map[string]string)

	// 1. yt-dlp repair
	if _, err := EnsureYtDlp(); err != nil {
		results["yt-dlp"] = fmt.Sprintf("FAILED: %v", err)
	} else {
		results["yt-dlp"] = "INSTALLED / READY"
	}

	// 2. cloudflared repair
	if _, err := EnsureCloudflared(); err != nil {
		results["cloudflared"] = fmt.Sprintf("FAILED: %v", err)
	} else {
		results["cloudflared"] = "INSTALLED / READY"
	}

	// 3. FFmpeg check & system package manager suggestion
	if ok, path := FindBinary("ffmpeg"); ok {
		results["ffmpeg"] = fmt.Sprintf("INSTALLED (%s)", path)
	} else {
		results["ffmpeg"] = attemptPackageInstall("ffmpeg")
	}

	// 4. Icecast check & fallback notice
	if ok, path := FindBinary("icecast"); ok {
		results["icecast"] = fmt.Sprintf("INSTALLED (%s)", path)
	} else if ok2, path2 := FindBinary("icecast2"); ok2 {
		results["icecast"] = fmt.Sprintf("INSTALLED (%s)", path2)
	} else {
		results["icecast"] = "NOT INSTALLED (Aircoda will automatically use Built-in Embedded Streamer)"
	}

	return results
}

func attemptPackageInstall(pkg string) string {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("apt-get"); err == nil {
		cmd = exec.Command("sudo", "apt-get", "install", "-y", pkg)
	} else if _, err := exec.LookPath("dnf"); err == nil {
		cmd = exec.Command("sudo", "dnf", "install", "-y", pkg)
	} else if _, err := exec.LookPath("pacman"); err == nil {
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", pkg)
	} else if _, err := exec.LookPath("brew"); err == nil {
		cmd = exec.Command("brew", "install", pkg)
	}

	if cmd != nil {
		logger.Info("[Self-Repair] Attempting system package manager auto-install for %s...", pkg)
		out, err := cmd.CombinedOutput()
		if err == nil {
			return fmt.Sprintf("AUTO-INSTALLED VIA SYSTEM PACKAGE MANAGER (%s)", strings.TrimSpace(string(out)))
		}
	}

	return fmt.Sprintf("MISSING — Please install '%s' via system package manager", pkg)
}

func downloadFile(url string, destPath string) error {
	client := &http.Client{Timeout: 90 * time.Second}
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

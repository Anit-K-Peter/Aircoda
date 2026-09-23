package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"aircoda/internal/audio"
	"aircoda/internal/config"
	"aircoda/internal/deps"
	"aircoda/internal/library"
	"aircoda/internal/public"
	"aircoda/internal/service"
	"aircoda/internal/system"
)

// RunDoctor performs comprehensive system diagnostics and outputs health report.
func RunDoctor(autoFix bool) error {
	if autoFix {
		return RunRepair()
	}

	fmt.Printf("Aircoda System Doctor (%s)\n", system.GetVersion())
	fmt.Println("──────────────────────────────────────────────────")
	fmt.Printf("Platform         %s (%s)\n", system.GetPlatformName(), system.GetArch())
	fmt.Printf("Go Runtime       %s\n", system.GetGoVersion())
	fmt.Println()

	fmt.Println("Dependencies")
	ffmpegStatus, ffmpegPath := deps.FindBinary("ffmpeg")
	fmt.Printf("  FFmpeg         %-14s\n", formatCheck(ffmpegStatus, ffmpegPath))

	ffprobeStatus, ffprobePath := deps.FindBinary("ffprobe")
	fmt.Printf("  FFprobe        %-14s\n", formatCheck(ffprobeStatus, ffprobePath))

	icecastStatus, icecastPath := deps.FindBinary("icecast")
	if !icecastStatus {
		icecastStatus, icecastPath = deps.FindBinary("icecast2")
	}
	if icecastStatus {
		fmt.Printf("  Icecast        %-14s\n", formatCheck(true, icecastPath))
	} else {
		fmt.Printf("  Icecast        %-14s\n", "OPTIONAL (Built-in Embedded Streamer active)")
	}

	ytdlpStatus, ytdlpPath := deps.FindBinary("yt-dlp")
	fmt.Printf("  yt-dlp         %-14s\n", formatCheck(ytdlpStatus, ytdlpPath))
	fmt.Println()

	fmt.Println("Environment")

	// Config check
	cfg, cfgPath, cfgErr := config.Load()
	if cfgErr != nil {
		fmt.Printf("  Configuration  FAILED (%s)\n", cfgErr.Error())
	} else {
		fmt.Printf("  Configuration  OK (%s)\n", cfgPath)
	}

	// Audio Folder check
	if cfg != nil && cfg.Audio.Directory != "" {
		tracks, scanErr := library.ScanDirectory(cfg.Audio.Directory, cfg.Playlist.Recursive)
		if scanErr != nil {
			fmt.Printf("  Audio Folder   FAILED (%s)\n", scanErr.Error())
		} else {
			fmt.Printf("  Audio Folder   OK (%s - %d tracks)\n", cfg.Audio.Directory, len(tracks))
		}
	} else {
		fmt.Println("  Audio Folder   NOT CONFIGURED (Run 'radio setup')")
	}

	// Port check
	port := 8000
	if cfg != nil && cfg.Streaming.Port > 0 {
		port = cfg.Streaming.Port
	}
	host := "127.0.0.1"
	if cfg != nil && cfg.Streaming.Host != "" {
		host = cfg.Streaming.Host
	}

	portErr := audio.CheckPortAvailable(host, port)
	if portErr != nil {
		if strings.Contains(portErr.Error(), "already in use") {
			if state, err := audio.LoadServerState(); err == nil && state.IsLive {
				fmt.Printf("  Stream Port    OK (Port %d in use by Aircoda)\n", port)
			} else {
				fmt.Printf("  Stream Port    CONFLICT (Port %d in use by another process)\n", port)
			}
		} else {
			fmt.Printf("  Stream Port    FAILED (%v)\n", portErr)
		}
	} else {
		fmt.Printf("  Stream Port    OK (Port %d available)\n", port)
	}

	// Service support check
	svcSysName := system.GetServiceSystemName()
	if service.IsServiceSupported() {
		fmt.Printf("  Service        OK (%s supported)\n", svcSysName)
	} else {
		fmt.Printf("  Service        UNSUPPORTED (Manual mode available)\n")
	}

	fmt.Println()
	fmt.Println("Public Access")
	publicMode := "local"
	if cfg != nil && cfg.Public.Mode != "" {
		publicMode = cfg.Public.Mode
	}
	fmt.Printf("  Mode           %s\n", strings.ToUpper(publicMode))

	cloudflaredStatus, cloudflaredPath := deps.FindBinary("cloudflared")
	if publicMode == "tunnel" {
		fmt.Printf("  cloudflared    %-14s\n", formatCheck(cloudflaredStatus, cloudflaredPath))
	} else {
		if cloudflaredStatus {
			fmt.Printf("  cloudflared    OK (%s)\n", cloudflaredPath)
		} else {
			fmt.Println("  cloudflared    OPTIONAL (Auto-downloads on demand)")
		}
	}

	if cfg != nil {
		fmt.Printf("  Listener URL   %s\n", public.BuildPublicStreamURL(cfg))
	}

	fmt.Println("──────────────────────────────────────────────────")

	if ffmpegStatus && cfgErr == nil && (portErr == nil || strings.Contains(portErr.Error(), "already in use")) {
		fmt.Println("Status: All critical system checks passed. Ready for radio broadcast.")
	} else {
		fmt.Println("Status: Some dependencies or settings need attention.")
		fmt.Println("Tip: Run 'radio repair' or 'radio doctor --fix' to auto-repair missing components.")
	}

	return nil
}

func formatCheck(ok bool, detail string) string {
	if ok {
		return fmt.Sprintf("OK (%s)", detail)
	}
	return fmt.Sprintf("MISSING (%s)", detail)
}

func newDoctorCmd() *cobra.Command {
	var autoFix bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run comprehensive system diagnostics and check dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDoctor(autoFix)
		},
	}
	cmd.Flags().BoolVarP(&autoFix, "fix", "f", false, "Automatically repair missing dependencies")
	return cmd
}

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"aircoda/internal/config"
	"aircoda/internal/public"
	"aircoda/internal/tunnel"
)

// NewTunnelCmd returns parent Cobra command for 'radio tunnel'.
func NewTunnelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Manage Cloudflare Tunnel integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunTunnelStatus()
		},
	}

	cmd.AddCommand(newTunnelStatusCmd())
	cmd.AddCommand(newTunnelSetupCmd())
	cmd.AddCommand(newTunnelStartCmd())
	cmd.AddCommand(newTunnelStopCmd())
	cmd.AddCommand(newTunnelRestartCmd())
	cmd.AddCommand(newTunnelEnableCmd())
	cmd.AddCommand(newTunnelDisableCmd())
	cmd.AddCommand(newTunnelRemoveCmd())

	return cmd
}

func newTunnelStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display Cloudflare tunnel status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunTunnelStatus()
		},
	}
}

func newTunnelSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Run interactive Cloudflare tunnel setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunTunnelSetup()
		},
	}
}

func newTunnelStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start Cloudflare tunnel process",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.Load()
			if err != nil {
				return err
			}
			if err := tunnel.Start(cfg); err != nil {
				return err
			}
			fmt.Println("Cloudflare tunnel started successfully.")
			return RunTunnelStatus()
		},
	}
}

func newTunnelStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop Cloudflare tunnel process",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tunnel.Stop(); err != nil {
				return err
			}
			fmt.Println("Cloudflare tunnel stopped.")
			return nil
		},
	}
}

func newTunnelRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart Cloudflare tunnel process",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.Load()
			if err != nil {
				return err
			}
			if err := tunnel.Restart(cfg); err != nil {
				return err
			}
			fmt.Println("Cloudflare tunnel restarted successfully.")
			return RunTunnelStatus()
		},
	}
}

func newTunnelEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable Cloudflare tunnel mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, cfgPath, err := config.Load()
			if err != nil {
				return err
			}
			cfg.Tunnel.Enabled = true
			cfg.Public.Mode = "tunnel"
			if _, err := config.Save(cfg, cfgPath); err != nil {
				return err
			}
			fmt.Println("Cloudflare tunnel mode enabled.")
			return RunTunnelStatus()
		},
	}
}

func newTunnelDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable Cloudflare tunnel mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = tunnel.Stop()
			cfg, cfgPath, err := config.Load()
			if err != nil {
				return err
			}
			cfg.Tunnel.Enabled = false
			if cfg.Public.Mode == "tunnel" {
				cfg.Public.Mode = "local"
			}
			if _, err := config.Save(cfg, cfgPath); err != nil {
				return err
			}
			fmt.Println("Cloudflare tunnel mode disabled.")
			return nil
		},
	}
}

func newTunnelRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove Cloudflare tunnel configuration and stop process",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = tunnel.Stop()
			cfg, cfgPath, err := config.Load()
			if err != nil {
				return err
			}
			cfg.Tunnel = config.TunnelConfig{
				Enabled:  false,
				Provider: "cloudflare",
				Target:   fmt.Sprintf("127.0.0.1:%d", cfg.Streaming.Port),
			}
			if cfg.Public.Mode == "tunnel" {
				cfg.Public.Mode = "local"
			}
			if _, err := config.Save(cfg, cfgPath); err != nil {
				return err
			}
			fmt.Println("Cloudflare tunnel configuration removed.")
			return nil
		},
	}
}

// RunTunnelStatus displays Cloudflare tunnel runtime status and listener URL.
func RunTunnelStatus() error {
	cfg, _, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	st, _ := tunnel.GetStatus(cfg)

	statusStr := "DISCONNECTED"
	if st != nil && st.IsRunning {
		statusStr = "CONNECTED"
	}

	hostname := cfg.Tunnel.Hostname
	if hostname == "" {
		hostname = cfg.Public.Hostname
	}
	if hostname == "" && st != nil && st.PublicURL != "" {
		hostname = st.PublicURL
	}
	if hostname == "" {
		hostname = "Not configured"
	}

	target := cfg.Tunnel.Target
	if target == "" {
		target = fmt.Sprintf("127.0.0.1:%d", cfg.Streaming.Port)
	}

	fmt.Println()
	fmt.Println("╭────────────────────────────────────────────╮")
	fmt.Println("│              CLOUDFLARE TUNNEL             │")
	fmt.Println("╰────────────────────────────────────────────╯")
	fmt.Println()
	fmt.Printf("Status      : %s\n", statusStr)
	fmt.Printf("Hostname    : %s\n", hostname)
	fmt.Printf("Target      : %s\n", target)
	fmt.Println()
	fmt.Println("Listener:")
	if hostname != "Not configured" {
		fmt.Printf("%s\n", public.BuildPublicStreamURL(cfg))
	} else {
		fmt.Println("Run 'radio tunnel setup' to configure your public hostname.")
	}
	fmt.Println()

	return nil
}

// RunTunnelSetup performs interactive setup wizard for Cloudflare tunnel.
func RunTunnelSetup() error {
	cfg, cfgPath, err := config.Load()
	if err != nil {
		cfg = config.Default()
		cfgPath = config.ResolveConfigPath()
	}

	if _, errBin := tunnel.EnsureCloudflared(); errBin != nil {
		fmt.Println()
		fmt.Printf("Warning: Failed to locate or download cloudflared: %v\n", errBin)
		fmt.Println()
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("╭────────────────────────────────────────────╮")
	fmt.Println("│           CLOUDFLARE TUNNEL SETUP          │")
	fmt.Println("╰────────────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("Cloudflare Tunnel setup options:")
	fmt.Println("  1. Quick Tunnel (Automatic temporary trycloudflare.com URL)")
	fmt.Println("  2. Custom Domain (e.g. radio.example.com)")
	fmt.Println()
	fmt.Print("Choice [1-2] (default 1): ")

	inputChoice, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(inputChoice)

	if choice == "2" {
		fmt.Println()
		fmt.Print("Domain / hostname (e.g. radio.example.com): ")
		domainInput, _ := reader.ReadString('\n')
		cfg.Tunnel.Hostname = strings.TrimSpace(domainInput)
		cfg.Public.Hostname = cfg.Tunnel.Hostname
	} else {
		cfg.Tunnel.Hostname = ""
	}

	targetDefault := fmt.Sprintf("127.0.0.1:%d", cfg.Streaming.Port)
	fmt.Println()
	fmt.Printf("Local stream target [%s]: ", targetDefault)
	targetInput, _ := reader.ReadString('\n')
	targetVal := strings.TrimSpace(targetInput)
	if targetVal == "" {
		targetVal = targetDefault
	}
	cfg.Tunnel.Target = targetVal

	cfg.Tunnel.Enabled = true
	cfg.Public.Mode = "tunnel"

	savedPath, err := config.Save(cfg, cfgPath)
	if err != nil {
		return fmt.Errorf("failed to save tunnel configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("Configuration saved to:", savedPath)
	fmt.Println("Starting Cloudflare tunnel background service...")
	
	if err := tunnel.Start(cfg); err != nil {
		fmt.Printf("Notice: Tunnel configuration saved, but process failed to start: %v\n", err)
	} else {
		fmt.Println("✓ Cloudflare Tunnel process started successfully in background.")
	}

	fmt.Println()
	return RunTunnelStatus()
}

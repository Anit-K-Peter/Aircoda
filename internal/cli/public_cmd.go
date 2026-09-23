package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"aircoda/internal/config"
	"aircoda/internal/public"
)

// NewPublicCmd returns the parent Cobra command for 'radio public'.
func NewPublicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public",
		Short: "Manage station public access and listener URLs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPublicStatus()
		},
	}

	cmd.AddCommand(newPublicStatusCmd())
	cmd.AddCommand(newPublicSetupCmd())
	cmd.AddCommand(newPublicEnableCmd())
	cmd.AddCommand(newPublicDisableCmd())

	return cmd
}

func newPublicStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display public access status and listener URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPublicStatus()
		},
	}
}

func newPublicSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure public stream access mode and hostname",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPublicSetup()
		},
	}
}

func newPublicEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable direct or tunnel public access mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPublicEnable()
		},
	}
}

func newPublicDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable public access (reset mode to local)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPublicDisable()
		},
	}
}

// RunPublicStatus renders public stream configuration and listener URL.
func RunPublicStatus() error {
	cfg, _, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	mode := strings.ToUpper(cfg.Public.Mode)
	if mode == "" {
		mode = "LOCAL"
	}

	statusStr := "DISABLED"
	if mode != "LOCAL" {
		statusStr = "AVAILABLE"
	}

	publicURL := public.BuildPublicStreamURL(cfg)

	fmt.Println()
	fmt.Println("╭────────────────────────────────────────────╮")
	fmt.Println("│             PUBLIC ACCESS                  │")
	fmt.Println("╰────────────────────────────────────────────╯")
	fmt.Println()
	fmt.Printf("Mode       : %s\n", mode)
	fmt.Printf("Status     : %s\n", statusStr)
	fmt.Println()
	fmt.Println("Listener URL:")
	fmt.Printf("%s\n", publicURL)
	fmt.Println()

	if mode == "LOCAL" {
		fmt.Println("This stream is not publicly accessible.")
		fmt.Println("To expose your station publicly, run: radio public setup")
	} else if mode == "DIRECT" {
		fmt.Println("Direct public access active.")
		fmt.Println("Ensure TCP port", cfg.Streaming.Port, "is open and forwarded on your router/firewall.")
	} else if mode == "TUNNEL" {
		fmt.Println("Cloudflare Tunnel mode active.")
		fmt.Println("Listeners connect securely via Cloudflare HTTPS.")
	}
	fmt.Println()

	return nil
}

// RunPublicSetup runs interactive prompt wizard for public access configuration.
func RunPublicSetup() error {
	cfg, cfgPath, err := config.Load()
	if err != nil {
		cfg = config.Default()
		cfgPath = config.ResolveConfigPath()
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("╭────────────────────────────────────────────╮")
	fmt.Println("│          PUBLIC ACCESS SETUP               │")
	fmt.Println("╰────────────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("Select how listeners should reach your radio station:")
	fmt.Println("  1. LOCAL ONLY  - Private local stream (127.0.0.1)")
	fmt.Println("  2. DIRECT      - Public IP address + port")
	fmt.Println("  3. TUNNEL      - Cloudflare Tunnel HTTPS hostname")
	fmt.Println()
	fmt.Print("Choice [1-3] (default 1): ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	switch choice {
	case "2":
		cfg.Public.Mode = "direct"
		fmt.Println()
		fmt.Print("Enter public domain name or IP address (leave blank for auto-detect): ")
		hostInput, _ := reader.ReadString('\n')
		cfg.Public.Hostname = strings.TrimSpace(hostInput)
	case "3":
		cfg.Public.Mode = "tunnel"
		cfg.Tunnel.Enabled = true
		fmt.Println()
		fmt.Print("Enter your Cloudflare tunnel domain (e.g. radio.example.com): ")
		domainInput, _ := reader.ReadString('\n')
		domain := strings.TrimSpace(domainInput)
		cfg.Public.Hostname = domain
		cfg.Tunnel.Hostname = domain
	case "1":
		fallthrough
	default:
		cfg.Public.Mode = "local"
		cfg.Public.Hostname = ""
	}

	savedPath, err := config.Save(cfg, cfgPath)
	if err != nil {
		return fmt.Errorf("failed to save public configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("Public access configuration saved to:", savedPath)
	return RunPublicStatus()
}

// RunPublicEnable enables direct public mode or activates existing configuration.
func RunPublicEnable() error {
	cfg, cfgPath, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.Public.Mode == "local" || cfg.Public.Mode == "" {
		if cfg.Tunnel.Hostname != "" {
			cfg.Public.Mode = "tunnel"
		} else {
			cfg.Public.Mode = "direct"
		}
	}

	if _, err := config.Save(cfg, cfgPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("Public access enabled (Mode:", strings.ToUpper(cfg.Public.Mode) + ")")
	return RunPublicStatus()
}

// RunPublicDisable resets public mode to local.
func RunPublicDisable() error {
	cfg, cfgPath, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg.Public.Mode = "local"
	cfg.Tunnel.Enabled = false

	if _, err := config.Save(cfg, cfgPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("Public access disabled. Stream is now private (LOCAL ONLY).")
	return RunPublicStatus()
}

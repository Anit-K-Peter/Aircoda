package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"aircoda/internal/config"
	"aircoda/internal/logger"
)

// NewRootCmd constructs and returns the root Cobra command for radio.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "radio",
		Short: "Aircoda - Self-hosted internet radio server",
		Long:  `Aircoda is an open-source, cross-platform, self-hosted internet radio server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.Load()
			if err != nil {
				if err == config.ErrConfigNotFound {
					// First-run behavior: trigger setup wizard
					return RunSetup()
				}
				return fmt.Errorf("configuration error: %w", err)
			}

			// Validate loaded config
			if err := config.Validate(cfg); err != nil {
				logger.PrintUserError(fmt.Errorf("configuration file is invalid: %w", err))
				fmt.Println("Launching setup wizard to reconfigure...")
				return RunSetup()
			}

			// Configured: open normal CLI menu
			return RunMenu(cfg)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Subcommands
	rootCmd.AddCommand(newSetupCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(NewConfigCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(NewPlaylistCmd())
	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(newStopCmd())
	rootCmd.AddCommand(newRestartCmd())
	rootCmd.AddCommand(newLogsCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(NewServiceCmd())
	rootCmd.AddCommand(newDaemonCmd())
	rootCmd.AddCommand(NewPublicCmd())
	rootCmd.AddCommand(NewTunnelCmd())
	rootCmd.AddCommand(newResetCmd())
	rootCmd.AddCommand(newRepairCmd())

	// Future Phase Placeholders
	futureCommands := []string{"update"}
	for _, name := range futureCommands {
		rootCmd.AddCommand(newFutureCmd(name))
	}

	return rootCmd
}

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Run the interactive first-run setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunSetup()
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display station and service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunStatus()
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display Aircoda version and build environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunVersion()
		},
	}
}

func newFutureCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Execute %s command (future phase)", name),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Not implemented yet.")
		},
	}
}

// Execute runs the main application CLI entrypoint.
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		logger.PrintUserError(err)
		os.Exit(1)
	}
}

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"aircoda/internal/audio"
	"aircoda/internal/config"
	"aircoda/internal/playlist"

)

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset station configuration while preserving audio files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunReset()
		},
	}
}

// RunReset prompts for explicit confirmation before resetting configuration.
func RunReset() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("This will reset your Aircoda station configuration.")
	fmt.Println("Your music files will NOT be deleted.")
	fmt.Println()
	fmt.Print("Continue? [y/N]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}

	confirmation := strings.ToLower(strings.TrimSpace(input))
	if confirmation != "y" && confirmation != "yes" {
		fmt.Println("Reset cancelled.")
		return nil
	}

	// Stop any active processes cleanly
	_, _ = audio.LoadServerState()
	audio.KillExistingProcesses()
	playlist.ClearState()

	if err := config.Reset(); err != nil {
		return fmt.Errorf("failed to reset station configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("Station configuration reset successfully.")
	fmt.Println("Run 'radio setup' to re-initialize your radio station.")
	fmt.Println()
	return nil
}

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"aircoda/internal/deps"
)

// RunRepair executes automatic self-repair for missing dependencies and configuration issues.
func RunRepair() error {
	fmt.Println("Aircoda Self-Repair System")
	fmt.Println("──────────────────────────────────────────────────")
	fmt.Println("Scanning host environment and repairing missing dependencies...")
	fmt.Println()

	results := deps.SelfRepairAll()
	for component, status := range results {
		fmt.Printf("  %-15s %s\n", component, status)
	}

	fmt.Println("──────────────────────────────────────────────────")
	fmt.Println("Self-repair completed. All core and optional components checked.")
	return nil
}

func newRepairCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repair",
		Short: "Auto-install missing dependencies and self-repair station environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunRepair()
		},
	}
}

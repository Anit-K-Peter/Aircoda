package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"aircoda/internal/config"
)

// NewConfigCmd returns the parent Cobra command for 'radio config'.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display and manage station configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunConfig()
		},
	}

	cmd.AddCommand(newConfigBackupCmd())
	cmd.AddCommand(newConfigRestoreCmd())

	return cmd
}

func newConfigBackupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backup [destination_file]",
		Short: "Backup station configuration TOML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			dest := ""
			if len(args) > 0 {
				dest = args[0]
			}
			backupPath, err := config.Backup(dest)
			if err != nil {
				return err
			}
			fmt.Println("Configuration backup saved to:", backupPath)
			return nil
		},
	}
}

func newConfigRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore <backup_file>",
		Short: "Restore station configuration from a backup file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupFile := args[0]
			if err := config.Restore(backupFile); err != nil {
				return err
			}
			fmt.Println("Configuration restored successfully from:", backupFile)
			return nil
		},
	}
}

// RunConfig displays the current configuration file path and contents.
func RunConfig() error {
	cfgPath := config.ResolveConfigPath()

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no configuration file found at %s. Run 'radio setup' to initialize", cfgPath)
		}
		return fmt.Errorf("error reading configuration file (%s): %w", cfgPath, err)
	}

	fmt.Printf("Configuration Path: %s\n", cfgPath)
	fmt.Println("----------------------------------------")
	fmt.Print(string(data))
	return nil
}

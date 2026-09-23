package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"aircoda/internal/service"
)

// NewServiceCmd constructs the 'radio service' subcommand router.
func NewServiceCmd() *cobra.Command {
	svcCmd := &cobra.Command{
		Use:   "service",
		Short: "Manage OS background service (systemd / Windows Service / launchd)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceStatus()
		},
	}

	svcCmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install Aircoda as an OS background service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceInstall()
		},
	})

	svcCmd.AddCommand(&cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Aircoda background service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceUninstall()
		},
	})

	svcCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start Aircoda OS background service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceStart()
		},
	})

	svcCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop Aircoda OS background service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceStop()
		},
	})

	svcCmd.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart Aircoda OS background service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceRestart()
		},
	})

	svcCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Display Aircoda OS background service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceStatus()
		},
	})

	svcCmd.AddCommand(&cobra.Command{
		Use:   "enable",
		Short: "Enable Aircoda OS background service on system boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceInstall()
		},
	})

	svcCmd.AddCommand(&cobra.Command{
		Use:   "disable",
		Short: "Disable Aircoda OS background service from system boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServiceUninstall()
		},
	})

	return svcCmd
}

func getServiceManager() (service.ServiceManager, error) {
	mgr, err := service.NewServiceManager()
	if err != nil {
		return nil, fmt.Errorf("service initialization error: %w", err)
	}
	return mgr, nil
}

func RunServiceInstall() error {
	mgr, err := getServiceManager()
	if err != nil {
		return err
	}
	if err := mgr.Install(); err != nil {
		return err
	}
	fmt.Printf("Service successfully installed (%s).\n", mgr.SystemName())
	return nil
}

func RunServiceUninstall() error {
	mgr, err := getServiceManager()
	if err != nil {
		return err
	}
	if err := mgr.Uninstall(); err != nil {
		return err
	}
	fmt.Println("Service successfully uninstalled.")
	return nil
}

func RunServiceStart() error {
	mgr, err := getServiceManager()
	if err != nil {
		return err
	}
	if err := mgr.Start(); err != nil {
		return err
	}
	fmt.Println("Service started.")
	return nil
}

func RunServiceStop() error {
	mgr, err := getServiceManager()
	if err != nil {
		return err
	}
	if err := mgr.Stop(); err != nil {
		return err
	}
	fmt.Println("Service stopped.")
	return nil
}

func RunServiceRestart() error {
	mgr, err := getServiceManager()
	if err != nil {
		return err
	}
	if err := mgr.Restart(); err != nil {
		return err
	}
	fmt.Println("Service restarted.")
	return nil
}

func RunServiceStatus() error {
	mgr, err := getServiceManager()
	if err != nil {
		return err
	}
	st := mgr.Status()
	fmt.Printf("Service System: %s\n", mgr.SystemName())
	fmt.Printf("Service Status: %s\n", strings.ToUpper(st))
	return nil
}

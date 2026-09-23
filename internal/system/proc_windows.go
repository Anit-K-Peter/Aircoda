//go:build windows

package system

import "syscall"

// GetDetachSysProcAttr returns platform-specific process detachment attributes for Windows systems.
func GetDetachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
	}
}

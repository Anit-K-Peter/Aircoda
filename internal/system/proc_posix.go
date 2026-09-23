//go:build !windows

package system

import "syscall"

// GetDetachSysProcAttr returns platform-specific process detachment attributes for POSIX systems.
func GetDetachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

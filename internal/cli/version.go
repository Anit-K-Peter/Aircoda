package cli

import (
	"fmt"

	"aircoda/internal/system"
)

// RunVersion prints version and system runtime information.
func RunVersion() error {
	fmt.Printf("Aircoda %s\n", system.GetVersion())
	fmt.Printf("commit: %s\n", system.GetCommit())
	fmt.Printf("platform: %s/%s\n", system.GetOS(), system.GetArch())
	fmt.Printf("go: %s\n", system.GetGoVersion())
	return nil
}

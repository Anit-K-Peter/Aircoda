package cli

import (
	"aircoda/internal/ui"
)

// RunSetup Wizard executes the interactive first-run terminal setup.
func RunSetup() error {
	return ui.RunTUI(nil, true)
}


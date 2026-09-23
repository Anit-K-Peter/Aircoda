package cli

import (
	"aircoda/internal/config"
	"aircoda/internal/ui"
)

// RunMenu presents the main interactive menu for configured Aircoda instances.
func RunMenu(cfg *config.Config) error {
	return ui.RunTUI(cfg, false)
}


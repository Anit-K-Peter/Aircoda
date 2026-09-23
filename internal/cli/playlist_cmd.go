package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"aircoda/internal/config"
	"aircoda/internal/playlist"
)

// NewPlaylistCmd constructs the 'radio playlist' command group.
func NewPlaylistCmd() *cobra.Command {
	playlistCmd := &cobra.Command{
		Use:   "playlist",
		Short: "Manage station audio library and playback queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPlaylistList()
		},
	}

	playlistCmd.AddCommand(&cobra.Command{
		Use:   "scan",
		Short: "Rescan audio directory and update library tracks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPlaylistScan()
		},
	})

	playlistCmd.AddCommand(&cobra.Command{
		Use:   "current",
		Short: "Display the current track in the playback queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPlaylistCurrent()
		},
	})

	playlistCmd.AddCommand(&cobra.Command{
		Use:   "next",
		Short: "Advance and display the next track in the queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunPlaylistNext()
		},
	})

	return playlistCmd
}

// RunPlaylistList displays the discovered audio library files.
func RunPlaylistList() error {
	cfg, _, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("Aircoda is not configured yet. Run 'radio setup' to initialize")
		}
		return err
	}

	mgr := playlist.NewManager(cfg)
	_, tracks, err := mgr.GetOrInitQueue()
	if err != nil {
		return err
	}

	fmt.Printf("AIRCODA — %s\n\n", strings.ToUpper(cfg.Station.Name))
	fmt.Println("Audio Library")
	fmt.Println("─────────────────────────────────")
	fmt.Println()

	if len(tracks) == 0 {
		fmt.Printf("No supported audio files were found in:\n%s\n", cfg.Audio.Directory)
		return nil
	}

	for i, track := range tracks {
		fmt.Printf("%d. %s\n", i+1, track.Filename)
	}

	fmt.Println()
	fmt.Printf("Total: %d tracks\n", len(tracks))
	return nil
}

// RunPlaylistScan rescans directory and updates queue.
func RunPlaylistScan() error {
	cfg, _, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("Aircoda is not configured yet. Run 'radio setup' to initialize")
		}
		return err
	}

	mgr := playlist.NewManager(cfg)
	_, tracks, err := mgr.Rescan()
	if err != nil {
		return err
	}

	fmt.Println("Playlist rescan complete.")
	fmt.Printf("Audio directory: %s\n", cfg.Audio.Directory)
	fmt.Printf("Total tracks found: %d\n", len(tracks))
	return nil
}

// RunPlaylistCurrent prints the active queue track.
func RunPlaylistCurrent() error {
	cfg, _, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("Aircoda is not configured yet. Run 'radio setup' to initialize")
		}
		return err
	}

	mgr := playlist.NewManager(cfg)
	track, err := mgr.Current()
	if err != nil {
		if strings.Contains(err.Error(), "no tracks available") {
			fmt.Printf("No supported audio files were found in:\n%s\n", cfg.Audio.Directory)
			return nil
		}
		return err
	}

	fmt.Println("Current track:")
	fmt.Println(track.Filename)
	return nil
}

// RunPlaylistNext advances the queue to next track and displays it.
func RunPlaylistNext() error {
	cfg, _, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("Aircoda is not configured yet. Run 'radio setup' to initialize")
		}
		return err
	}

	mgr := playlist.NewManager(cfg)
	track, err := mgr.Next()
	if err != nil {
		if strings.Contains(err.Error(), "no tracks available") {
			fmt.Printf("No supported audio files were found in:\n%s\n", cfg.Audio.Directory)
			return nil
		}
		return err
	}

	fmt.Println("Next track:")
	fmt.Println(track.Filename)
	return nil
}

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#00ADD8")
	ColorSecondary = lipgloss.Color("#3B82F6")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorWarning   = lipgloss.Color("#F59E0B")
	ColorError     = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorWhite     = lipgloss.Color("#FFFFFF")
	ColorDarkGray  = lipgloss.Color("#1F2937")

	// Styles
	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	StyleHeaderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 2).
			Align(lipgloss.Center)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSelectedOption = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary)

	StyleNormalOption = lipgloss.NewStyle().
				Foreground(ColorWhite)

	StyleDisabledOption = lipgloss.NewStyle().
				Foreground(ColorMuted)

	StyleBadgeLive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	StyleBadgeOffline = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorError)

	StyleBadgeWarning = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWarning)

	StyleLabel = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(16)

	StyleValue = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	StyleCheck = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	StyleCross = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)

// RenderAsciiBanner returns formatted ASCII banner centered with version subtitle.
func RenderAsciiBanner(subtitle string) string {
	ascii := "     _    _       ____          _       \n" +
		"    / \\  (_)_ __ / ___|___   __| | __ _ \n" +
		"   / _ \\ | | '__| |   / _ \\ / _` |/ _` |\n" +
		"  / ___ \\| | |  | |__| (_) | (_| | (_| |\n" +
		" /_/   \\_\\_|_|   \\____\\___/ \\__,_|\\__,_|"

	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(ascii)

	sub := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorMuted).
		Render(subtitle)

	boxContent := lipgloss.JoinVertical(lipgloss.Center, banner, sub)
	return StyleHeaderBox.Render(boxContent)
}

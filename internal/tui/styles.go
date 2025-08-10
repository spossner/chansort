package tui

import "github.com/charmbracelet/lipgloss"

// UI styling definitions using retro green terminal colors

var (
	// HeaderStyle for title text
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46")). // bright green
		PaddingBottom(1)

	// SelectedStyle for currently highlighted item
	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).   // black text
		Background(lipgloss.Color("46"))   // bright green background

	// NormalStyle for regular channel list items
	NormalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("34")) // darker green

	// ToastStyle for notification messages
	ToastStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).    // black text
		Background(lipgloss.Color("82")).   // bright lime green
		Padding(0, 1)

	// MoveModeStyle for move mode status
	MoveModeStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("82")) // bright lime green

	// StatusStyle for help text and status information
	StatusStyle = lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("28")) // darker green
)
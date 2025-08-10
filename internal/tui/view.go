package tui

import (
	"fmt"
	"strings"
)

// UI rendering functions

// View renders the complete UI
func (m Model) View() string {
	var b strings.Builder

	// Render header
	b.WriteString(HeaderStyle.Render("Samsung TV Channel List"))
	b.WriteString("\n\n")

	// Render column headers
	b.WriteString(fmt.Sprintf("   %-4s %-6s %s\n", "SLOT", "ID", "NAME"))
	b.WriteString(strings.Repeat("─", 53))
	b.WriteString("\n")

	// Render channel list
	b.WriteString(m.renderChannelList())

	// Render footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderChannelList renders the visible portion of the channel list
func (m Model) renderChannelList() string {
	var b strings.Builder

	// Calculate visible range for filtered channels
	start := m.ViewportTop
	end := m.ViewportTop + m.Height
	if end > len(m.FilteredChannels) {
		end = len(m.FilteredChannels)
	}

	// Display only visible filtered channels
	for i := start; i < end; i++ {
		channel := m.FilteredChannels[i]

		// Always show move column space, with indicator only in move mode
		moveIndicator := "   "
		if m.Mode == ModeMove && i == m.Cursor {
			moveIndicator = "▶︎  "
		}
		line := fmt.Sprintf("%s%-4d %-6d %s", moveIndicator, channel.OrderId, channel.ID, channel.Name)

		if i == m.Cursor {
			b.WriteString(SelectedStyle.Render(line))
		} else {
			b.WriteString(NormalStyle.Render(line))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	return b.String()
}

// renderFooter renders the status bar and toast messages
func (m Model) renderFooter() string {
	var b strings.Builder

	// Add scroll indicator and search info
	scrollInfo := ""
	if len(m.FilteredChannels) > m.Height {
		scrollInfo = fmt.Sprintf(" • %d/%d", m.Cursor+1, len(m.FilteredChannels))
	}

	// Show filter info if active
	filterInfo := ""
	if len(m.FilteredChannels) < len(m.Channels) {
		filterInfo = fmt.Sprintf(" • %d/%d matches", len(m.FilteredChannels), len(m.Channels))
	}

	// Show toast message if present
	if m.ToastMessage != "" {
		b.WriteString(ToastStyle.Render(m.ToastMessage))
	} else {
		// show standard status bar
		if m.Mode == ModeMove {
			helpText := "MOVE MODE: ↑/↓ to reorder • Enter to confirm • Esc to cancel"
			b.WriteString(MoveModeStyle.Render(helpText + scrollInfo))
		} else if m.SearchBuffer != "" {
			if filterInfo != "" {
				b.WriteString(StatusStyle.Render(fmt.Sprintf("Filtering: '%s'%s", m.SearchBuffer, filterInfo)))
			} else {
				b.WriteString(StatusStyle.Render(fmt.Sprintf("Searching: %s", m.SearchBuffer)))
			}
		} else {
			helpText := "Press ↑/↓ or j/k to navigate • type to filter • m to move • q to quit"
			b.WriteString(StatusStyle.Render(helpText + scrollInfo + filterInfo))
		}
	}

	return b.String()
}

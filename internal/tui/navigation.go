package tui

import (
	"github.com/spossner/chansort/internal/scm"
)

// Navigation and viewport management functions

// UpdateViewport ensures the cursor is visible within the viewport bounds
func (m Model) UpdateViewport() Model {
	if m.Height <= 0 {
		return m
	}

	// Ensure cursor is within viewport
	if m.Cursor < m.ViewportTop {
		m.ViewportTop = m.Cursor
	} else if m.Cursor >= m.ViewportTop+m.Height {
		m.ViewportTop = m.Cursor - m.Height + 1
	}

	// Ensure viewport doesn't go beyond bounds
	if m.ViewportTop < 0 {
		m.ViewportTop = 0
	}
	maxTop := len(m.FilteredChannels) - m.Height
	if maxTop < 0 {
		maxTop = 0
	}
	if m.ViewportTop > maxTop {
		m.ViewportTop = maxTop
	}

	return m
}

func (m Model) calculateNewCursorPosition(delta int) int {
	return min(max(0, m.Cursor+delta), len(m.FilteredChannels)-1)
}

func (m Model) MoveCursorBy(delta int) Model {
	newPosition := m.calculateNewCursorPosition(delta)
	if m.Cursor == newPosition {
		return m // no change needed
	}
	m.Cursor = newPosition
	m = m.UpdateViewport()
	return m
}

func (m Model) MoveChannelBy(delta int) Model {
	currentPosition := m.Cursor
	m = m.MoveCursorBy(delta)
	m = m.MoveChannel(currentPosition, m.Cursor)
	return m
}

// FindSelectedChannelInFullList finds the currently selected channel's position in the full channel list
func (m Model) FindSelectedChannelInFullList() int {
	if m.Cursor >= len(m.FilteredChannels) {
		return 0
	}

	selectedChannel := m.FilteredChannels[m.Cursor]

	for i, channel := range m.Channels {
		if channel.ID == selectedChannel.ID {
			return i
		}
	}

	return 0
}

// ResetToFullList clears filtering and returns to the full channel list
func (m Model) ResetToFullList() Model {
	// Find the currently selected channel before clearing filter
	var selectedChannel *scm.Channel
	if m.Cursor < len(m.FilteredChannels) {
		selectedChannel = &m.FilteredChannels[m.Cursor]
	}

	// Calculate current screen position
	screenPosition := m.Cursor - m.ViewportTop

	// Reset filter and clear search
	m.SearchBuffer = ""
	m.FilteredChannels = m.Channels
	m = m.ClearSearchTimer()

	// Find the selected channel's position in the full list
	newCursor := 0
	if selectedChannel != nil {
		for i, channel := range m.Channels {
			if channel.ID == selectedChannel.ID {
				newCursor = i
				break
			}
		}
	}

	// Set cursor and adjust viewport to maintain screen position
	m.Cursor = newCursor

	// Try to maintain the same screen position
	targetViewportTop := m.Cursor - screenPosition

	// Ensure viewport is within bounds
	maxViewportTop := len(m.Channels) - m.Height
	if maxViewportTop < 0 {
		maxViewportTop = 0
	}

	if targetViewportTop < 0 {
		m.ViewportTop = 0
	} else if targetViewportTop > maxViewportTop {
		m.ViewportTop = maxViewportTop
	} else {
		m.ViewportTop = targetViewportTop
	}

	m = m.UpdateViewport()
	return m
}

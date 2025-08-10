package tui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spossner/chansort/internal/scm"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles all input events and state changes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height - 7 // reserve space for header and footer
		m = m.UpdateViewport()

	case ClearSearchMsg:
		// Only clear if this is the current timer
		if msg.TimerID == m.SearchTimerID {
			m.SearchBuffer = ""
			m = m.ClearSearchTimer()
		}

	case ClearToastMsg:
		// Only clear if this is the current timer
		if msg.TimerID == m.ToastTimerID {
			m.ToastMessage = ""
			m = m.ClearToastTimer()
		}

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	if m.Mode == ModeSearch {
		switch msg.String() {
		case "esc":
			return m.handleResetToFullList()

		case "backspace":
			return m.handleBackspace()

		case "up":
			m = m.MoveUp()
		case "down":
			m = m.MoveDown()

		default:
			return m.handleCharacterInput(msg)
		}
	}

	if m.Mode == ModeMove {
		switch msg.String() {
		case "esc":
			return m.handleExitMoveMode()
		case "enter":
			return m.handleEnter()
		case "up":
			return m.MoveChannelUp(), nil
		case "down":
			return m.MoveChannelDown(), nil
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "m":
		return m.handleMoveMode()

	case "up":
		m = m.MoveUp()
	case "down":
		m = m.MoveDown()

	default:
		return m.handleCharacterInput(msg)
	}

	return m, nil
}

func (m Model) handleResetToFullList() (tea.Model, tea.Cmd) {
	// Reset filtering and return to full list
	m = m.ResetToFullList()
	m.Mode = ModeNav
	return m, nil
}

func (m Model) handleExitMoveMode() (tea.Model, tea.Cmd) {
	if m.Mode == ModeMove {
		// Exit move mode (keep current order)
		m.Mode = ModeNav
		m.MoveIndex = -1
		return m.ShowToast("Move mode exited")
	}
	return m, nil
}

// handleEnter processes Enter key press
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.Mode == ModeMove {
		// Exit move mode (keep current order)
		m.Mode = ModeNav
		m.MoveIndex = -1
		return m.ShowToast("Move confirmed")
	}
	return m, nil
}

// handleMoveMode processes 'm' key press to enter move mode
func (m Model) handleMoveMode() (tea.Model, tea.Cmd) {
	if m.Mode != ModeMove {
		// Enter move mode - first reset filtering and prepare state

		// Find the currently selected channel before clearing filter
		var selectedChannel *scm.Channel
		if m.Cursor < len(m.FilteredChannels) {
			selectedChannel = &m.FilteredChannels[m.Cursor]
		}

		// Clear filtering and reset to full list
		m.SearchBuffer = ""
		m.FilteredChannels = m.Channels
		m = m.ClearSearchTimer()

		// Find selected channel's position in the unfiltered list
		newCursor := 0
		if selectedChannel != nil {
			for i, channel := range m.Channels {
				if channel.ID == selectedChannel.ID && channel.OrderId == selectedChannel.OrderId {
					newCursor = i
					break
				}
			}
		}

		// Set move mode state
		m.Cursor = newCursor
		m.Mode = ModeMove
		m.MoveIndex = newCursor
		m = m.UpdateViewport()
		return m.ShowToast("Move mode: use ↑/↓ to reorder, Enter to confirm, Esc to exit")
	}
	return m, nil
}

// handleBackspace processes backspace key press
func (m Model) handleBackspace() (tea.Model, tea.Cmd) {
	if len(m.SearchBuffer) > 0 {
		m.SearchBuffer = m.SearchBuffer[:len(m.SearchBuffer)-1]
		m = m.ClearSearchTimer()

		var toastCmd tea.Cmd

		if m.SearchBuffer == "" {
			return m.handleResetToFullList()
		} else if m.IsNumeric(m.SearchBuffer) {
			// Still numeric after backspace - try to navigate to ID with timer
			id, _ := strconv.Atoi(m.SearchBuffer)
			newCursor, found := m.FindChannelByOrderId(id)

			if found {
				m.Cursor = newCursor
				m = m.UpdateViewport()
			} else {
				m, toastCmd = m.ShowToast(fmt.Sprintf("Channel %d not found", id))
			}

			// Use timer for numeric input
			m, searchCmd := m.StartSearchTimer()
			return m, tea.Batch(searchCmd, toastCmd)
		} else {
			// Now text mode - update filter (no timer, persistent filtering)
			m.FilteredChannels = m.FilterChannels(m.SearchBuffer)
			m.Cursor = 0
			m.ViewportTop = 0
			m = m.UpdateViewport()

			if len(m.FilteredChannels) == 0 {
				m, toastCmd = m.ShowToast(fmt.Sprintf("No matches for: '%s'", m.SearchBuffer))
			}

			return m, toastCmd
		}
	}
	return m, nil
}

// handleCharacterInput processes printable character input for search
func (m Model) handleCharacterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle printable characters for search
	if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
		m.Mode = ModeSearch
		m = m.ClearSearchTimer()
		m.SearchBuffer += msg.String()

		var toastCmd tea.Cmd

		// Check if search buffer is numeric - navigate to ID with timer
		if m.IsNumeric(m.SearchBuffer) {
			id, _ := strconv.Atoi(m.SearchBuffer)
			newCursor, found := m.FindChannelByOrderId(id)

			if found {
				m.Cursor = newCursor
				m = m.UpdateViewport()
			} else {
				m, toastCmd = m.ShowToast(fmt.Sprintf("Channel %d not found", id))
			}

			// Use timer for numeric input (allows multi-digit entry)
			m, searchCmd := m.StartSearchTimer()
			return m, tea.Batch(searchCmd, toastCmd)
		} else {
			// Text search - update filter (no timer, persistent filtering)
			m.FilteredChannels = m.FilterChannels(m.SearchBuffer)

			// Reset cursor to first match if available
			if len(m.FilteredChannels) > 0 {
				m.Cursor = 0
			} else {
				m.Cursor = 0
			}
			m.ViewportTop = 0
			m = m.UpdateViewport()

			if len(m.FilteredChannels) == 0 {
				m, toastCmd = m.ShowToast(fmt.Sprintf("No matches for: '%s'", m.SearchBuffer))
			}

			return m, toastCmd
		}
	}
	return m, nil
}

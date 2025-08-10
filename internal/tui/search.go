package tui

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spossner/chansort/internal/scm"
)

// Search and filter related functions

// IsNumeric checks if a string contains only digits
func (m Model) IsNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// FilterChannels returns channels that match the query string (case-insensitive)
func (m Model) FilterChannels(query string) []scm.Channel {
	if query == "" {
		return m.Channels
	}

	query = strings.ToLower(query)
	var filtered []scm.Channel

	for _, channel := range m.Channels {
		if strings.Contains(strings.ToLower(channel.Name), query) {
			filtered = append(filtered, channel)
		}
	}

	return filtered
}

// FindChannelByOrderId finds a channel by its OrderId and returns the index in filtered channels
func (m Model) FindChannelByOrderId(orderId int) (int, bool) {
	for i, channel := range m.FilteredChannels {
		if int(channel.OrderId) == orderId {
			return i, true
		}
	}
	return -1, false
}

// ClearSearchTimer stops the current search timer if it exists
func (m Model) ClearSearchTimer() Model {
	if m.SearchTimer != nil {
		m.SearchTimer.Stop()
	}
	return m
}

// StartSearchTimer creates a new search timer that will clear the buffer after 1800ms
func (m Model) StartSearchTimer() (Model, tea.Cmd) {
	m.SearchTimerID++
	timerID := m.SearchTimerID
	return m, tea.Tick(1800*time.Millisecond, func(t time.Time) tea.Msg {
		return ClearSearchMsg{TimerID: timerID}
	})
}
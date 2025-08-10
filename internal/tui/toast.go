package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Toast notification system functions

// ClearToastTimer stops the current toast timer if it exists
func (m Model) ClearToastTimer() Model {
	if m.ToastTimer != nil {
		m.ToastTimer.Stop()
	}
	return m
}

// ShowToast displays a toast message and sets up auto-clear timer
func (m Model) ShowToast(message string) (Model, tea.Cmd) {
	m = m.ClearToastTimer()
	m.ToastTimerID++
	timerID := m.ToastTimerID
	m.ToastMessage = message
	return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ClearToastMsg{TimerID: timerID}
	})
}
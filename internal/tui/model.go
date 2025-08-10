package tui

import (
	"time"

	"github.com/spossner/chansort/internal/scm"
)

const (
	ModeNav = iota
	ModeSearch
	ModeMove
)

// Model represents the application state
type Model struct {
	Channels         []scm.Channel
	FilteredChannels []scm.Channel
	Cursor           int
	ViewportTop      int
	Height           int
	Mode             int
	SearchBuffer     string
	SearchTimer      *time.Timer
	SearchTimerID    int
	ToastMessage     string
	ToastTimer       *time.Timer
	ToastTimerID     int
	MoveIndex        int
}

// Message types for Bubble Tea communication
type ClearSearchMsg struct {
	TimerID int
}

type ClearToastMsg struct {
	TimerID int
}

// NewModel creates a new model with the given channels
func NewModel(channels []scm.Channel) Model {
	return Model{
		Channels:         channels,
		FilteredChannels: channels, // initially show all channels
		Cursor:           0,
		ViewportTop:      0,
		Height:           25, // default height, will be updated on window size msg
		Mode:             ModeNav,
		SearchBuffer:     "",
		SearchTimer:      nil,
		SearchTimerID:    0,
		ToastMessage:     "",
		ToastTimer:       nil,
		ToastTimerID:     0,
		MoveIndex:        -1,
	}
}

package tui

import (
	"log"
	"time"

	"github.com/spossner/chansort/internal/scm"
)

const (
	ModeNav = iota
	ModeSearch
	ModeJump
	ModeMove
)

// Model represents the application state
type Model struct {
	Path             string
	Recs             []scm.Channel
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
func NewModel(path string, recs []scm.Channel) Model {
	channels := scm.SortTVOrder(recs)
	if len(channels) == 0 {
		log.Fatal("no channels found in the SCM file")
	}
	return Model{
		Path:             path, // the original file path
		Recs:             recs, // original channels read (all)
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

func (m Model) MoveChannel(position, newPosition int) Model {
	if position < 0 || position >= len(m.Channels) || newPosition < 0 || newPosition >= len(m.Channels) {
		return m // invalid positions
	}

	if position == newPosition {
		return m // no change needed
	}

	// Move the channel in the slice
	channel := m.Channels[position]
	m.Channels = append(m.Channels[:position], m.Channels[position+1:]...)
	m.Channels = append(m.Channels[:newPosition], append([]scm.Channel{channel}, m.Channels[newPosition:]...)...)

	// refresh the OrderID
	for i := range m.Channels {
		m.Channels[i].OrderId = uint16(i + 1)
	}

	// Update filtered channels if necessary
	m.FilteredChannels = m.FilterChannels(m.SearchBuffer)

	return m
}

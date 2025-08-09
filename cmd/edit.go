package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spossner/chansort/internal/scm"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Interactive channel browser with keyboard navigation",
	RunE: func(cmd *cobra.Command, args []string) error {
		recs, err := scm.ReadSatelliteRecords(scmPath)
		if err != nil {
			return err
		}

		channels := scm.SortTVOrder(recs)
		if len(channels) == 0 {
			return fmt.Errorf("no channels found in the SCM file")
		}

		p := tea.NewProgram(initialModel(channels))
		_, err = p.Run()
		return err
	},
}

type model struct {
	channels         []scm.Channel
	filteredChannels []scm.Channel
	cursor           int
	viewportTop      int
	height           int
	searchBuffer     string
	searchTimer      *time.Timer
	searchTimerID    int
	toastMessage     string
	toastTimer       *time.Timer
	toastTimerID     int
	moveMode         bool
	moveIndex        int
	originalChannels []scm.Channel
	originalCursor   int
}

type clearSearchMsg struct {
	timerID int
}
type clearToastMsg struct {
	timerID int
}

func initialModel(channels []scm.Channel) model {
	return model{
		channels:         channels,
		filteredChannels: channels, // initially show all channels
		cursor:           0,
		viewportTop:      0,
		height:           25, // default height, will be updated on window size msg
		searchBuffer:     "",
		searchTimer:      nil,
		searchTimerID:    0,
		toastMessage:     "",
		toastTimer:       nil,
		toastTimerID:     0,
		moveMode:         false,
		moveIndex:        -1,
		originalChannels: nil,
		originalCursor:   0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height - 5 // reserve space for header and footer
		m = m.updateViewport()
	case clearSearchMsg:
		// Only clear if this is the current timer
		if msg.timerID == m.searchTimerID {
			m.searchBuffer = ""
			m = m.clearSearchTimer()
		}
	case clearToastMsg:
		// Only clear if this is the current timer
		if msg.timerID == m.toastTimerID {
			m.toastMessage = ""
			m = m.clearToastTimer()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.moveMode {
				// Abort move mode - restore original state
				m.filteredChannels = m.originalChannels
				m.channels = m.originalChannels
				m.cursor = m.originalCursor
				m.moveMode = false
				m.moveIndex = -1
				m.originalChannels = nil
				m = m.updateViewport()
				return m.showToast("Move cancelled")
			}

			// Find the currently selected channel in the full list
			var selectedChannel *scm.Channel
			if m.cursor < len(m.filteredChannels) {
				selectedChannel = &m.filteredChannels[m.cursor]
			}

			// Calculate current screen position
			screenPosition := m.cursor - m.viewportTop

			// Reset filter and clear search
			m.searchBuffer = ""
			m.filteredChannels = m.channels
			m = m.clearSearchTimer()

			// Find the selected channel's position in the full list
			newCursor := 0
			if selectedChannel != nil {
				for i, channel := range m.channels {
					if channel.ID == selectedChannel.ID && channel.OrderId == selectedChannel.OrderId {
						newCursor = i
						break
					}
				}
			}

			// Set cursor and adjust viewport to maintain screen position
			m.cursor = newCursor

			// Try to maintain the same screen position
			targetViewportTop := m.cursor - screenPosition

			// Ensure viewport is within bounds
			maxViewportTop := len(m.channels) - m.height
			if maxViewportTop < 0 {
				maxViewportTop = 0
			}

			if targetViewportTop < 0 {
				m.viewportTop = 0
			} else if targetViewportTop > maxViewportTop {
				m.viewportTop = maxViewportTop
			} else {
				m.viewportTop = targetViewportTop
			}

			m = m.updateViewport()
		case "enter":
			if m.moveMode {
				// Confirm move mode - keep current state
				m.moveMode = false
				m.moveIndex = -1
				m.originalChannels = nil
				return m.showToast("Move confirmed")
			}
		case "m":
			if !m.moveMode {
				// Enter move mode - first reset filtering and prepare state

				// Find the currently selected channel before clearing filter
				var selectedChannel *scm.Channel
				if m.cursor < len(m.filteredChannels) {
					selectedChannel = &m.filteredChannels[m.cursor]
				}

				// Save original state (full channels list for potential reset)
				m.originalChannels = make([]scm.Channel, len(m.channels))
				copy(m.originalChannels, m.channels)

				// Clear filtering and reset to full list
				m.searchBuffer = ""
				m.filteredChannels = m.channels
				m = m.clearSearchTimer()

				// Find selected channel's position in the unfiltered list
				newCursor := 0
				if selectedChannel != nil {
					for i, channel := range m.channels {
						if channel.ID == selectedChannel.ID && channel.OrderId == selectedChannel.OrderId {
							newCursor = i
							break
						}
					}
				}

				// Set move mode state
				m.cursor = newCursor
				m.originalCursor = newCursor
				m.moveMode = true
				m.moveIndex = newCursor
				m = m.updateViewport()
			}
		case "up", "k":
			if m.moveMode {
				if m.cursor > 0 {
					// Swap current item with the one above
					m.filteredChannels[m.cursor], m.filteredChannels[m.cursor-1] = m.filteredChannels[m.cursor-1], m.filteredChannels[m.cursor]
					m.channels[m.cursor], m.channels[m.cursor-1] = m.channels[m.cursor-1], m.channels[m.cursor]
					m.cursor--
					m = m.updateViewport()
				}
			} else {
				if m.cursor > 0 {
					m.cursor--
					m = m.updateViewport()
				}
			}
		case "down", "j":

			if m.cursor >= len(m.filteredChannels)-1 {
				break
			}

			if m.moveMode {
				// Swap current item with the one below
				m.filteredChannels[m.cursor], m.filteredChannels[m.cursor+1] = m.filteredChannels[m.cursor+1], m.filteredChannels[m.cursor]
				m.channels[m.cursor], m.channels[m.cursor+1] = m.channels[m.cursor+1], m.channels[m.cursor]
				m.cursor++
				m = m.updateViewport()
			} else {
				m.cursor++
				m = m.updateViewport()
			}
		case "backspace":
			if len(m.searchBuffer) > 0 {
				m.searchBuffer = m.searchBuffer[:len(m.searchBuffer)-1]
				m = m.clearSearchTimer()

				var toastCmd tea.Cmd

				if m.searchBuffer == "" {
					// Empty search buffer - reset to full list
					m.filteredChannels = m.channels
					m.cursor = 0
					m.viewportTop = 0
					m = m.updateViewport()
				} else if m.isNumeric(m.searchBuffer) {
					// Still numeric after backspace - try to navigate to ID
					id, _ := strconv.Atoi(m.searchBuffer)
					newCursor, found := m.findChannelByOrderId(id)

					if found {
						m.cursor = newCursor
						m = m.updateViewport()
					} else {
						m, toastCmd = m.showToast(fmt.Sprintf("Channel %d not found", id))
					}
				} else {
					// Now text mode - update filter
					m.filteredChannels = m.filterChannels(m.searchBuffer)
					m.cursor = 0
					m.viewportTop = 0
					m = m.updateViewport()

					if len(m.filteredChannels) == 0 {
						m, toastCmd = m.showToast(fmt.Sprintf("No matches for: '%s'", m.searchBuffer))
					}
				}

				if m.searchBuffer != "" {
					m, searchCmd := m.startSearchTimer()
					return m, tea.Batch(searchCmd, toastCmd)
				}

				return m, toastCmd
			}
		default:
			// Handle printable characters for search
			if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
				m = m.clearSearchTimer()
				m.searchBuffer += msg.String()

				var toastCmd tea.Cmd

				// Check if search buffer is numeric - navigate to ID
				if m.isNumeric(m.searchBuffer) {
					id, _ := strconv.Atoi(m.searchBuffer)
					newCursor, found := m.findChannelByOrderId(id)

					if found {
						m.cursor = newCursor
						m = m.updateViewport()
					} else {
						m, toastCmd = m.showToast(fmt.Sprintf("Channel %d not found", id))
					}
				} else {
					// Text search - update filter
					m.filteredChannels = m.filterChannels(m.searchBuffer)

					// Reset cursor to first match if available
					if len(m.filteredChannels) > 0 {
						m.cursor = 0
					} else {
						m.cursor = 0
					}
					m.viewportTop = 0
					m = m.updateViewport()

					if len(m.filteredChannels) == 0 {
						m, toastCmd = m.showToast(fmt.Sprintf("No matches for: '%s'", m.searchBuffer))
					}
				}

				m, searchCmd := m.startSearchTimer()
				return m, tea.Batch(searchCmd, toastCmd)
			}
		}
	}
	return m, nil
}

func (m model) updateViewport() model {
	if m.height <= 0 {
		return m
	}

	// Ensure cursor is within viewport
	if m.cursor < m.viewportTop {
		m.viewportTop = m.cursor
	} else if m.cursor >= m.viewportTop+m.height {
		m.viewportTop = m.cursor - m.height + 1
	}

	// Ensure viewport doesn't go beyond bounds
	if m.viewportTop < 0 {
		m.viewportTop = 0
	}
	maxTop := len(m.channels) - m.height
	if maxTop < 0 {
		maxTop = 0
	}
	if m.viewportTop > maxTop {
		m.viewportTop = maxTop
	}

	return m
}

func (m model) isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (m model) findChannelByOrderId(id int) (int, bool) {
	for i, channel := range m.filteredChannels {
		if int(channel.OrderId) == id {
			return i, true
		}
	}
	return -1, false
}

func (m model) filterChannels(query string) []scm.Channel {
	if query == "" {
		return m.channels
	}

	query = strings.ToLower(query)
	var filtered []scm.Channel

	for _, channel := range m.channels {
		if strings.Contains(strings.ToLower(channel.Name), query) {
			filtered = append(filtered, channel)
		}
	}

	return filtered
}

func (m model) clearSearchTimer() model {
	if m.searchTimer != nil {
		m.searchTimer.Stop()
	}
	return m
}

func (m model) startSearchTimer() (model, tea.Cmd) {
	m.searchTimerID++
	timerID := m.searchTimerID
	return m, tea.Tick(1800*time.Millisecond, func(t time.Time) tea.Msg {
		return clearSearchMsg{timerID: timerID}
	})
}

func (m model) clearToastTimer() model {
	if m.toastTimer != nil {
		m.toastTimer.Stop()
	}
	return m
}

func (m model) showToast(message string) (model, tea.Cmd) {
	m = m.clearToastTimer()
	m.toastTimerID++
	timerID := m.toastTimerID
	m.toastMessage = message
	return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearToastMsg{timerID: timerID}
	})
}

func (m model) View() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46")). // bright green
		PaddingBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")). // black text
		Background(lipgloss.Color("46")) // bright green background

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("34")) // darker green

	b.WriteString(headerStyle.Render("Samsung TV Channel List"))
	b.WriteString("\n\n")

	// Header - always include space for move column
	b.WriteString(fmt.Sprintf("   %-4s %-6s %s\n", "SLOT", "ID", "NAME"))
	b.WriteString(strings.Repeat("─", 53))
	b.WriteString("\n")

	// Calculate visible range for filtered channels
	start := m.viewportTop
	end := m.viewportTop + m.height
	if end > len(m.filteredChannels) {
		end = len(m.filteredChannels)
	}

	// Display only visible filtered channels
	for i := start; i < end; i++ {
		channel := m.filteredChannels[i]

		// Always show move column space, with indicator only in move mode
		moveIndicator := "   "
		if m.moveMode && i == m.cursor {
			moveIndicator = "▶︎  "
		}
		line := fmt.Sprintf("%s%-4d %-6d %s", moveIndicator, channel.OrderId, channel.ID, channel.Name)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Add scroll indicator and search info
	scrollInfo := ""
	if len(m.filteredChannels) > m.height {
		scrollInfo = fmt.Sprintf(" • %d/%d", m.cursor+1, len(m.filteredChannels))
	}

	// Show filter info if active
	filterInfo := ""
	if len(m.filteredChannels) < len(m.channels) {
		filterInfo = fmt.Sprintf(" • %d/%d matches", len(m.filteredChannels), len(m.channels))
	}

	// Show toast message if present
	if m.toastMessage != "" {
		toastStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).  // black text
			Background(lipgloss.Color("82")). // bright lime green
			Padding(0, 1)
		b.WriteString(toastStyle.Render(m.toastMessage))
	}
	b.WriteString("\n")

	// show standard status bar
	if m.moveMode {
		helpText := "MOVE MODE: ↑/↓ to reorder • Enter to confirm • Esc to cancel"
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Render(helpText + scrollInfo))
	} else if m.searchBuffer != "" {
		if filterInfo != "" {
			b.WriteString(lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("28")).Render(fmt.Sprintf("Filtering: '%s'%s", m.searchBuffer, filterInfo)))
		} else {
			b.WriteString(lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("28")).Render(fmt.Sprintf("Searching: %s", m.searchBuffer)))
		}
	} else {
		helpText := "Press ↑/↓ or j/k to navigate • type to filter • m to move • ESC to reset • n for next match • q to quit"
		b.WriteString(lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("28")).Render(helpText + scrollInfo + filterInfo))
	}

	return b.String()
}

func init() {
	rootCmd.AddCommand(editCmd)
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spossner/chansort/internal/scm"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
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
	channels         []scm.Record
	filteredChannels []scm.Record
	cursor           int
	viewportTop      int
	height           int
	searchBuffer     string
	searchTimer      *time.Timer
	searchTimerID    int
	lastSearchTerm   string
	toastMessage     string
	toastTimer       *time.Timer
	toastTimerID     int
}

type clearSearchMsg struct {
	timerID int
}
type clearToastMsg struct {
	timerID int
}

func initialModel(channels []scm.Record) model {
	return model{
		channels:         channels,
		filteredChannels: channels, // initially show all channels
		cursor:           0,
		viewportTop:      0,
		height:           25, // default height, will be updated on window size msg
		searchBuffer:     "",
		searchTimer:      nil,
		searchTimerID:    0,
		lastSearchTerm:   "",
		toastMessage:     "",
		toastTimer:       nil,
		toastTimerID:     0,
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
			// Save search term before clearing buffer
			if m.searchBuffer != "" {
				m.lastSearchTerm = m.searchBuffer
			}
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
		case "escape":
			// Reset filter and clear search
			m.searchBuffer = ""
			m.filteredChannels = m.channels
			m.cursor = 0
			m.viewportTop = 0
			m = m.clearSearchTimer()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m = m.updateViewport()
			}
		case "down", "j":
			if m.cursor < len(m.filteredChannels)-1 {
				m.cursor++
				m = m.updateViewport()
			}
		case "backspace":
			if len(m.searchBuffer) > 0 {
				m.searchBuffer = m.searchBuffer[:len(m.searchBuffer)-1]
				m = m.clearSearchTimer()
				
				// Update filter and reset cursor
				m.filteredChannels = m.filterChannels(m.searchBuffer)
				m.cursor = 0
				m.viewportTop = 0
				m = m.updateViewport()
				
				if m.searchBuffer != "" {
					return m.startSearchTimer()
				}
			}
		case "n":
			if m.lastSearchTerm == "" {
				return m.showToast("No previous search term")
			}
			newCursor, ok := m.searchNext(m.lastSearchTerm, m.cursor+1)
			if !ok {
				return m.showToast("No more matches for '" + m.lastSearchTerm + "'")
			}
			wrapped := newCursor < m.cursor
			m.cursor = newCursor
			m = m.updateViewport()
			if wrapped {
				return m.showToast("Wrapped...")
			}
		default:
			// Handle printable characters for search
			if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
				m = m.clearSearchTimer()
				m.searchBuffer += msg.String()

				// Update filter
				m.filteredChannels = m.filterChannels(m.searchBuffer)
				
				// Reset cursor to first match if available
				if len(m.filteredChannels) > 0 {
					m.cursor = 0
				} else {
					m.cursor = 0
				}
				m.viewportTop = 0
				m = m.updateViewport()

				var toastCmd tea.Cmd
				if len(m.filteredChannels) == 0 {
					m, toastCmd = m.showToast(fmt.Sprintf("No matches for: '%s'", m.searchBuffer))
				}

				// Save search term when buffer is cleared
				m.lastSearchTerm = m.searchBuffer
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

func (m model) filterChannels(query string) []scm.Record {
	if query == "" {
		return m.channels
	}

	query = strings.ToLower(query)
	var filtered []scm.Record

	for _, channel := range m.channels {
		if strings.Contains(strings.ToLower(channel.Name), query) {
			filtered = append(filtered, channel)
		}
	}

	return filtered
}

func (m model) searchNext(query string, start int) (int, bool) {
	if query == "" {
		return m.cursor, false
	}

	query = strings.ToLower(query)
	channels := m.filteredChannels

	// Search from start position to end
	for i := start; i < len(channels); i++ {
		if strings.Contains(strings.ToLower(channels[i].Name), query) {
			return i, true
		}
	}

	// Wrap around: search from beginning to current position
	for i := 0; i < m.cursor; i++ {
		if strings.Contains(strings.ToLower(channels[i].Name), query) {
			return i, true
		}
	}

	// No match found, stay at current position
	return m.cursor, false
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
		Foreground(lipgloss.Color("205")).
		PaddingBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Background(lipgloss.Color("57"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	b.WriteString(headerStyle.Render("Samsung TV Channel List"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%-4s %-6s %s\n", "LCN", "SLOT", "NAME"))
	b.WriteString(strings.Repeat("─", 50))
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
		line := fmt.Sprintf("%-4d %-6d %s", channel.LCN, channel.SlotIndex, channel.Name)

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
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("208")).
			Padding(0, 1)
		b.WriteString(toastStyle.Render(m.toastMessage))
	}
	b.WriteString("\n")

	// show standard status bar
	if m.searchBuffer != "" {
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Filtering: '%s'%s", m.searchBuffer, filterInfo)))
	} else {
		helpText := "Press ↑/↓ or j/k to navigate • type to filter • ESC to reset • n for next match • q to quit"
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(helpText + scrollInfo + filterInfo))
	}

	return b.String()
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

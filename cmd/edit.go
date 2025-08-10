package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spossner/chansort/internal/scm"
	"github.com/spossner/chansort/internal/tui"
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

		p := tea.NewProgram(tui.NewModel(channels))
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}


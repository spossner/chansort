package cmd

import (
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

		p := tea.NewProgram(tui.NewModel(scmPath, recs))
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

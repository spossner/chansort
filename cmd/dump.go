package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spossner/chansort/internal/scm"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump channels in the TV's current order",
	RunE: func(cmd *cobra.Command, args []string) error {
		recs, err := scm.ReadSatelliteRecords(scmPath)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "LCN\tSLOT\tNAME")
		for _, r := range scm.SortTVOrder(recs) {
			fmt.Fprintf(w, "%d\t%d\t%s\n", r.LCN, r.SlotIndex, r.Name)
		}
		return w.Flush()
	},
}

func init() { rootCmd.AddCommand(dumpCmd) }

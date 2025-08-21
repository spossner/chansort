package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spossner/chansort/internal/scm"
)

var swapCmd = &cobra.Command{
	Use:   "swap",
	Short: "Swaps ARD to 2 and ZDF to 1",
	RunE: func(cmd *cobra.Command, args []string) error {
		recs, err := scm.ReadSatelliteRecords(scmPath)
		if err != nil {
			return err
		}

		channels := scm.SortTVOrder(recs)

		// reverse first 10 items
		for i := range 10 {
			channels[i].OrderId = uint16(10 - i)
		}

		// that should update the underlying Raw slice (which is identical to the one in recs...)
		scm.WriteOrderIdsBackIntoChannelData(channels)

		scm.WriteSatelliteRecords(scmPath, recs, scmPath+"_shuffled")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(swapCmd)
}

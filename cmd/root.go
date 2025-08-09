package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var scmPath string

var rootCmd = &cobra.Command{
	Use:   "chansort",
	Short: "Samsung SCM tools (satellite) — dump and, later, reorder",
	Long:  "Chansort reads Samsung .scm archives and lets you inspect satellite channel lists.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&scmPath, "scm", "s", "", "Path to Samsung .scm archive (exported from TV)")
	rootCmd.MarkPersistentFlagRequired("scm")
}

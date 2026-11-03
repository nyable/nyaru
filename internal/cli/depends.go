package cli

import (
	"github.com/spf13/cobra"
)

var dependsCmd = &cobra.Command{
	Use:   "depends <app>",
	Short: "列出应用依赖",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("depends", args[0])
	},
}

func init() {
	rootCmd.AddCommand(dependsCmd)
}

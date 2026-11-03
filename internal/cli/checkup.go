package cli

import (
	"github.com/spf13/cobra"
)

var checkupCmd = &cobra.Command{
	Use:     "checkup",
	Short:   "检查 Scoop 环境中的潜在问题",
	Args:    cobra.NoArgs,
	Aliases: []string{"doctor"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCommand(scoopCommand("checkup"))
	},
}

func init() {
	rootCmd.AddCommand(checkupCmd)
}

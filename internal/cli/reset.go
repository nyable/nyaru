package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var resetAll bool

var resetCmd = &cobra.Command{
	Use:   "reset [app...]",
	Short: "重置应用以解决冲突",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if resetAll && len(args) > 0 {
			return fmt.Errorf("--all 不能与具体应用同时使用")
		}
		if !resetAll && len(args) == 0 {
			return fmt.Errorf("请指定应用名，或使用 --all 重置全部应用")
		}
		scoopArgs := []string{"reset"}
		if resetAll {
			scoopArgs = append(scoopArgs, "--all")
		} else {
			scoopArgs = append(scoopArgs, args...)
		}
		if !confirm("确认重置应用？") {
			return nil
		}
		return runCommand(scoopCommand(scoopArgs...))
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&resetAll, "all", "a", false, "重置全部应用")
	rootCmd.AddCommand(resetCmd)
}

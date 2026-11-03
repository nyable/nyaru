package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scoopCmd = &cobra.Command{
	Use:                "scoop <args...>",
	Short:              "直接透传参数给 scoop",
	Long:               `把参数原样传给 scoop，用于覆盖 nyaru 尚未封装的 Scoop 子命令。`,
	Example:            `  nyaru scoop bucket known`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		for len(args) > 0 && args[0] == "--" {
			args = args[1:]
		}
		if len(args) == 0 {
			return fmt.Errorf("请在 scoop 后提供参数")
		}
		return runCommand(scoopCommand(args...))
	},
}

func init() {
	rootCmd.AddCommand(scoopCmd)
}

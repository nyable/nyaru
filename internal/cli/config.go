package cli

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "读取或修改 Scoop 配置",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("config")
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "读取一个 Scoop 配置项",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("config", args[0])
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置一个 Scoop 配置项",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCommand(scoopCommand("config", args[0], args[1]))
	},
}

var configRemoveCmd = &cobra.Command{
	Use:     "rm <key>",
	Short:   "删除一个 Scoop 配置项",
	Aliases: []string{"remove"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCommand(scoopCommand("config", "rm", args[0]))
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出全部 Scoop 配置",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("config")
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}

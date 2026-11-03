package cli

import (
	"github.com/spf13/cobra"
)

var whichCmd = &cobra.Command{
	Use:   "which <command>",
	Short: "查找 Scoop 管理的可执行文件",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("which", args[0])
	},
}

var prefixCmd = &cobra.Command{
	Use:   "prefix <app>",
	Short: "显示应用安装目录",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("prefix", args[0])
	},
}

var catCmd = &cobra.Command{
	Use:   "cat <app>",
	Short: "显示应用的 manifest 内容",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("cat", args[0])
	},
}

var homeCmd = &cobra.Command{
	Use:   "home <app>",
	Short: "在浏览器中打开应用主页",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCommand(scoopCommand("home", args[0]))
	},
}

func init() {
	rootCmd.AddCommand(whichCmd)
	rootCmd.AddCommand(prefixCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(homeCmd)
}

package cli

import (
	"fmt"
	"os"

	"github.com/nyable/nyaru/internal/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:           "nyaru",
	Short:         "Scoop 的交互式 CLI 包装器",
	Long:          `Nyaru 用统一的交互界面包装 scoop 和 sfsu 命令。`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if flagNoColor {
			_ = os.Setenv("NO_COLOR", "1")
		}
		_, err := config.Load()
		return err
	},
}

func Execute() {
	if len(os.Args) == 1 {
		if !isTTY() {
			_ = rootCmd.Help()
			return
		}
		os.Args = append(os.Args, "interactive")
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = AppVersion
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "以 JSON 输出（只适用于查询命令）")
	rootCmd.PersistentFlags().BoolVar(&flagNoTUI, "no-tui", false, "不进入 TUI，输出纯文本")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "跳过确认")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "关闭颜色输出")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "输出更多诊断信息")
}

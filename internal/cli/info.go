package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info <app>",
	Short:   "查看应用详细信息(别名:detail)",
	Long:    `查看指定应用的详细信息，包括名称、版本、描述、主页等。`,
	Example: `nyaru info git`,
	Aliases: []string{"detail"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		InfoAction(args[0])
	},
}

func InfoAction(appName string) {
	pm := core.GetManager(config.GetActiveMode())

	res, err := tui.RunWithSpinner("正在获取应用信息...", func() (any, error) {
		return pm.Info(appName)
	})

	if err != nil {
		tui.PrintError(fmt.Sprintf("获取应用信息出错:\n%v", err))
		return
	}

	rawContent := res.(string)
	if rawContent == "" {
		tui.PrintWarning("未找到该应用的信息！")
		return
	}

	fmt.Println()
	fmt.Println(tui.FormatInfoContent(rawContent))
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

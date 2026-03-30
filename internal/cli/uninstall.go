package cli

import (
	"fmt"
	"strings"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall <app> [app...]",
	Short:   "卸载指定的应用(别名:rm/remove)",
	Long:    `通过 scoop 卸载一个或多个应用程序。`,
	Example: `  nyaru uninstall git
  nyaru uninstall aria2 curl wget`,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		UninstallAction(args)
	},
}

func UninstallAction(apps []string) {
	pm := core.GetManager(config.GetActiveMode())

	if len(apps) > 1 {
		fmt.Printf("⚠ 确认卸载这 %d 个应用程序吗？(%s) (y/n): ", len(apps), strings.Join(apps, ", "))
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			tui.PrintInfo("已取消卸载")
			return
		}
	}

	for _, app := range apps {
		tui.PrintInfo(fmt.Sprintf("正在卸载: %s", app))
		if err := pm.Uninstall(app); err != nil {
			tui.PrintError(fmt.Sprintf("卸载 %s 失败: %v", app, err))
		} else {
			tui.PrintSuccess(fmt.Sprintf("卸载 %s 成功!", app))
		}
	}
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

package cli

import (
	"fmt"
	"strings"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install <app> [app...]",
	Short:   "安装指定的应用(别名:add/i)",
	Long:    `通过 scoop 安装一个或多个应用程序。`,
	Example: `  nyaru install git
  nyaru install aria2 curl wget`,
	Aliases: []string{"add", "i"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		InstallAction(args)
	},
}

func InstallAction(apps []string) {
	pm := core.GetManager(config.GetActiveMode())

	if len(apps) > 1 {
		fmt.Printf("⚠ 确认安装这 %d 个应用程序吗？(%s) (y/n): ", len(apps), strings.Join(apps, ", "))
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			tui.PrintInfo("已取消安装")
			return
		}
	}

	for _, app := range apps {
		tui.PrintInfo(fmt.Sprintf("正在安装: %s", app))
		if err := pm.Install(app); err != nil {
			tui.PrintError(fmt.Sprintf("安装 %s 失败: %v", app, err))
		} else {
			tui.PrintSuccess(fmt.Sprintf("安装 %s 成功!", app))
		}
	}
}

func init() {
	rootCmd.AddCommand(installCmd)
}

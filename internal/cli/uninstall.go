package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var uninstallOpts core.UninstallOptions

var uninstallCmd = &cobra.Command{
	Use:     "uninstall <app> [app...]",
	Short:   "卸载指定的应用",
	Long:    `通过 scoop 卸载一个或多个应用程序。`,
	Example: `  nyaru uninstall git`,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return UninstallAction(args)
	},
}

func UninstallAction(apps []string) error {
	if !confirm(fmt.Sprintf("确认卸载这 %d 个应用？(%s)", len(apps), joinApps(apps))) {
		tui.PrintInfo("已取消卸载")
		return nil
	}
	pm := newBackend()
	cmd := pm.Uninstall(apps, uninstallOpts)
	tui.PrintInfo("正在卸载: " + joinApps(apps))
	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("卸载失败: %w", err)
	}
	tui.PrintSuccess("卸载完成")
	return nil
}

func init() {
	uninstallCmd.Flags().BoolVarP(&uninstallOpts.Global, "global", "g", false, "卸载全局应用")
	uninstallCmd.Flags().BoolVarP(&uninstallOpts.Purge, "purge", "p", false, "同时删除持久化数据")
	rootCmd.AddCommand(uninstallCmd)
}

package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var unholdCmd = &cobra.Command{
	Use:     "unhold <app> [app...]",
	Short:   "解锁应用，允许其被更新(别名:unpin)",
	Long:    `解锁指定的应用，使其在执行 scoop update 时恢复正常更新。`,
	Example: `  nyaru unhold git
  nyaru unhold aria2 curl`,
	Aliases: []string{"unpin"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		UnholdAction(args)
	},
}

func UnholdAction(apps []string) {
	pm := core.GetManager(config.GetActiveMode())
	for _, app := range apps {
		tui.PrintInfo(fmt.Sprintf("正在解锁: %s", app))
		if err := pm.Unhold(app); err != nil {
			tui.PrintError(fmt.Sprintf("解锁 %s 失败: %v", app, err))
		} else {
			tui.PrintSuccess(fmt.Sprintf("解锁 %s 成功!", app))
		}
	}
}

func init() {
	rootCmd.AddCommand(unholdCmd)
}

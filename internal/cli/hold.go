package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var holdCmd = &cobra.Command{
	Use:     "hold <app> [app...]",
	Short:   "锁定应用，阻止其被更新(别名:pin)",
	Long:    `锁定指定的应用，使其在执行 scoop update 时不会被更新。`,
	Example: `  nyaru hold git
  nyaru hold aria2 curl`,
	Aliases: []string{"pin"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		HoldAction(args)
	},
}

func HoldAction(apps []string) {
	pm := core.GetManager(config.GetActiveMode())
	for _, app := range apps {
		tui.PrintInfo(fmt.Sprintf("正在锁定: %s", app))
		if err := pm.Hold(app); err != nil {
			tui.PrintError(fmt.Sprintf("锁定 %s 失败: %v", app, err))
		} else {
			tui.PrintSuccess(fmt.Sprintf("锁定 %s 成功!", app))
		}
	}
}

func init() {
	rootCmd.AddCommand(holdCmd)
}

package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var unholdOpts core.HoldOptions

var unholdCmd = &cobra.Command{
	Use:     "unhold <app> [app...]",
	Short:   "解锁应用，允许其被更新",
	Aliases: []string{"unpin"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return UnholdAction(args)
	},
}

func UnholdAction(apps []string) error {
	pm := newBackend()
	tui.PrintInfo("正在解锁: " + joinApps(apps))
	if err := runCommand(pm.Unhold(apps, unholdOpts)); err != nil {
		return fmt.Errorf("解锁失败: %w", err)
	}
	tui.PrintSuccess("解锁完成")
	return nil
}

func init() {
	unholdCmd.Flags().BoolVarP(&unholdOpts.Global, "global", "g", false, "解锁全局应用")
	rootCmd.AddCommand(unholdCmd)
}

package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var holdOpts core.HoldOptions

var holdCmd = &cobra.Command{
	Use:     "hold <app> [app...]",
	Short:   "锁定应用，阻止其被更新",
	Aliases: []string{"pin"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return HoldAction(args)
	},
}

func HoldAction(apps []string) error {
	pm := newBackend()
	tui.PrintInfo("正在锁定: " + joinApps(apps))
	if err := runCommand(pm.Hold(apps, holdOpts)); err != nil {
		return fmt.Errorf("锁定失败: %w", err)
	}
	tui.PrintSuccess("锁定完成")
	return nil
}

func init() {
	holdCmd.Flags().BoolVarP(&holdOpts.Global, "global", "g", false, "锁定全局应用")
	rootCmd.AddCommand(holdCmd)
}

package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var cleanupOpts core.CleanupOptions

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [app...]",
	Short: "清理应用的旧版本",
	Long:  `删除已安装应用的旧版本目录；不带参数时在交互模式打开旧版本列表，使用 * 或 --all 清理全部应用。`,
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return CleanupAction(args)
	},
}

func CleanupAction(apps []string) error {
	if cleanupOpts.All && len(apps) > 0 {
		return fmt.Errorf("--all 不能与具体应用同时使用")
	}
	if len(apps) == 0 && !cleanupOpts.All && !cleanupOpts.Global && !cleanupOpts.Cache && tuiEnabled() {
		return runTUI(tui.Options{
			Backend: newBackend(),
			Start:   tui.StartCleanup,
			Version: AppVersion,
		})
	}
	if cleanupOpts.All && len(apps) == 0 {
		if !confirm("确认清理全部应用的旧版本？") {
			tui.PrintInfo("已取消清理")
			return nil
		}
	} else if len(apps) == 0 {
		return fmt.Errorf("请指定应用名，或使用 --all 清理全部应用")
	} else if len(apps) > 1 && configConfirmBatch() {
		if !confirm(fmt.Sprintf("确认清理这 %d 个应用的旧版本？(%s)", len(apps), joinApps(apps))) {
			tui.PrintInfo("已取消清理")
			return nil
		}
	}
	pm := newBackend()
	tui.PrintInfo("正在清理旧版本...")
	if err := runCommand(pm.Cleanup(apps, cleanupOpts)); err != nil {
		return err
	}
	tui.PrintSuccess("清理完成")
	return nil
}

func init() {
	cleanupCmd.Flags().BoolVarP(&cleanupOpts.All, "all", "a", false, "清理全部应用")
	cleanupCmd.Flags().BoolVarP(&cleanupOpts.Global, "global", "g", false, "清理全局应用")
	cleanupCmd.Flags().BoolVarP(&cleanupOpts.Cache, "cache", "k", false, "同时删除过期下载缓存")
	rootCmd.AddCommand(cleanupCmd)
}

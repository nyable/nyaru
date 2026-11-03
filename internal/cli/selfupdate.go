package cli

import (
	"context"

	"github.com/nyable/nyaru/internal/selfupdate"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var (
	selfCheck  bool
	selfForce  bool
	selfGlobal bool
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "更新 nyaru 自身（Scoop 托管）",
	Args:  cobra.NoArgs,
	Long: `检查并更新 nyaru。通过 Scoop 安装时，程序会退出并启动 helper，
等待当前进程结束后执行 scoop update nyaru；独立安装只做版本检查与提示。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return SelfUpdateAction(selfCheck)
	},
}

func SelfUpdateAction(checkOnly bool) error {
	status, err := selfupdate.Check(context.Background(), AppVersion)
	if err != nil {
		return err
	}
	if flagJSON {
		return printJSON(status)
	}
	tui.PrintInfo("当前版本: v" + status.Current)
	tui.PrintInfo("安装方式: " + string(status.Method))
	if status.Latest != "" {
		tui.PrintInfo("可用版本: v" + status.Latest)
	}
	if checkOnly {
		if status.UpdateAvailable {
			tui.PrintWarning("发现新版本，运行 nyaru self-update 更新")
		} else {
			tui.PrintSuccess("当前已是最新版本")
		}
		return nil
	}
	if status.Method != selfupdate.MethodScoop {
		tui.PrintWarning("当前不是 Scoop 托管安装，无法自动替换")
		tui.PrintInfo("请从 https://github.com/nyable/nyaru/releases 下载最新 nyaru.exe")
		return nil
	}
	if !status.UpdateAvailable && !selfForce {
		tui.PrintSuccess("当前已是最新版本")
		return nil
	}
	if !confirm("确认更新 nyaru 自身？程序将退出并在后台继续更新") {
		tui.PrintInfo("已取消更新")
		return nil
	}
	if err := selfupdate.Apply(context.Background(), status, selfupdate.ApplyOptions{
		Global: selfGlobal,
		Force:  selfForce,
	}); err != nil {
		return err
	}
	tui.PrintSuccess("更新 helper 已启动，nyaru 即将退出")
	if logPath, err := selfupdate.LogPath(); err == nil {
		tui.PrintInfo("更新日志: " + logPath)
	}
	return nil
}

func init() {
	selfUpdateCmd.Flags().BoolVar(&selfCheck, "check", false, "只检查是否有新版本")
	selfUpdateCmd.Flags().BoolVarP(&selfForce, "force", "f", false, "强制重新安装当前版本")
	selfUpdateCmd.Flags().BoolVarP(&selfGlobal, "global", "g", false, "更新全局安装")
	rootCmd.AddCommand(selfUpdateCmd)
}

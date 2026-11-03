package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var installOpts core.InstallOptions

var installCmd = &cobra.Command{
	Use:   "install <app> [app...]",
	Short: "安装指定的应用",
	Long:  `通过 scoop 安装一个或多个应用程序。`,
	Example: `  nyaru install git
  nyaru install aria2 curl wget`,
	Aliases: []string{"add", "i"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallAction(args)
	},
}

func InstallAction(apps []string) error {
	if len(apps) > 1 && configConfirmBatch() {
		if !confirm(fmt.Sprintf("确认安装这 %d 个应用？(%s)", len(apps), joinApps(apps))) {
			tui.PrintInfo("已取消安装")
			return nil
		}
	}
	pm := newBackend()
	cmd := pm.Install(apps, installOpts)
	tui.PrintInfo("正在安装: " + joinApps(apps))
	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}
	tui.PrintSuccess("安装完成")
	return nil
}

func joinApps(apps []string) string {
	out := ""
	for i, app := range apps {
		if i > 0 {
			out += ", "
		}
		out += app
	}
	return out
}

func init() {
	installCmd.Flags().BoolVarP(&installOpts.Global, "global", "g", false, "全局安装")
	installCmd.Flags().StringVar(&installOpts.Arch, "arch", "", "指定架构 (32bit/64bit/arm64)")
	installCmd.Flags().BoolVar(&installOpts.NoCache, "no-cache", false, "不使用下载缓存")
	installCmd.Flags().BoolVar(&installOpts.Independent, "independent", false, "不自动安装依赖")
	installCmd.Flags().BoolVar(&installOpts.SkipHashCheck, "skip-hash-check", false, "跳过哈希校验")
	installCmd.Flags().BoolVar(&installOpts.NoUpdateScoop, "no-update-scoop", false, "安装前不更新 Scoop")
	rootCmd.AddCommand(installCmd)
}

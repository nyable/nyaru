package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var updateOpts core.UpdateOptions

var updateCmd = &cobra.Command{
	Use:   "update [app...]",
	Short: "更新 Scoop、已安装应用或指定应用",
	Long: `不带参数时更新 Scoop 自身和 bucket；指定应用时更新应用；
使用 *、--all 或 -a 更新全部本地应用；再加 --global 会同时更新全局应用。
nyaru 自身请使用 self-update 更新。`,
	Example: `  nyaru update
  nyaru update git
  nyaru update *`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return UpdateAction(args...)
	},
}

func UpdateAction(args ...string) error {
	opts := updateOpts
	for _, arg := range args {
		if arg == "*" || arg == "--all" {
			opts.All = true
			args = nil
			break
		}
	}
	if opts.All && len(args) > 0 {
		return fmt.Errorf("--all 不能与具体应用同时使用")
	}

	var self bool
	for _, app := range args {
		if strings.EqualFold(app, "nyaru") {
			self = true
		}
	}
	if self {
		if len(args) > 1 || opts.All {
			return fmt.Errorf("nyaru 需要单独更新，请运行 nyaru self-update")
		}
		selfGlobal = opts.Global
		selfForce = opts.Force
		return SelfUpdateAction(false)
	}

	pm := newBackend()
	if len(args) == 0 && !opts.All {
		if opts.Global {
			return fmt.Errorf("--global 需要指定应用，或使用 --all --global 更新全部本地和全局应用")
		}
		if opts.NoCache {
			return fmt.Errorf("--no-cache 需要指定应用")
		}
		tui.PrintInfo("正在更新 Scoop 与 bucket...")
		if err := runCommand(pm.Update(nil, opts)); err != nil {
			return fmt.Errorf("更新失败: %w", err)
		}
		tui.PrintSuccess("Scoop 更新完成")
		return nil
	}

	batchAll := opts.All
	apps := args
	if opts.All {
		scope := "全部本地应用"
		if opts.Global {
			scope = "全部应用（含全局）"
		}
		if !confirm("确认更新" + scope + "？") {
			tui.PrintInfo("已取消更新")
			return nil
		}
		installed, err := pm.List(context.Background(), "")
		if err != nil {
			return err
		}

		if opts.Global && pm.Name() == "sfsu" {
			// sfsu's list command only reports the user context. Keep the
			// wildcard so global updates still work, but warn about nyaru.
			for _, app := range installed {
				if strings.EqualFold(app.Name, "nyaru") {
					tui.PrintInfo("已跳过 nyaru 自身，请使用 nyaru self-update 更新它")
					break
				}
			}
			tui.PrintWarning("sfsu 无法列出全局应用，将使用 scoop update * --global")
			tui.PrintWarning("若 nyaru 为全局安装，请改用 nyaru self-update --global")
			tui.PrintInfo("正在更新: 全部应用（含全局）")
			if err := runCommand(pm.Update(nil, opts)); err != nil {
				return fmt.Errorf("更新失败: %w", err)
			}
			tui.PrintSuccess("更新完成")
			return nil
		}

		var localApps, globalApps []string
		hasSelf := false
		for _, app := range installed {
			if strings.EqualFold(app.Name, "nyaru") {
				hasSelf = true
				continue
			}
			if app.Global {
				globalApps = append(globalApps, app.Name)
			} else {
				localApps = append(localApps, app.Name)
			}
		}
		if hasSelf {
			tui.PrintInfo("已跳过 nyaru 自身，请使用 nyaru self-update 更新它")
		}
		if opts.Global {
			if len(localApps) == 0 && len(globalApps) == 0 {
				tui.PrintSuccess("没有需要更新的应用")
				return nil
			}
			localOpts := opts
			localOpts.All = false
			localOpts.Global = false
			globalOpts := opts
			globalOpts.All = false
			globalOpts.Global = true
			if len(localApps) > 0 {
				tui.PrintInfo("正在更新: " + joinApps(localApps))
				if err := runCommand(pm.Update(localApps, localOpts)); err != nil {
					return fmt.Errorf("更新失败: %w", err)
				}
			}
			if len(globalApps) > 0 {
				tui.PrintInfo("正在更新全局应用: " + joinApps(globalApps))
				if err := runCommand(pm.Update(globalApps, globalOpts)); err != nil {
					return fmt.Errorf("更新失败: %w", err)
				}
			}
			tui.PrintSuccess("更新完成")
			return nil
		}
		apps = localApps
		opts.All = false
		if len(apps) == 0 {
			tui.PrintSuccess("没有需要更新的应用")
			return nil
		}
	}

	if !batchAll && len(apps) > 1 && configConfirmBatch() {
		if !confirm(fmt.Sprintf("确认更新这 %d 个应用？(%s)", len(apps), joinApps(apps))) {
			tui.PrintInfo("已取消更新")
			return nil
		}
	}

	target := displayTargets(apps, opts.All)
	tui.PrintInfo("正在更新: " + target)
	if err := runCommand(pm.Update(apps, opts)); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	tui.PrintSuccess("更新完成")
	return nil
}

func displayTargets(apps []string, all bool) string {
	if all {
		return "全部应用"
	}
	if len(apps) == 0 {
		return "Scoop 与 bucket"
	}
	return joinApps(apps)
}

func init() {
	updateCmd.Flags().BoolVarP(&updateOpts.All, "all", "a", false, "更新全部本地应用（加 --global 时同时更新全局应用）")
	updateCmd.Flags().BoolVarP(&updateOpts.Global, "global", "g", false, "更新全局应用")
	updateCmd.Flags().BoolVarP(&updateOpts.Force, "force", "f", false, "强制更新")
	updateCmd.Flags().BoolVarP(&updateOpts.Independent, "independent", "i", false, "不自动安装依赖")
	updateCmd.Flags().BoolVar(&updateOpts.NoCache, "no-cache", false, "不使用下载缓存")
	updateCmd.Flags().BoolVar(&updateOpts.SkipHashCheck, "skip-hash-check", false, "跳过哈希校验")
	rootCmd.AddCommand(updateCmd)
}

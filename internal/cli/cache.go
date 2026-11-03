package cli

import (
	"context"
	"fmt"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var cacheListCmd = &cobra.Command{
	Use:     "list",
	Short:   "显示缓存内容",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls", "show"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return CacheAction()
	},
}

func CacheAction() error {
	pm := newBackend()
	if !tuiEnabled() {
		cache, err := pm.CacheList(context.Background())
		if err != nil {
			return err
		}
		if flagJSON {
			return printJSON(cache)
		}
		printCacheTable(cache)
		return nil
	}
	return runTUI(tui.Options{
		Backend: pm,
		Start:   tui.StartCache,
		Version: AppVersion,
	})
}

var cacheRemoveAll bool

var cacheRemoveCmd = &cobra.Command{
	Use:     "rm [app...]",
	Short:   "删除缓存",
	Long:    `删除指定应用的全部缓存版本；使用 --all 删除所有缓存。`,
	Aliases: []string{"remove"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if cacheRemoveAll && len(args) > 0 {
			return fmt.Errorf("--all 不能与具体应用同时使用")
		}
		if !cacheRemoveAll && len(args) == 0 {
			return fmt.Errorf("请指定应用名，或使用 --all 删除全部缓存")
		}
		target := "全部缓存"
		if !cacheRemoveAll {
			target = joinApps(args)
		}
		if !confirm("确认删除 " + target + "？") {
			tui.PrintInfo("已取消删除")
			return nil
		}
		pm := newBackend()
		tui.PrintInfo("正在清理缓存: " + target)
		if err := runCommand(pm.CacheRemove(args, cacheRemoveAll)); err != nil {
			return err
		}
		tui.PrintSuccess("缓存清理完成")
		return nil
	},
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "缓存管理",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return CacheAction()
	},
}

func init() {
	cacheRemoveCmd.Flags().BoolVarP(&cacheRemoveAll, "all", "a", false, "删除全部缓存")
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheRemoveCmd)
	rootCmd.AddCommand(cacheCmd)
}

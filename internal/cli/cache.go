package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/nyable/nyaru/internal/utils"
	"github.com/spf13/cobra"
)

var cacheListCmd = &cobra.Command{
	Use:     "list",
	Short:   "显示缓存内容",
	Long:    `显示缓存内容`,
	Aliases: []string{"ls", "show"},
	Run: func(cmd *cobra.Command, args []string) {
		CacheAction()
	},
}

func CacheAction() {
	pm := core.GetManager(config.GetActiveMode())

	res, err := tui.RunWithSpinner("正在列出缓存内容", func() (any, error) {
		return pm.CacheList()
	})

	if err != nil {
		tui.PrintError(fmt.Sprintf("列出缓存内容出错:\n%v", err))
		return
	}

	dataList := res.([]models.CacheResult)
	if len(dataList) == 0 {
		tui.PrintWarning("没有缓存！")
		return
	}

	var totalSize int64
	for _, item := range dataList {
		totalSize += item.Length
	}
	formatTotal := utils.SizeToHumanRead(totalSize)

	var items []list.Item
	for i := range dataList {
		items = append(items, dataList[i])
	}

	title := fmt.Sprintf("缓存列表 (共 %d 个文件, 总计 %s)", len(dataList), formatTotal)
	results, err := tui.RunListInteractive(title, items, pm.Info)
	if err != nil {
		tui.PrintError(fmt.Sprintf("TUI Error: %v", err))
		return
	}

	if len(results) > 0 {
		cmdActions := []models.CmdAction{
			{Command: "rm", Desc: "删除缓存"},
			{Command: "none", Desc: "什么也不做"},
		}

		options := []string{}
		actionMap := make(map[string]models.CmdAction)
		for _, action := range cmdActions {
			label := fmt.Sprintf("%s (%s)", action.Command, action.Desc)
			options = append(options, label)
			actionMap[label] = action
		}

		selLabel, err := tui.RunSingleSelect("想要进行的操作是?", options)
		if err != nil {
			tui.PrintError(fmt.Sprintf("选择操作出错: %v", err))
			return
		}

		action := actionMap[selLabel]
		if action.Command == "none" {
			return
		}

		if action.Command == "rm" {
			var names []string
			for _, item := range results {
				if cache, ok := item.(models.CacheResult); ok {
					names = append(names, cache.Name)
				}
			}

			if len(names) == 0 {
				return
			}

			// confirm if multiselect
			if len(names) > 1 {
				fmt.Printf("⚠ 确认删除这 %d 个缓存文件吗？(y/n): ", len(names))
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" {
					tui.PrintInfo("已取消删除")
					return
				}
			}

			// Optimization: if all items selected, use '*'
			if len(names) == len(dataList) {
				tui.PrintInfo("正在清理所有缓存...")
				if err := pm.CacheRemove("*"); err != nil {
					tui.PrintError(fmt.Sprintf("清理失败: %v", err))
				} else {
					tui.PrintSuccess("所有缓存已清理!")
				}
			} else {
				tui.PrintInfo(fmt.Sprintf("正在批量删除 %d 个缓存...", len(names)))
				if err := pm.CacheRemove(names...); err != nil {
					tui.PrintError(fmt.Sprintf("批量删除失败: %v", err))
				} else {
					tui.PrintSuccess(fmt.Sprintf("已成功删除 %d 个缓存!", len(names)))
				}
			}
		}
	}
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "缓存管理",
	Long:  `缓存管理`,
	Run: func(cmd *cobra.Command, args []string) {
		CacheAction()
	},
}

func init() {
	cacheCmd.AddCommand(cacheListCmd)
	rootCmd.AddCommand(cacheCmd)
}

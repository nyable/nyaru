package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "搜索可安装的应用程序(别名:find/query/s)",
	Long:    `搜索与查询匹配的应用名称。使用配置文件中指定的包管理器（sfsu 或 scoop）。`,
	Example: `nyaru search aria2`,
	Aliases: []string{"find", "query", "s"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var query string
		if len(args) > 0 {
			query = args[0]
		}
		SearchAction(query)
	},
}

func SearchAction(query string) {
	pm := core.GetManager(config.GetActiveMode())

	res, err := tui.RunWithSpinner("正在搜索中...", func() (any, error) {
		return pm.Search(query)
	})

	if err != nil {
		tui.PrintError(fmt.Sprintf("搜索出错:\n%v", err))
		return
	}

	dataList := res.([]models.AppInfo)

	if len(dataList) == 0 {
		tui.PrintWarning("没有匹配的搜索结果！")
		return
	}

		var items []list.Item
		for _, v := range dataList {
			// Create a local copy to avoid pointer sharing mechanics issues in range loop
			app := v
			items = append(items, app)
		}

		results, err := tui.RunListInteractive("Search Results ("+config.GetActiveMode()+")", items, pm.Info)


	if err != nil {
		tui.PrintError(fmt.Sprintf("TUI Error: %v", err))
		return
	}

		if len(results) > 0 {
			cmdActions := []models.CmdAction{
				{Command: "install", Desc: "安装选中的应用"},
				{Command: "none", Desc: "什么也不做"},
			}

			options := []string{}
			actionMap := make(map[string]models.CmdAction)
			for _, action := range cmdActions {
				label := fmt.Sprintf("%s (%s)", action.Command, action.Desc)
				options = append(options, label)
				actionMap[label] = action
			}

			selLabel, err := tui.RunSingleSelect("想要执行的操作是?", options)
			if err != nil {
				tui.PrintError(fmt.Sprintf("选择操作出错: %v", err))
				return
			}

			action := actionMap[selLabel]
			if action.Command == "none" {
				return
			}

			if action.Command == "install" {
				var names []string
				for _, item := range results {
					if choice, ok := item.(models.AppInfo); ok {
						names = append(names, choice.FullName())
					}
				}

				if len(names) == 0 {
					return
				}

				// Confirm if multiselect
				if len(names) > 1 {
					fmt.Printf("⚠ 确认安装这 %d 个应用程序吗？(y/n): ", len(names))
					var confirm string
					fmt.Scanln(&confirm)
					if strings.ToLower(confirm) != "y" {
						tui.PrintInfo("已取消安装")
						return
					}
				}

				for _, name := range names {
					tui.PrintInfo(fmt.Sprintf("正在安装: %s", name))
					if err := pm.Install(name); err != nil {
						tui.PrintError(fmt.Sprintf("安装 %s 失败: %v", name, err))
					} else {
						tui.PrintSuccess(fmt.Sprintf("安装 %s 成功!", name))
					}
				}
			}
		}

}

func init() {
	rootCmd.AddCommand(searchCmd)
}

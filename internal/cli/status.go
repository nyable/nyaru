package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "检查已安装应用的更新状态",
	Long:  `检查已安装应用的更新状态`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := core.GetManager(config.GetActiveMode())

		res, err := tui.RunWithSpinner("正在列出可更新的应用程序...", func() (any, error) {
			return pm.Status()
		})

		if err != nil {
			tui.PrintError(fmt.Sprintf("列出状态出错:\n%v", err))
			os.Exit(1)
		}

		dataList := res.([]models.AppInfo)

		if len(dataList) == 0 {
			tui.PrintWarning("没有可更新的应用程序！")
			os.Exit(0)
		}

		var items []list.Item
		for _, v := range dataList {
			app := v
			app.Installed = true
			items = append(items, app)
		}

		results, err := tui.RunListInteractive("Updatable Apps ("+config.GetActiveMode()+")", items, pm.Info)

		if err != nil {
			tui.PrintError(fmt.Sprintf("TUI Error: %v", err))
			os.Exit(1)
		}

		if len(results) > 0 {
			cmdActions := []models.CmdAction{
				{Command: "update", Desc: "更新选中的应用"},
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
				os.Exit(1)
			}

			action := actionMap[selLabel]
			if action.Command == "none" {
				return
			}

			if action.Command == "update" {
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
					fmt.Printf("⚠ 确认更新这 %d 个应用程序吗？(y/n): ", len(names))
					var confirm string
					fmt.Scanln(&confirm)
					if strings.ToLower(confirm) != "y" {
						tui.PrintInfo("已取消更新")
						return
					}
				}

				tui.PrintInfo(fmt.Sprintf("正在更新选中的 %d 个应用: %s", len(names), strings.Join(names, ", ")))
				if err := pm.Update(names...); err != nil {
					tui.PrintError(fmt.Sprintf("更新失败: %v", err))
				} else {
					tui.PrintSuccess("所有应用更新尝试已完成!")
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

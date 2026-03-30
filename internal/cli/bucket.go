package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var bucketListCmd = &cobra.Command{
	Use:     "list",
	Short:   "列出Scoop存储桶(别名:ls)",
	Long:    `列出已添加的Scoop存储桶`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		BucketAction()
	},
}

func BucketAction() {
	pm := core.GetManager(config.GetActiveMode())

	res, err := tui.RunWithSpinner("正在列出已添加的存储桶...", func() (any, error) {
		return pm.BucketList()
	})

	if err != nil {
		tui.PrintError(fmt.Sprintf("列出存储桶出错:\n%v", err))
		return
	}

	dataList := res.([]models.BucketResult)
	if len(dataList) == 0 {
		tui.PrintWarning("没有添加任何存储桶！")
		return
	}

	var items []list.Item
	for _, v := range dataList {
		items = append(items, v)
	}

	results, err := tui.RunListInteractive("Bucket List ("+config.GetActiveMode()+")", items, pm.Info)
	if err != nil {
		tui.PrintError(fmt.Sprintf("TUI Error: %v", err))
		return
	}

	if len(results) > 0 {
		cmdActions := []models.CmdAction{
			{Command: "rm", Desc: "删除存储桶"},
			{Command: "add", Desc: "显示添加此存储桶的命令"},
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

		for _, item := range results {
			if bucket, ok := item.(models.BucketResult); ok {
				switch action.Command {
				case "rm":
					tui.PrintInfo(fmt.Sprintf("正在删除存储桶: %s", bucket.Name))
					if err := pm.BucketRemove(bucket.Name); err != nil {
						tui.PrintError(fmt.Sprintf("删除失败: %v", err))
					} else {
						tui.PrintSuccess(fmt.Sprintf("删除 %s 成功!", bucket.Name))
					}
				case "add":
					tui.PrintInfo(fmt.Sprintf("添加 %s 的命令为:", bucket.Name))
					fmt.Printf("scoop bucket add %s %s\n", bucket.Name, bucket.Source)
				}
			}
		}
	}
}

var bucketAddCmd = &cobra.Command{
	Use:     "add <name> [source]",
	Short:   "添加存储桶",
	Long:    `添加一个 Scoop 存储桶。官方桶只需名称，第三方桶需要提供 git 地址。`,
	Example: `  nyaru bucket add extras
  nyaru bucket add my-bucket https://github.com/user/my-bucket`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		source := ""
		if len(args) > 1 {
			source = args[1]
		}
		BucketAddAction(name, source)
	},
}

func BucketAddAction(name, source string) {
	pm := core.GetManager(config.GetActiveMode())
	tui.PrintInfo(fmt.Sprintf("正在添加存储桶: %s", name))
	if err := pm.BucketAdd(name, source); err != nil {
		tui.PrintError(fmt.Sprintf("添加存储桶 %s 失败: %v", name, err))
	} else {
		tui.PrintSuccess(fmt.Sprintf("添加存储桶 %s 成功!", name))
	}
}

var bucketRemoveCmd = &cobra.Command{
	Use:     "rm <name>",
	Short:   "删除存储桶(别名:remove)",
	Long:    `删除一个已添加的 Scoop 存储桶。`,
	Example: `  nyaru bucket rm extras`,
	Aliases: []string{"remove"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		BucketRemoveAction(args[0])
	},
}

func BucketRemoveAction(name string) {
	pm := core.GetManager(config.GetActiveMode())
	tui.PrintInfo(fmt.Sprintf("正在删除存储桶: %s", name))
	if err := pm.BucketRemove(name); err != nil {
		tui.PrintError(fmt.Sprintf("删除存储桶 %s 失败: %v", name, err))
	} else {
		tui.PrintSuccess(fmt.Sprintf("删除存储桶 %s 成功!", name))
	}
}

var bucketCmd = &cobra.Command{
	Use:     "bucket",
	Short:   "管理Scoop存储桶(别名:bt)",
	Long:    `管理Scoop存储桶`,
	Aliases: []string{"bt"},
	Run: func(cmd *cobra.Command, args []string) {
		BucketAction()
	},
}

func init() {
	bucketCmd.AddCommand(bucketListCmd)
	bucketCmd.AddCommand(bucketAddCmd)
	bucketCmd.AddCommand(bucketRemoveCmd)
	rootCmd.AddCommand(bucketCmd)
}


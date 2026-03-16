package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/nyable/nyaru/internal/utils"
	"github.com/spf13/cobra"
)

type CacheResult struct {
	Index      int
	Name       string
	Version    string
	Length     int
	FormatSize string
}

var cacheListCmd = &cobra.Command{
	Use:     "list",
	Short:   "显示缓存内容",
	Long:    `显示缓存内容`,
	Aliases: []string{"ls", "show"},
	Run: func(cmd *cobra.Command, args []string) {
		pureCmdStr := "scoop cache show"
		tui.PrintInfo(pureCmdStr)

		res, err := tui.RunWithSpinner("正在列出缓存内容", func() (any, error) {
			strOutput, _, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
			if err != nil {
				return nil, err
			}
			return utils.PsDirtyJSONToStructList[CacheResult](strOutput)
		})
		
		if err != nil {
			tui.PrintError(fmt.Sprintf("执行命令 %s 时出错:\n%v", pureCmdStr, err))
			os.Exit(1)
		}

		dataList := res.([]CacheResult)
		dataSize := len(dataList)
		tui.PrintSuccess(pureCmdStr)

		if dataSize == 0 {
			tui.PrintWarning("没有缓存！")
			os.Exit(0)
		}
		
		maxNumLen := 1
		maxNameLen := 0
		maxVersionLen := 0
		maxFormatSizeLen := 0

		for i, data := range dataList {
			dataIndex := i + 1
			dataList[i].Index = dataIndex
			dataList[i].FormatSize = utils.SizeToHumanRead(int64(data.Length))
			dataName := data.Name
			dataVersion := data.Version
			cNameLen := len(dataName)
			if cNameLen > maxNameLen {
				maxNameLen = cNameLen
			}
			cVersionLen := len(dataVersion)
			if cVersionLen > maxVersionLen {
				maxVersionLen = cVersionLen
			}
			cIndexLen := len(fmt.Sprintf("%d", dataIndex))
			if cIndexLen > maxNumLen {
				maxNumLen = cIndexLen
			}
			cFormatSizeLen := len(fmt.Sprint(data.FormatSize))
			if cFormatSizeLen > maxFormatSizeLen {
				maxFormatSizeLen = cFormatSizeLen
			}
		}
		
		var optList []string
		optMap := make(map[string]CacheResult)
		for _, app := range dataList {
			optLabel := fmt.Sprintf("%-*d | %-*s | %-*s | %-*s", maxNumLen, app.Index, maxNameLen, app.Name, maxVersionLen, app.Version, maxFormatSizeLen, app.FormatSize)
			optMap[optLabel] = app
			optList = append(optList, optLabel)
		}
		
		optSize := len(optList)
		selOptList, err := tui.RunMultiSelect("请选择需要操作的缓存", optList)
		if err != nil {
			tui.PrintError(fmt.Sprintf("获取选项时出错: %v", err))
			os.Exit(1)
		}
		
		selOptSize := len(selOptList)
		tui.PrintInfo(fmt.Sprintf("选中了 %d 个缓存", selOptSize))
		if selOptSize == 0 {
			tui.PrintWarning("没有选择任何缓存!退出运行!")
			os.Exit(0)
		}

		cmdActions := []models.CmdAction{
			{Command: "none", Desc: "什么也不做"},
			{Command: "rm", Desc: "删除缓存"},
		}

		actionMap := make(map[string]models.CmdAction)
		options := []string{}
		for _, cmdAction := range cmdActions {
			optLabel := fmt.Sprintf("%s (%s)", cmdAction.Command, cmdAction.Desc)
			actionMap[optLabel] = cmdAction
			options = append(options, optLabel)
		}
		
		selAction, err := tui.RunSingleSelect("想要进行的操作是?", options)
		if err != nil {
			tui.PrintError(fmt.Sprintf("选择操作时出错: %v", err))
			os.Exit(1)
		}
		tui.PrintInfo(fmt.Sprintf("选择: %s", selAction))
		
		command := actionMap[selAction].Command
		tui.PrintWarning(fmt.Sprintf("对所有选中缓存执行命令:%s", command))
		
		if command == "rm" {
			if selOptSize == optSize {
				rmAllCacheCmd := exec.Command("scoop", "cache", "rm", "*")
				rmAllCacheCmdStr := strings.Join(rmAllCacheCmd.Args, " ")
				tui.PrintInfo("开始执行命令:")
				fmt.Println(rmAllCacheCmdStr)
				tui.PrintInfo("==========")
				rmAllCacheCmd.Stdout = os.Stdout
				rmAllCacheCmd.Stderr = os.Stderr
				err := rmAllCacheCmd.Run()
				if err != nil {
					tui.PrintError(fmt.Sprintf("执行命令: %s 时出错:\n%v", rmAllCacheCmd, err))
				} else {
					tui.PrintSuccess("执行完毕!")
				}
			} else {
				var sucCount = 0
				var errCount = 0
				for _, selData := range selOptList {
					data := optMap[selData]
					bucketName := data.Name
					rmBucketCmd := exec.Command("scoop", "cache", "rm", bucketName)
					rmBucketCmdStr := strings.Join(rmBucketCmd.Args, " ")
					tui.PrintInfo("开始执行命令:")
					fmt.Println(rmBucketCmdStr)
					tui.PrintInfo("==========")
					rmBucketCmd.Stdout = os.Stdout
					rmBucketCmd.Stderr = os.Stderr
					err := rmBucketCmd.Run()
					if err != nil {
						errCount++
						tui.PrintError(fmt.Sprintf("执行命令: %s 时出错:\n%v", rmBucketCmd, err))
					} else {
						sucCount++
						tui.PrintSuccess("执行完毕!")
					}
					tui.PrintInfo("==========")
					tui.PrintInfo(fmt.Sprintf("成功 %d 个，失败 %d 个", sucCount, errCount))
				}
			}
		} else {
			os.Exit(0)
		}
	},
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "缓存管理",
	Long:  `缓存管理`,
	Run: func(cmd *cobra.Command, args []string) {
		cacheListCmd.Run(cmd, args)
	},
}

func init() {
	cacheCmd.AddCommand(cacheListCmd)
	rootCmd.AddCommand(cacheCmd)
}

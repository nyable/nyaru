package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/utils"
	"github.com/pterm/pterm"
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
		fmt.Println(pureCmdStr)
		spinner, _ := pterm.DefaultSpinner.Start("正在列出缓存内容")
		print("\n")
		strOutput, cmdStr, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}
		println(strOutput)
		dataList, err := utils.PsDirtyJSONToStructList[CacheResult](strOutput)
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}

		dataSize := len(dataList)
		spinner.Success(pureCmdStr)

		if dataSize == 0 {
			pterm.Warning.Println("没有缓存！")
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
		selOptList, err := pterm.DefaultInteractiveMultiselect.WithDefaultText("请选择需要操作的缓存").WithOptions(optList).WithMaxHeight(20).Show()
		if err != nil {
			pterm.Error.Println("获取选项时出错:", err.Error())
			os.Exit(1)
		}
		var selOptSize = len(selOptList)
		pterm.Info.Printfln("选中了 %d 个缓存", selOptSize)
		if selOptSize == 0 {
			pterm.Warning.Println("没有选择任何存储桶!退出运行!")
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
		selAction, _ := pterm.DefaultInteractiveSelect.WithDefaultText("想要进行的操作是?").WithOptions(options).Show()
		pterm.Printfln("选择: %s", selAction)
		command := actionMap[selAction].Command
		pterm.Warning.Printfln("对所有选中存储桶执行命令:%s", command)
		if command == "rm" {
			if selOptSize == optSize {
				rmAllCacheCmd := exec.Command("scoop", "cache", "rm", "*")
				rmAllCacheCmdStr := strings.Join(rmAllCacheCmd.Args, " ")
				pterm.Info.Println("开始执行命令:")
				println(rmAllCacheCmdStr)
				pterm.Info.Println("==========")
				rmAllCacheCmd.Stdout = os.Stdout
				rmAllCacheCmd.Stderr = os.Stderr
				err := rmAllCacheCmd.Run()
				if err != nil {
					pterm.Error.Println(fmt.Sprintf("执行命令: %s 时出错:\n%s", rmAllCacheCmd, err.Error()))
				} else {
					pterm.Success.Printfln("执行完毕!")
				}
			} else {
				var sucCount = 0
				var errCount = 0
				for _, selData := range selOptList {
					data := optMap[selData]
					bucketName := data.Name
					rmBucketCmd := exec.Command("scoop", "cache", "rm", bucketName)
					rmBucketCmdStr := strings.Join(rmBucketCmd.Args, " ")
					pterm.Info.Println("开始执行命令:")
					println(rmBucketCmdStr)
					pterm.Info.Println("==========")
					rmBucketCmd.Stdout = os.Stdout
					rmBucketCmd.Stderr = os.Stderr
					err := rmBucketCmd.Run()
					if err != nil {
						errCount++
						pterm.Error.Println(fmt.Sprintf("执行命令: %s 时出错:\n%s", rmBucketCmd, err.Error()))
					} else {
						sucCount++
						pterm.Success.Printfln("执行完毕!")
					}
					pterm.Info.Println("==========")
					pterm.Info.Printfln("成功 %d 个，失败 %d 个", sucCount, errCount)
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

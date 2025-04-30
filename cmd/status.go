package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nyable/nyaru/internal/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type StatusResult struct {
	Index      int
	Name       string `json:"Name"`
	Version    string `json:"Installed Version"`
	NewVersion string `json:"Latest Version"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "检查已安装应用的更新状态",
	Long:  `检查已安装应用的更新状态`,
	Run: func(cmd *cobra.Command, args []string) {
		pureCmdStr := "scoop status"
		fmt.Println(pureCmdStr)
		spinner, _ := pterm.DefaultSpinner.Start("正在列出已安装应用程序的更新状态")
		strOutput, cmdStr, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}
		println(strOutput)
		dataList, err := utils.PsDirtyJSONToStructList[StatusResult](strOutput)
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}

		dataSize := len(dataList)
		spinner.Success(pureCmdStr)

		if dataSize == 0 {
			pterm.Warning.Println("没有可更新的应用程序！")
			os.Exit(0)
		}

		maxNumLen := 1
		maxNameLen := 0
		maxVersionLen := 0
		maxNewVersion := 0

		for i, data := range dataList {
			dataIndex := i + 1
			dataList[i].Index = dataIndex
			dataName := data.Name
			dataVersion := data.Version
			dataNewVersion := data.NewVersion
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
			cNewVersionLen := len(dataNewVersion)
			if cNewVersionLen > maxNewVersion {
				maxNewVersion = cNewVersionLen
			}

		}
		var optList []string
		optMap := make(map[string]StatusResult)
		for _, data := range dataList {
			optLabel := fmt.Sprintf("%-*d | %-*s | %-*s | %-*s", maxNumLen, data.Index, maxNameLen, data.Name, maxVersionLen, data.Version, maxNewVersion, data.NewVersion)
			optMap[optLabel] = data
			optList = append(optList, optLabel)
		}
		selOptList, err := pterm.DefaultInteractiveMultiselect.WithDefaultText("请选择需要更新的应用程序").WithOptions(optList).WithMaxHeight(20).Show()
		if err != nil {
			pterm.Error.Println("获取选项时出错:", err.Error())
			os.Exit(1)
		}
		var selOptSize = len(selOptList)
		pterm.Info.Printfln("选中了 %d 个应用程序", selOptSize)
		if selOptSize == 0 {
			pterm.Warning.Println("没有选择任何应用程序!退出运行!")
			os.Exit(0)
		}
		var sucCount = 0
		var errCount = 0
		for _, selData := range selOptList {
			data := optMap[selData]
			bucketName := data.Name
			rmBucketCmd := exec.Command("scoop", "update", bucketName)
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

	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/nyable/nyaru/internal/utils"
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
	Short: "检查已安装应用的更新状态(先执行update然后执行status)",
	Long:  `检查已安装应用的更新状态`,
	Run: func(cmd *cobra.Command, args []string) {
		updateCmd.Run(cmd, []string{})
		pureCmdStr := "scoop status"
		tui.PrintInfo(pureCmdStr)
		
		res, err := tui.RunWithSpinner("正在列出已安装应用程序的更新状态", func() (any, error) {
			strOutput, _, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
			if err != nil {
				return nil, err
			}
			return utils.PsDirtyJSONToStructList[StatusResult](strOutput)
		})
		
		if err != nil {
			tui.PrintError(fmt.Sprintf("执行命令 %s 时出错:\n%v", pureCmdStr, err))
			os.Exit(1)
		}
		
		dataList := res.([]StatusResult)
		dataSize := len(dataList)
		tui.PrintSuccess(pureCmdStr)

		if dataSize == 0 {
			tui.PrintWarning("没有可更新的应用程序！")
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
		
		selOptList, err := tui.RunMultiSelect("请选择需要更新的应用程序", optList)
		if err != nil {
			tui.PrintError(fmt.Sprintf("获取选项时出错: %v", err))
			os.Exit(1)
		}
		
		selOptSize := len(selOptList)
		tui.PrintInfo(fmt.Sprintf("选中了 %d 个应用程序", selOptSize))
		if selOptSize == 0 {
			tui.PrintWarning("没有选择任何应用程序!退出运行!")
			os.Exit(0)
		}
		
		var sucCount = 0
		var errCount = 0
		for _, selData := range selOptList {
			data := optMap[selData]
			bucketName := data.Name
			rmBucketCmd := exec.Command("scoop", "update", bucketName)
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
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

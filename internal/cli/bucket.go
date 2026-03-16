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

type BucketUpdatedInfo struct {
	Value       string `json:"value"`
	DisplayHint int    `json:"DisplayHint"`
	DateTime    string `json:"DateTime"`
}

type BucketResult struct {
	Index     int
	Name      string
	Source    string
	Updated   BucketUpdatedInfo
	Manifests int
}

var bucketListCmd = &cobra.Command{
	Use:     "list",
	Short:   "列出Scoop存储桶",
	Long:    `列出Scoop存储桶`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		pureCmdStr := "scoop bucket list"
		tui.PrintInfo(pureCmdStr)
		
		res, err := tui.RunWithSpinner("正在列出已添加的存储桶", func() (any, error) {
			strOutput, _, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
			if err != nil {
				return nil, err
			}
			return utils.PsDirtyJSONToStructList[BucketResult](strOutput)
		})
		
		if err != nil {
			tui.PrintError(fmt.Sprintf("执行命令 %s 时出错:\n%v", pureCmdStr, err))
			os.Exit(1)
		}
		
		dataList := res.([]BucketResult)
		dataSize := len(dataList)
		tui.PrintSuccess(pureCmdStr)

		if dataSize == 0 {
			tui.PrintWarning("没有添加任何存储桶！")
			os.Exit(0)
		}
		tui.PrintInfo(fmt.Sprintf("已添加 %d 个存储桶", dataSize))

		maxNumLen := 1
		maxNameLen := 0
		maxSourceLen := 0
		maxManifestsLen := 0

		for i, data := range dataList {
			dataIndex := i + 1
			dataList[i].Index = dataIndex
			dataName := data.Name
			dataSource := data.Source
			cNameLen := len(dataName)
			if cNameLen > maxNameLen {
				maxNameLen = cNameLen
			}
			cSourceLen := len(dataSource)
			if cSourceLen > maxSourceLen {
				maxSourceLen = cSourceLen
			}
			cIndexLen := len(fmt.Sprintf("%d", dataIndex))
			if cIndexLen > maxNumLen {
				maxNumLen = cIndexLen
			}
			cBinariesLen := len(fmt.Sprint(data.Manifests))
			if cBinariesLen > maxManifestsLen {
				maxManifestsLen = cBinariesLen
			}
		}
		
		var optList []string
		optMap := make(map[string]BucketResult)
		for _, app := range dataList {
			optLabel := fmt.Sprintf("%-*d | %-*s | %-*s | %-*d | %-*s", maxNumLen, app.Index, maxNameLen, app.Name, maxSourceLen, app.Source, maxManifestsLen, app.Manifests, 20, app.Updated.DateTime)
			optMap[optLabel] = app
			optList = append(optList, optLabel)
		}
		
		selOptList, err := tui.RunMultiSelect("选择一个想要操作的存储桶", optList)
		if err != nil {
			tui.PrintError(fmt.Sprintf("选择存储桶时出错: %v", err))
			os.Exit(1)
		}
		
		selOptSize := len(selOptList)
		tui.PrintInfo(fmt.Sprintf("选中了 %d 个存储桶", selOptSize))
		if selOptSize == 0 {
			tui.PrintWarning("没有选择任何存储桶!退出运行!")
			os.Exit(0)
		}

		cmdActions := []models.CmdAction{
			{Command: "none", Desc: "什么也不做"},
			{Command: "add", Desc: "生成添加存储桶的命令"},
			{Command: "rm", Desc: "删除存储桶"},
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
		tui.PrintWarning(fmt.Sprintf("对所有选中存储桶执行命令:%s", command))
		
		var sucCount = 0
		var errCount = 0
		for _, selOpt := range selOptList {
			selData := optMap[selOpt]
			bucketName := selData.Name

			if command == "rm" {
				rmBucketCmd := exec.Command("scoop", "bucket", "rm", bucketName)
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
			} else if command == "add" {
				fmt.Printf("scoop bucket add %s %s\n", bucketName, selData.Source)
			} else {
				os.Exit(0)
			}
		}

	},
}

var bucketCmd = &cobra.Command{
	Use:     "bucket",
	Short:   "管理Scoop存储桶",
	Long:    `管理Scoop存储桶`,
	Aliases: []string{"bt"},
	Run: func(cmd *cobra.Command, args []string) {
		bucketListCmd.Run(cmd, args)
	},
}

func init() {
	bucketCmd.AddCommand(bucketListCmd)
	rootCmd.AddCommand(bucketCmd)
}

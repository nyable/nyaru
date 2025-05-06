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
		fmt.Println(pureCmdStr)
		spinner, _ := pterm.DefaultSpinner.Start("正在列出已添加的存储桶")
		print("\n")
		strOutput, cmdStr, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}
		println(strOutput)
		dataList, err := utils.PsDirtyJSONToStructList[BucketResult](strOutput)
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}

		dataSize := len(dataList)
		spinner.Success(pureCmdStr)

		if dataSize == 0 {
			pterm.Warning.Println("没有添加任何存储桶！")
			os.Exit(0)
		}
		pterm.Println(fmt.Sprintf("已添加 %d 个存储桶", dataSize))

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
		selOptList, err := pterm.DefaultInteractiveMultiselect.WithDefaultText("选择一个想要操作的存储桶").WithOptions(optList).WithMaxHeight(20).Show()
		if err != nil {
			pterm.Error.Println("选择存储桶时出错:", err.Error())
			os.Exit(1)
		}
		var selOptSize = len(selOptList)
		pterm.Info.Printfln("选中了 %d 个存储桶", selOptSize)
		if selOptSize == 0 {
			pterm.Warning.Println("没有选择任何存储桶!退出运行!")
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
		selAction, _ := pterm.DefaultInteractiveSelect.WithDefaultText("想要进行的操作是?").WithOptions(options).Show()
		pterm.Printfln("选择: %s", selAction)
		command := actionMap[selAction].Command
		pterm.Warning.Printfln("对所有选中存储桶执行命令:%s", command)
		var sucCount = 0
		var errCount = 0
		for _, selOpt := range selOptList {
			selData := optMap[selOpt]
			bucketName := selData.Name

			if command == "rm" {
				rmBucketCmd := exec.Command("scoop", "bucket", "rm", bucketName)
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

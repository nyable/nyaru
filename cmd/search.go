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

type SearchResult struct {
	Index    int
	Name     string
	Version  string
	Source   string
	FullName string
	Binaries string
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "搜索可安装的应用程序",
	Long: `如果与 [query] 一起使用，则会显示与查询匹配的应用名称。
- 启用“use_sqlite_cache”后，[query] 会与应用名称、二进制文件和快捷方式进行部分匹配。
- 如果不启用“use_sqlite_cache”，[query] 可以使用正则表达式来匹配应用名称和二进制文件。
如果不启用 [query]，则会显示所有可用的应用。`,
	Example: `nyaru search aria2 # 搜索 aria2`,
	Aliases: []string{"find", "query", "s"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var query string
		if len(args) > 0 {
			query = args[0]
		}
		pureCmdStr := fmt.Sprintf("scoop search %s", query)
		fmt.Println(pureCmdStr)
		spinner, _ := pterm.DefaultSpinner.Start("正在从现有仓库中搜索")
		strOutput, cmdStr, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}
		println(strOutput)
		appList, err := utils.PsDirtyJSONToStructList[SearchResult](strOutput)
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}

		appSize := len(appList)
		spinner.Success(pureCmdStr)

		if appSize == 0 {
			pterm.Warning.Println("没有匹配的搜索结果！")
			os.Exit(0)
		}
		pterm.Println(fmt.Sprintf("以下是: %s 的匹配结果,共 %d 条", query, appSize))

		maxNumLen := 1
		maxNameLen := 0
		maxVersionLen := 0
		maxSourceLen := 0
		maxBinariesLen := 0

		for i, app := range appList {
			appIndex := i + 1
			appList[i].Index = appIndex
			appList[i].FullName = fmt.Sprintf("%s/%s", app.Source, app.Name)
			appName := app.Name
			appVersion := app.Version
			appSource := app.Source
			cNameLen := len(appName)
			if cNameLen > maxNameLen {
				maxNameLen = cNameLen
			}
			cVersionLen := len(appVersion)
			if cVersionLen > maxVersionLen {
				maxVersionLen = cVersionLen
			}
			cSourceLen := len(appSource)
			if cSourceLen > maxSourceLen {
				maxSourceLen = cSourceLen
			}
			cIndexLen := len(fmt.Sprintf("%d", appIndex))
			if cIndexLen > maxNumLen {
				maxNumLen = cIndexLen
			}
			cBinariesLen := len(app.Binaries)
			if cBinariesLen > maxBinariesLen {
				maxBinariesLen = cBinariesLen
			}

		}
		var optList []string
		optMapper := make(map[string]SearchResult)
		for _, app := range appList {
			var optLabel = fmt.Sprintf("%-*d | %-*s | %-*s | %-*s | %-*s", maxNumLen, app.Index, maxNameLen, app.Name, maxVersionLen, app.Version, maxSourceLen, app.Source, maxBinariesLen, app.Binaries)
			optMapper[optLabel] = app
			optList = append(optList, optLabel)
		}
		result, err := pterm.DefaultInteractiveSelect.WithDefaultText("选择一个应用程序进行安装(回车确认,Ctrl+C 取消)").WithOptions(optList).WithMaxHeight(20).Show()

		if err != nil {
			pterm.Error.Println("选择应用程序时出错:", err.Error())
			os.Exit(1)
		}
		selectedApp := optMapper[result]
		if selectedApp.Index > -1 {
			fullName := selectedApp.FullName
			pterm.Info.Println(fmt.Sprintf("您选择了: %s", fullName))

			setupCmd := exec.Command("scoop", "install", fullName)
			setupCmdStr := strings.Join(setupCmd.Args, " ")
			pterm.Info.Printfln("执行安装命令: %s", setupCmdStr)

			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			err := setupCmd.Run()
			if err != nil {
				pterm.Error.Println(fmt.Sprintf("执行命令: %s 时出错:\n%s", setupCmdStr, err.Error()))
				os.Exit(1)
			}
			pterm.Success.Printfln("执行完毕!")
			pterm.Info.Println("==========相关命令==========")
			println("查看该应用程序: ")
			println("scoop info " + fullName)
			println("卸载该应用程序: ")
			println("scoop uninstall " + fullName)
		}

	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

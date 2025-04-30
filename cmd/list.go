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

type ListResult struct {
	Index    int
	Info     string `json:"Info"`
	Source   string `json:"Source"`
	Name     string `json:"Name"`
	Version  string `json:"Version"`
	Updated  string `json:"Updated"`
	FullName string
}
type CmdAction struct {
	Command string
	Desc    string
}

var listCmd = &cobra.Command{
	Use:     "list [query]",
	Short:   "列出所有已安装的应用程序，或与提供的查询匹配的应用程序",
	Long:    `列出所有已安装的应用程序，或与提供的查询匹配的应用程序`,
	Example: `nyaru list`,
	Aliases: []string{"ls"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var query string
		if len(args) > 0 {
			query = args[0]
		}
		spinner, _ := pterm.DefaultSpinner.Start("正在列出已安装的应用程序...")
		pureCmdStr := fmt.Sprintf("scoop list %s", query)
		fmt.Println(pureCmdStr)
		strOutput, listCmdStr, err := utils.RunWithPowerShellCombined("powershell", "-Command", fmt.Sprintf(" %s | ConvertTo-Json -Compress", pureCmdStr))
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", listCmdStr, err.Error()))
			os.Exit(1)
		}
		appList, err := utils.PsDirtyJSONToStructList[ListResult](strOutput)
		if err != nil {
			spinner.Fail(fmt.Sprintf("执行命令 %s 时出错:\n%s", listCmdStr, err.Error()))
			os.Exit(1)
		}
		appSize := len(appList)
		spinner.Success(pureCmdStr)
		pterm.Printfln("获取到 %d 个已安装应用", appSize)
		if appSize == 0 {
			pterm.Warning.Println("没有匹配的已安装应用，不进行任何操作!")
			os.Exit(0)
		}
		maxNumLen := 1
		maxNameLen := 0
		maxVersionLen := 0
		maxSourceLen := 0
		const maxUpdatedLen = 14
		for i, app := range appList {
			appIndex := i + 1
			appSource := app.Source
			appName := app.Name
			appVersion := app.Version
			if strings.HasPrefix(appSource, "http") {
				appList[i].FullName = appName
			} else {
				appList[i].FullName = fmt.Sprintf("%s/%s", appSource, appName)
			}
			appList[i].Index = appIndex
			appList[i].Updated = utils.FormatDateWithWrapper(app.Updated, "/Date(", ")/", "06-01-02 15:04")

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
		}

		var optList []string
		optMapper := make(map[string]ListResult)
		for _, app := range appList {
			var optLabel = fmt.Sprintf("%-*d | %-*s | %-*s | %-*s | %-*s", maxNumLen, app.Index, maxNameLen, app.Name, maxVersionLen, app.Version, maxUpdatedLen, app.Updated, maxSourceLen, app.Source)
			optMapper[optLabel] = app
			optList = append(optList, optLabel)
		}
		selectedAppList, err := pterm.DefaultInteractiveMultiselect.
			WithDefaultText("选取需要操作的应用程序").
			WithOptions(optList).
			WithMaxHeight(20).
			Show()

		if err != nil {
			pterm.Error.Println("选择应用程序时出错:", err.Error())
			os.Exit(1)
		}
		var selectedAppLen = len(selectedAppList)
		pterm.Info.Printfln("选中了 %d 个应用程序", selectedAppLen)
		if selectedAppLen == 0 {
			pterm.Warning.Println("没有选择任何应用程序!退出运行!")
			os.Exit(0)
		}
		cmdActions := []CmdAction{
			{Command: "info", Desc: "查看应用程序信息"},
			{Command: "update", Desc: "更新应用程序"},
			{Command: "uninstall", Desc: "卸载应用程序"},
			{Command: "home", Desc: "打开应用程序主页"},
			{Command: "cache rm", Desc: "删除应用程序的下载缓存"},
			{Command: "cleanup", Desc: "会清理该应用的旧版本（如果存在）"},
			{Command: "reset", Desc: "用于解决冲突，以支持特定应用。例如，如果您安装了“python”和“python27”，则可以使用“scoop reset”在两者之间切换。"},
		}
		actionMap := make(map[string]CmdAction)
		options := []string{}
		for _, cmdAction := range cmdActions {
			optLabel := fmt.Sprintf("%s (%s)", cmdAction.Command, cmdAction.Desc)
			actionMap[optLabel] = cmdAction
			options = append(options, optLabel)
		}
		selAction, _ := pterm.DefaultInteractiveSelect.WithDefaultText("想要进行的操作是?").WithOptions(options).Show()
		pterm.Printfln("选择: %s", selAction)
		command := actionMap[selAction].Command
		pterm.Warning.Printfln("对所有选中应用执行命令:%s", command)
		var sucCount = 0
		var errCount = 0
		for _, selectedApp := range selectedAppList {
			selApp := optMapper[selectedApp]
			fullName := selApp.FullName
			actionCmd := exec.Command("scoop", command, fullName)
			actionCmdStr := strings.Join(actionCmd.Args, " ")
			pterm.Info.Println("开始执行命令:")
			println(actionCmdStr)
			pterm.Info.Println("==========")
			actionCmd.Stdout = os.Stdout
			actionCmd.Stderr = os.Stderr
			err := actionCmd.Run()
			if err != nil {
				errCount++
				pterm.Error.Println(fmt.Sprintf("执行命令: %s 时出错:\n%s", actionCmd, err.Error()))
			} else {
				sucCount++
				pterm.Success.Printfln("执行完毕!")
			}
		}
		pterm.Info.Println("==========")
		pterm.Info.Printfln("成功 %d 个，失败 %d 个", sucCount, errCount)

	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

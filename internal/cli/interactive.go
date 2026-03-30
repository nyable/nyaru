package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "进入交互模式主菜单 (别名: gui, tui)",
	Long:    `显示一个主菜单，允许你交互式地执行各种 Scoop 命令。`,
	Aliases: []string{"gui", "tui"},
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)
		for {
			choice, err := tui.RunMainMenu()
			if err != nil {
				tui.PrintError(fmt.Sprintf("TUI Error: %v", err))
				os.Exit(1)
			}

			if choice == "" {
				return
			}

			switch choice {
			case "Search":
				fmt.Print("输入要搜索的内容: ")
				var query string
				fmt.Scanln(&query)
				SearchAction(query)
			case "Install":
				fmt.Print("输入要安装的应用名(多个用空格分隔): ")
				line, _ := reader.ReadString('\n')
				apps := strings.Fields(strings.TrimSpace(line))
				if len(apps) > 0 {
					InstallAction(apps)
				}
			case "Uninstall":
				fmt.Print("输入要卸载的应用名(多个用空格分隔): ")
				line, _ := reader.ReadString('\n')
				apps := strings.Fields(strings.TrimSpace(line))
				if len(apps) > 0 {
					UninstallAction(apps)
				}
			case "Info":
				fmt.Print("输入要查看的应用名: ")
				var appName string
				fmt.Scanln(&appName)
				InfoAction(appName)
			case "List":
				ListAction()
			case "Status":
				StatusAction()
			case "Update All":
				UpdateAction()
			case "Buckets":
				BucketAction()
			case "Cache":
				CacheAction()
			case "Exit":
				return
			}

			fmt.Println("\n按回车键返回主菜单...")
			fmt.Scanln()
		}
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}


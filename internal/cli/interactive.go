package cli

import (
	"fmt"
	"os"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "进入交互模式主菜单 (别名: i, gui, tui)",
	Long:    `显示一个主菜单，允许你交互式地执行各种 Scoop 命令。`,
	Aliases: []string{"i", "gui", "tui"},
	Run: func(cmd *cobra.Command, args []string) {
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

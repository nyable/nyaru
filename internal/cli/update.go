package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "等同于 scoop update",
	Long:  "等同于 scoop update",
	Run: func(cmd *cobra.Command, args []string) {
		pm := core.GetManager(config.GetActiveMode())

		tui.PrintInfo("开始执行更新命令...")
		if len(args) == 0 {
			fmt.Println("scoop update")
		} else {
			fmt.Printf("scoop update %s\n", strings.Join(args, " "))
		}

		if err := pm.Update(args...); err != nil {
			tui.PrintError(fmt.Sprintf("更新失败: %v", err))
			os.Exit(1)
		}
		tui.PrintSuccess("更新完成!")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

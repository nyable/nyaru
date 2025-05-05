package cmd

import (
	"fmt"
	"os"

	"github.com/nyable/nyaru/internal/utils"
	"github.com/spf13/cobra"
)

var AppVersion = "0.0.4"

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "查询版本信息(别名:v/ver)",
	Long:    `查询版本信息`,
	Example: `nyaru version`,
	Aliases: []string{"v", "ver"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nyaru v%s\n\n", AppVersion)
		strOutput, cmdStr, err := utils.RunWithPowerShellCombined("powershell", "-Command", "scoop -v")
		if err != nil {
			println(fmt.Sprintf("执行命令 %s 时出错:\n%s", cmdStr, err.Error()))
			os.Exit(1)
		}
		fmt.Println(strOutput)

	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

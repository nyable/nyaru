package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "等同于 scoop update",
	Long:  "等同于 scoop update",
	Run: func(cmd *cobra.Command, args []string) {
		update := exec.Command("scoop", "update")
		pterm.Info.Println("开始执行命令:")
		println(strings.Join(update.Args, " "))
		update.Stdout = os.Stdout
		update.Stderr = os.Stderr
		update.Run()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

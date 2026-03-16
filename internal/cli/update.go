package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "等同于 scoop update",
	Long:  "等同于 scoop update",
	Run: func(cmd *cobra.Command, args []string) {
		update := exec.Command("scoop", "update")
		tui.PrintInfo("开始执行命令:")
		fmt.Println(strings.Join(update.Args, " "))
		update.Stdout = os.Stdout
		update.Stderr = os.Stderr
		update.Run()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

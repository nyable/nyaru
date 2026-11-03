package cli

import (
	"fmt"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "进入交互式主菜单",
	Args:    cobra.NoArgs,
	Long:    `显示主菜单，交互式执行搜索、安装、卸载、更新、存储桶和缓存操作。`,
	Aliases: []string{"gui", "tui"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return InteractiveAction()
	},
}

func InteractiveAction() error {
	if flagNoTUI {
		return fmt.Errorf("interactive 模式不能与 --no-tui 同时使用")
	}
	if flagJSON {
		return fmt.Errorf("interactive 模式不能与 --json 同时使用")
	}
	if !isTTY() {
		return fmt.Errorf("交互模式需要 TTY；请使用具体子命令或 --no-tui")
	}
	pm := newBackend()
	return runTUI(tui.Options{
		Backend: pm,
		Start:   tui.StartMenu,
		Version: AppVersion,
	})
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

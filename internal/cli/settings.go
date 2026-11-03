package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:     "settings",
	Short:   "查看和修改 nyaru 自身的设置",
	Long:    `查看和修改 nyaru 自身的设置。Scoop 的配置请使用 nyaru config。`,
	Aliases: []string{"prefs", "preferences"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tuiEnabled() {
			return runTUI(tui.Options{
				Backend: newBackend(),
				Start:   tui.StartSettings,
				Version: AppVersion,
			})
		}
		return settingsListAction()
	},
}

var settingsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "列出全部 nyaru 设置",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls", "show"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return settingsListAction()
	},
}

var settingsGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "读取一个 nyaru 设置项",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := config.Get().SettingValue(args[0])
		if err != nil {
			return err
		}
		if flagJSON {
			return printJSON(map[string]string{"key": args[0], "value": value})
		}
		fmt.Println(value)
		return nil
	},
}

var settingsSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "修改一个 nyaru 设置项",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if err := cfg.SetSetting(args[0], args[1]); err != nil {
			return err
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		if flagJSON {
			return printJSON(settingsSnapshot(cfg))
		}
		value, _ := cfg.SettingValue(args[0])
		tui.PrintSuccess(fmt.Sprintf("已设置 %s = %s", args[0], value))
		return nil
	},
}

var settingsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "恢复 nyaru 默认设置",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !confirm("确认恢复 nyaru 的默认设置？") {
			tui.PrintInfo("已取消恢复默认设置")
			return nil
		}
		cfg := config.Get()
		*cfg = *config.DefaultConfig()
		if err := cfg.Save(); err != nil {
			return err
		}
		if flagJSON {
			return printJSON(settingsSnapshot(cfg))
		}
		tui.PrintSuccess("已恢复默认设置")
		return nil
	},
}

var settingsPathCmd = &cobra.Command{
	Use:   "path",
	Short: "显示 nyaru 配置文件路径",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}
		if flagJSON {
			return printJSON(map[string]string{"path": path})
		}
		fmt.Println(path)
		return nil
	},
}

type settingsOutput struct {
	Path          string `json:"path"`
	Mode          string `json:"mode"`
	ConfirmBatch  bool   `json:"confirm_batch"`
	StatusTimeout int    `json:"status_timeout"`
}

func settingsSnapshot(cfg *config.Config) settingsOutput {
	path, _ := config.Path()
	return settingsOutput{
		Path:          path,
		Mode:          cfg.Mode,
		ConfirmBatch:  cfg.ConfirmBatch,
		StatusTimeout: cfg.StatusTimeoutSeconds(),
	}
}

func settingsListAction() error {
	cfg := config.Get()
	if flagJSON {
		return printJSON(settingsSnapshot(cfg))
	}
	path, _ := config.Path()
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "配置项\t值\t说明")
	fmt.Fprintf(w, "mode\t%s\t查询后端（auto / sfsu / scoop）\n", cfg.Mode)
	fmt.Fprintf(w, "confirm_batch\t%t\t批量操作前是否确认\n", cfg.ConfirmBatch)
	timeout := cfg.StatusTimeoutSeconds()
	timeoutText := fmt.Sprintf("%d 秒", timeout)
	if timeout == 0 {
		timeoutText = "不自动清除"
	}
	fmt.Fprintf(w, "status_timeout\t%s\t底部提示自动清除时间\n", timeoutText)
	_ = w.Flush()
	fmt.Printf("\n配置文件: %s\n", path)
	return nil
}

func init() {
	settingsCmd.AddCommand(settingsListCmd)
	settingsCmd.AddCommand(settingsGetCmd)
	settingsCmd.AddCommand(settingsSetCmd)
	settingsCmd.AddCommand(settingsResetCmd)
	settingsCmd.AddCommand(settingsPathCmd)
	rootCmd.AddCommand(settingsCmd)
}

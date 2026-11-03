package cli

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var AppVersion = "dev"

type versionInfo struct {
	Nyaru  string `json:"nyaru"`
	Mode   string `json:"mode"`
	Scoop  string `json:"scoop,omitempty"`
	Sfsu   string `json:"sfsu,omitempty"`
	Method string `json:"method,omitempty"`
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "查询版本信息",
	Args:    cobra.NoArgs,
	Aliases: []string{"v", "ver"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return VersionAction()
	},
}

func VersionAction() error {
	info := versionInfo{
		Nyaru:  AppVersion,
		Mode:   config.Get().ActiveMode(),
		Scoop:  commandVersion("scoop", "-v"),
		Sfsu:   commandVersion("sfsu", "--version"),
		Method: installMethod(),
	}
	if flagJSON {
		return printJSON(info)
	}
	tui.PrintInfo("nyaru v" + info.Nyaru)
	if info.Method != "" {
		tui.PrintInfo("安装方式: " + info.Method)
	}
	tui.PrintInfo("当前后端: " + info.Mode)
	if info.Scoop != "" {
		tui.PrintInfo("scoop: " + info.Scoop)
	} else {
		tui.PrintWarning("未检测到 scoop")
	}
	if info.Sfsu != "" {
		tui.PrintInfo("sfsu: " + info.Sfsu)
	}
	return nil
}

func commandVersion(name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := core.Command{Name: name, Args: args}
	out, err := cmd.ExecContext(ctx).CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func installMethod() string {
	// The self-update package provides the authoritative detector. Version
	// keeps this lightweight check to avoid an extra Scoop process.
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	lower := strings.ToLower(exe)
	if strings.Contains(lower, `\scoop\apps\nyaru\`) || strings.Contains(lower, `\scoop\shims\nyaru.exe`) {
		return "scoop"
	}
	return "standalone"
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

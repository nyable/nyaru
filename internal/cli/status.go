package cli

import (
	"context"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "检查已安装应用的更新状态",
	Args:  cobra.NoArgs,
	Long:  `检查已安装应用的更新状态，并可在 TUI 中选择更新。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return StatusAction()
	},
}

func StatusAction() error {
	pm := newBackend()
	if !tuiEnabled() {
		apps, err := pm.Status(context.Background())
		if err != nil {
			return err
		}
		if flagJSON {
			return printJSON(apps)
		}
		printAppsTable(apps)
		return nil
	}
	return runTUI(tui.Options{
		Backend: pm,
		Start:   tui.StartUpdates,
		Version: AppVersion,
	})
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

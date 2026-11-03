package cli

import (
	"context"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [query]",
	Short: "列出已安装的应用程序",
	Long:  `列出已安装的应用程序，可按名称过滤。`,
	Example: `  nyaru list
  nyaru list git`,
	Aliases: []string{"ls"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}
		return ListAction(query)
	},
}

func ListAction(query string) error {
	pm := newBackend()
	if !tuiEnabled() {
		apps, err := pm.List(context.Background(), query)
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
		Start:   tui.StartInstalled,
		Query:   query,
		Version: AppVersion,
	})
}

func init() {
	rootCmd.AddCommand(listCmd)
}

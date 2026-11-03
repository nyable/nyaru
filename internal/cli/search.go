package cli

import (
	"context"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "搜索可安装的应用程序",
	Long:    `搜索与查询匹配的应用。不带 query 时列出所有可安装应用。`,
	Example: `  nyaru search aria2`,
	Aliases: []string{"find", "query", "s"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}
		return SearchAction(query)
	},
}

func SearchAction(query string) error {
	pm := newBackend()
	if !tuiEnabled() {
		apps, err := pm.Search(context.Background(), query)
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
		Start:   tui.StartSearch,
		Query:   query,
		Version: AppVersion,
	})
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

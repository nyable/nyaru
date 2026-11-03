package cli

import (
	"context"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info <app>",
	Short:   "查看应用详细信息",
	Aliases: []string{"detail"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return InfoAction(args[0])
	},
}

func InfoAction(app string) error {
	pm := newBackend()
	detail, err := pm.Info(context.Background(), app)
	if err != nil {
		return err
	}
	if flagJSON {
		return printJSON(detail)
	}
	printDetail(detail)
	return nil
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var exportConfig bool

var exportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "导出已安装应用、存储桶和配置",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scoopArgs := []string{"export"}
		if exportConfig {
			scoopArgs = append(scoopArgs, "--config")
		}
		out, err := runCommandStdout(scoopCommand(scoopArgs...))
		if err != nil {
			return err
		}
		if len(args) == 0 {
			fmt.Print(out)
			return nil
		}
		if err := os.WriteFile(args[0], []byte(out), 0o644); err != nil {
			return err
		}
		fmt.Println("已导出到 " + args[0])
		return nil
	},
}

var importCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "从 Scoopfile 导入应用、存储桶和配置",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCommand(scoopCommand("import", args[0]))
	},
}

func init() {
	exportCmd.Flags().BoolVarP(&exportConfig, "config", "c", false, "同时导出 Scoop 配置")
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}

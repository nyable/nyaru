package cli

import (
	"context"

	"github.com/nyable/nyaru/internal/tui"
	"github.com/spf13/cobra"
)

var bucketListCmd = &cobra.Command{
	Use:     "list",
	Short:   "列出 Scoop 存储桶",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return BucketAction()
	},
}

func BucketAction() error {
	pm := newBackend()
	if !tuiEnabled() {
		buckets, err := pm.BucketList(context.Background())
		if err != nil {
			return err
		}
		if flagJSON {
			return printJSON(buckets)
		}
		printBucketsTable(buckets)
		return nil
	}
	return runTUI(tui.Options{
		Backend: pm,
		Start:   tui.StartBuckets,
		Version: AppVersion,
	})
}

var bucketAddCmd = &cobra.Command{
	Use:     "add <name> [source]",
	Short:   "添加存储桶",
	Example: `  nyaru bucket add extras`,
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := ""
		if len(args) > 1 {
			source = args[1]
		}
		return BucketAddAction(args[0], source)
	},
}

func BucketAddAction(name, source string) error {
	pm := newBackend()
	tui.PrintInfo("正在添加存储桶: " + name)
	if err := runCommand(pm.BucketAdd(name, source)); err != nil {
		return err
	}
	tui.PrintSuccess("存储桶添加完成")
	return nil
}

var bucketRemoveCmd = &cobra.Command{
	Use:     "rm <name>",
	Short:   "删除存储桶",
	Aliases: []string{"remove"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return BucketRemoveAction(args[0])
	},
}

func BucketRemoveAction(name string) error {
	if !confirm("确认删除存储桶 " + name + "？") {
		tui.PrintInfo("已取消删除")
		return nil
	}
	pm := newBackend()
	tui.PrintInfo("正在删除存储桶: " + name)
	if err := runCommand(pm.BucketRemove(name)); err != nil {
		return err
	}
	tui.PrintSuccess("存储桶删除完成")
	return nil
}

var bucketKnownCmd = &cobra.Command{
	Use:   "known",
	Short: "列出 Scoop 已知的官方存储桶",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScoopOutput("bucket", "known")
	},
}

var bucketCmd = &cobra.Command{
	Use:     "bucket",
	Short:   "管理 Scoop 存储桶",
	Aliases: []string{"bt"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return BucketAction()
	},
}

func init() {
	bucketCmd.AddCommand(bucketListCmd)
	bucketCmd.AddCommand(bucketAddCmd)
	bucketCmd.AddCommand(bucketRemoveCmd)
	bucketCmd.AddCommand(bucketKnownCmd)
	rootCmd.AddCommand(bucketCmd)
}

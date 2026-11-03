package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/models"
)

var (
	flagJSON    bool
	flagNoTUI   bool
	flagYes     bool
	flagNoColor bool
	flagVerbose bool
)

func newBackend() core.Backend {
	cfg := config.Get()
	mode := core.ParseMode(cfg.ActiveMode())
	return core.NewBackend(mode, core.NewRunner(mode))
}

func isTTY() bool {
	out, err := os.Stdout.Stat()
	if err != nil || out.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	in, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return in.Mode()&os.ModeCharDevice != 0
}

func tuiEnabled() bool {
	return !flagJSON && !flagNoTUI && isTTY()
}

func confirm(prompt string) bool {
	if flagYes {
		return true
	}
	if !isTTY() {
		fmt.Fprintln(os.Stderr, "非交互环境下请使用 --yes 确认操作")
		return false
	}
	fmt.Printf("%s (y/n): ", prompt)
	var answer string
	fmt.Scanln(&answer)
	return strings.EqualFold(strings.TrimSpace(answer), "y")
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func runCommand(cmd core.Command) error {
	if flagVerbose {
		fmt.Fprintln(os.Stderr, "执行:", cmd.String())
	}
	c := cmd.Exec()
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func runCommandOutput(cmd core.Command) (string, error) {
	c := cmd.ExecContext(context.Background())
	out, err := c.CombinedOutput()
	return string(out), err
}

// runCommandStdout captures stdout without mixing stderr into the result.
// Used when the output is saved to a file, such as scoop export.
func runCommandStdout(cmd core.Command) (string, error) {
	c := cmd.ExecContext(context.Background())
	var stderr bytes.Buffer
	c.Stderr = &stderr
	out, err := c.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return string(out), fmt.Errorf("%s: %w", msg, err)
		}
		return string(out), err
	}
	return string(out), nil
}

func scoopCommand(args ...string) core.Command {
	return core.Command{Name: "scoop", Args: args}
}

// runScoopOutput prints command output even when the command fails, so
// warnings such as "not installed" are not lost behind the exit status.
func runScoopOutput(args ...string) error {
	out, err := runCommandOutput(scoopCommand(args...))
	if out != "" {
		fmt.Print(out)
	}
	return err
}

func configConfirmBatch() bool {
	return config.Get().ConfirmBatch
}

func printAppsTable(apps []models.AppInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "名称\t版本\t来源\t范围\t状态")
	for _, app := range apps {
		status := app.Status.Label()
		if status == "" {
			status = "-"
		}
		scope := "-"
		if app.Installed {
			scope = "本地"
			if app.Global {
				scope = "全局"
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", app.Name, app.DisplayVersion(), app.Origin(), scope, status)
	}
	_ = w.Flush()
}

func printBucketsTable(buckets []models.BucketResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "名称\t来源\t清单数\t更新时间")
	for _, b := range buckets {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", b.Name, b.Source, b.Manifests, b.Updated.DateTime)
	}
	_ = w.Flush()
}

func printCacheTable(cache []models.CacheResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "应用\t版本\t大小")
	for _, c := range cache {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.App, strings.Join(c.Versions, ", "), c.FormatSize)
	}
	_ = w.Flush()
}

func printDetail(detail models.AppDetail) {
	if detail.Name == "" {
		fmt.Println("未找到应用信息")
		return
	}
	for _, f := range detail.Fields() {
		fmt.Printf("%-12s %s\n", f.Label, f.Value)
	}
	if detail.Description != "" {
		fmt.Printf("\n%s\n", detail.Description)
	}
	if detail.Notes != "" {
		fmt.Printf("\n%s\n", detail.Notes)
	}
}

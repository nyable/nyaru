package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/core/parse"
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/paths"
)

type Method string

const (
	MethodScoop      Method = "scoop"
	MethodStandalone Method = "standalone"
	MethodUnknown    Method = "unknown"
)

type Status struct {
	Method          Method `json:"method"`
	Managed         bool   `json:"managed"`
	Current         string `json:"current"`
	Latest          string `json:"latest,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
	Path            string `json:"path,omitempty"`
}

type ApplyOptions struct {
	Global bool
	Force  bool
}

// LogPath returns the location used by the self-update helper.
func LogPath() (string, error) {
	return paths.SelfUpdateLog()
}

// Detect determines whether nyaru is managed by Scoop.
func Detect(ctx context.Context) (Method, string) {
	if path, err := scoopPrefix(ctx); err == nil && strings.TrimSpace(path) != "" {
		return MethodScoop, strings.TrimSpace(path)
	}
	if exe, err := os.Executable(); err == nil {
		lower := strings.ToLower(exe)
		if strings.Contains(lower, `\scoop\apps\nyaru\`) ||
			strings.Contains(lower, `\scoop\shims\nyaru.exe`) {
			return MethodScoop, exe
		}
	}
	return MethodStandalone, ""
}

func scoopPrefix(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	runner := core.NewRunner(core.ModeScoop)
	return runner.ScoopOutput(ctx, "prefix", "nyaru")
}

// Check compares the current version with the version available in Scoop or
// the latest GitHub release for standalone installs.
func Check(ctx context.Context, current string) (Status, error) {
	method, path := Detect(ctx)
	status := Status{
		Method:  method,
		Managed: method == MethodScoop,
		Current: current,
		Path:    path,
	}
	switch method {
	case MethodScoop:
		latest, err := scoopLatest(ctx)
		if err != nil {
			return status, err
		}
		status.Latest = latest
		status.UpdateAvailable = latest != "" && current != "dev" && versionLess(current, latest)
		return status, nil
	default:
		latest, err := githubLatest(ctx)
		if err != nil {
			return status, err
		}
		status.Latest = latest
		status.UpdateAvailable = latest != "" && current != "dev" && versionLess(current, latest)
		return status, nil
	}
}

func scoopLatest(ctx context.Context) (string, error) {
	runner := core.NewRunner(core.ModeScoop)
	raw, err := runner.ScoopStatusJSON(ctx)
	if err != nil {
		return "", err
	}
	apps, err := parse.ScoopStatus(raw)
	if err != nil {
		return "", err
	}
	for _, app := range apps {
		if strings.EqualFold(app.Name, "nyaru") {
			if app.Status == models.StatusHeld {
				return "", fmt.Errorf("nyaru 已被 Scoop 锁定，请先运行 scoop unhold nyaru")
			}
			if app.AvailableVersion != "" {
				return app.AvailableVersion, nil
			}
			return app.Version, nil
		}
	}
	// Not present in status means Scoop sees nyaru as current.
	return "", nil
}

type release struct {
	TagName string `json:"tag_name"`
}

func githubLatest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/nyable/nyaru/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API 返回 %s", resp.Status)
	}
	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(r.TagName), "v"), nil
}

// Apply starts a detached helper that waits for the current process to exit
// before running scoop update nyaru. This is required on Windows because a
// running .exe cannot be replaced.
func Apply(ctx context.Context, status Status, opts ApplyOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if status.Method != MethodScoop {
		return fmt.Errorf("仅支持 Scoop 托管的 nyaru；独立安装请从 Releases 下载新版本")
	}
	if runtime.GOOS != "windows" {
		return fmt.Errorf("自更新 helper 仅支持 Windows")
	}
	ps, err := exec.LookPath("powershell")
	if err != nil {
		ps, err = exec.LookPath("pwsh")
	}
	if err != nil {
		return fmt.Errorf("未找到 powershell 或 pwsh: %w", err)
	}
	logPath, err := paths.SelfUpdateLog()
	if err != nil {
		return err
	}
	script := helperScript(os.Getpid(), opts, logPath)
	cmd := exec.Command(ps, "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func helperScript(pid int, opts ApplyOptions, logPath string) string {
	updateArgs := "nyaru"
	if opts.Global {
		updateArgs += " --global"
	}
	if opts.Force {
		updateArgs += " --force"
	}
	return fmt.Sprintf(`
$log = %s
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $log) | Out-Null
try { Start-Transcript -Path $log -Append | Out-Null } catch {}
$target = %d
while (Get-Process -Id $target -ErrorAction SilentlyContinue) {
    Start-Sleep -Milliseconds 200
}
Write-Host "正在刷新 Scoop 与 buckets..."
scoop update
Write-Host "正在更新 nyaru..."
scoop update %s
$code = $LASTEXITCODE
if ($code -ne 0) {
    Write-Host "nyaru 更新失败（退出码 $code），详见 $log"
}
try { Stop-Transcript | Out-Null } catch {}
exit $code
`, psQuote(logPath), pid, updateArgs)
}

func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func versionLess(current, latest string) bool {
	c := strings.Split(strings.TrimPrefix(strings.TrimSpace(current), "v"), ".")
	l := strings.Split(strings.TrimPrefix(strings.TrimSpace(latest), "v"), ".")
	for i := 0; i < len(c) || i < len(l); i++ {
		var cn, ln int
		if i < len(c) {
			cn, _ = strconv.Atoi(nonDigits(c[i]))
		}
		if i < len(l) {
			ln, _ = strconv.Atoi(nonDigits(l[i]))
		}
		if cn != ln {
			return cn < ln
		}
	}
	return false
}

func nonDigits(s string) string {
	out := strings.Builder{}
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out.WriteRune(r)
		}
	}
	if out.Len() == 0 {
		return "0"
	}
	return out.String()
}

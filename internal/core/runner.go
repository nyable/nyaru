package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Runner executes external commands. The configured mode lives here so every
// query has one predictable backend instead of re-detecting sfsu per call.
type Runner struct {
	Mode           Mode
	LookPath       func(string) (string, error)
	PowerShellPath string
	QueryTimeout   time.Duration
}

func NewRunner(mode Mode) *Runner {
	return &Runner{
		Mode:         mode,
		LookPath:     exec.LookPath,
		QueryTimeout: 2 * time.Minute,
	}
}

func (r *Runner) HasSfsu() bool {
	if r == nil {
		return false
	}
	_, err := r.LookPath("sfsu")
	return err == nil
}

func (r *Runner) powershell() (string, error) {
	if r.PowerShellPath != "" {
		return r.PowerShellPath, nil
	}
	for _, name := range []string{"powershell", "pwsh"} {
		if p, err := r.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("未找到 PowerShell，请确认 powershell 或 pwsh 在 PATH 中")
}

func (r *Runner) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if r.QueryTimeout <= 0 {
		return ctx, func() {}
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, r.QueryTimeout)
}

func (r *Runner) run(ctx context.Context, name string, args []string, env []string) (string, string, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = envWith(os.Environ(), env...)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("命令超时或被取消: %s", cmd.String())
	}
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			return stdout.String(), stderr.String(), fmt.Errorf("%s: %w", msg, err)
		}
		return stdout.String(), stderr.String(), err
	}
	return stdout.String(), stderr.String(), nil
}

func envWith(base []string, values ...string) []string {
	keys := make(map[string]struct{}, len(values))
	for _, v := range values {
		if i := strings.IndexByte(v, '='); i >= 0 {
			keys[v[:i]] = struct{}{}
		}
	}
	out := make([]string, 0, len(base)+len(values))
	for _, v := range base {
		if i := strings.IndexByte(v, '='); i >= 0 {
			if _, replace := keys[v[:i]]; replace {
				continue
			}
		}
		out = append(out, v)
	}
	return append(out, values...)
}

func envWithout(base []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(base))
	for _, value := range base {
		if len(value) >= len(prefix) && strings.EqualFold(value[:len(prefix)], prefix) {
			continue
		}
		out = append(out, value)
	}
	return out
}

// isWindowsPowerShell reports whether path is Windows PowerShell 5.1
// (powershell.exe) rather than PowerShell 7 (pwsh).
func isWindowsPowerShell(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return name == "powershell" || name == "powershell.exe"
}

// powershellEnv returns the environment for the given PowerShell executable.
// Windows PowerShell 5.1 inherits PSModulePath when nyaru is launched from
// PowerShell 7. That makes 5.1 look in PowerShell 7's module directories first
// and breaks cmdlet autoloading (for example Get-FileHash, which Scoop uses).
// Removing PSModulePath lets Windows PowerShell build its own default paths.
// Other shells keep the inherited environment unchanged.
func powershellEnv(ps string) []string {
	if runtime.GOOS != "windows" || !isWindowsPowerShell(ps) {
		return nil
	}
	return envWithout(os.Environ(), "PSModulePath")
}

// SfsuJSON runs sfsu with arguments and returns stdout. Arguments are passed
// as a vector; no shell string is constructed.
func (r *Runner) SfsuJSON(ctx context.Context, args ...string) (string, error) {
	if !r.HasSfsu() {
		return "", fmt.Errorf("sfsu 未安装，无法使用 sfsu 后端")
	}
	full := append([]string{}, args...)
	full = append(full, "--json")
	out, _, err := r.run(ctx, "sfsu", full, nil)
	return out, err
}

// ScoopJSON runs a fixed, whitelisted PowerShell query and returns stdout.
// User input is passed through NYARU_QUERY instead of being interpolated into
// the command string.
func (r *Runner) ScoopJSON(ctx context.Context, script, query string) (string, error) {
	ps, err := r.powershell()
	if err != nil {
		return "", err
	}
	env := powershellEnv(ps)
	if env == nil {
		env = os.Environ()
	}
	env = envWith(env, "NYARU_QUERY="+query)
	out, _, err := r.run(ctx, ps, []string{
		"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", script,
	}, env)
	return out, err
}

const (
	scoopSearchScript = `$result = if ($env:NYARU_QUERY) { scoop search $env:NYARU_QUERY } else { scoop search }; $result | ConvertTo-Json -Compress`
	scoopListScript   = `$result = if ($env:NYARU_QUERY) { scoop list $env:NYARU_QUERY } else { scoop list }; $result | ConvertTo-Json -Compress`
	scoopStatusScript = `$globalApps = @{}; scoop list | ForEach-Object { $globalApps[$_.Name] = [bool]$_.global }; $status = scoop status; $status | ForEach-Object { $isGlobal = $false; if ($globalApps.ContainsKey($_.Name)) { $isGlobal = $globalApps[$_.Name] }; $_ | Add-Member -NotePropertyName Global -NotePropertyValue $isGlobal -Force }; $status | ConvertTo-Json -Compress`
	scoopBucketScript = `scoop bucket list | ConvertTo-Json -Compress`
	// scoopOldVersionsScript scans local and global app directories. `current`
	// is a junction, so its target identifies the live version. Apps whose
	// current junction cannot be resolved are skipped: guessing could delete
	// a running version.
	scoopOldVersionsScript = `
$localRoot = if ($env:SCOOP) { $env:SCOOP } else { Join-Path $env:USERPROFILE 'scoop' }
$globalRoot = if ($env:SCOOP_GLOBAL) { $env:SCOOP_GLOBAL } else { Join-Path $env:ProgramData 'scoop' }
$roots = @(
  [pscustomobject]@{ Path = (Join-Path $localRoot 'apps'); Global = $false },
  [pscustomobject]@{ Path = (Join-Path $globalRoot 'apps'); Global = $true }
)
$result = @()
foreach ($root in $roots) {
  if (-not (Test-Path -LiteralPath $root.Path)) { continue }
  foreach ($appDir in @(Get-ChildItem -LiteralPath $root.Path -Directory -Force -ErrorAction SilentlyContinue)) {
    $currentPath = Join-Path $appDir.FullName 'current'
    if (-not (Test-Path -LiteralPath $currentPath)) { continue }
    $link = Get-Item -LiteralPath $currentPath -Force -ErrorAction SilentlyContinue
    if (-not $link -or -not $link.Target) { continue }
    $target = [string](@($link.Target)[0])
    $current = Split-Path -Leaf ($target.TrimEnd([char[]]@('\', '/')))
    if (-not $current) { continue }
    $old = @()
    $total = [int64]0
    foreach ($versionDir in @(Get-ChildItem -LiteralPath $appDir.FullName -Directory -Force -ErrorAction SilentlyContinue)) {
      if ($versionDir.Name -eq 'current' -or $versionDir.Name -eq $current) { continue }
      $sum = Get-ChildItem -LiteralPath $versionDir.FullName -Recurse -File -Force -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum
      $size = [int64]0
      if ($null -ne $sum.Sum) { $size = [int64]$sum.Sum }
      $old += [pscustomobject]@{ version = $versionDir.Name; length = $size }
      $total += $size
    }
    if ($old.Count -eq 0) { continue }
    $result += [pscustomobject]@{
      name = $appDir.Name
      current = $current
      global = $root.Global
      versions = $old
      length = $total
    }
  }
}
$json = $result | ConvertTo-Json -Compress -Depth 5
if (-not $json) { $json = '[]' }
$json`
)

func (r *Runner) ScoopSearchJSON(ctx context.Context, query string) (string, error) {
	return r.ScoopJSON(ctx, scoopSearchScript, query)
}

func (r *Runner) ScoopListJSON(ctx context.Context, query string) (string, error) {
	return r.ScoopJSON(ctx, scoopListScript, query)
}

func (r *Runner) ScoopStatusJSON(ctx context.Context) (string, error) {
	return r.ScoopJSON(ctx, scoopStatusScript, "")
}

func (r *Runner) ScoopBucketListJSON(ctx context.Context) (string, error) {
	return r.ScoopJSON(ctx, scoopBucketScript, "")
}

func (r *Runner) ScoopOldVersionsJSON(ctx context.Context) (string, error) {
	return r.ScoopJSON(ctx, scoopOldVersionsScript, "")
}

// ScoopOutput runs a scoop subcommand directly and captures its output.
// Used for textual commands such as info, config and cache show. If scoop is
// only available as a PowerShell function, it retries through PowerShell.
func (r *Runner) ScoopOutput(ctx context.Context, args ...string) (string, error) {
	// Scoop is a .cmd shim on Windows; PowerShell is the reliable entry point.
	if runtime.GOOS == "windows" {
		if ps, err := r.powershell(); err == nil {
			out, _, err := r.run(ctx, ps, []string{
				"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
				"-Command", psScoopScript(args),
			}, powershellEnv(ps))
			return out, err
		}
	}
	out, _, err := r.run(ctx, "scoop", args, nil)
	if err == nil {
		return out, err
	}
	if !errors.Is(err, exec.ErrNotFound) && !strings.Contains(strings.ToLower(err.Error()), "executable file not found") {
		return out, err
	}
	ps, psErr := r.powershell()
	if psErr != nil {
		return out, err
	}
	out, _, err = r.run(ctx, ps, []string{
		"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", psScoopScript(args),
	}, powershellEnv(ps))
	return out, err
}

func psScoopScript(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, psQuote(arg))
	}
	script := "& scoop"
	if len(quoted) > 0 {
		script += " " + strings.Join(quoted, " ")
	}
	// Preserve a real exit code; command-not-found leaves $LASTEXITCODE null.
	return script + `; $code = $LASTEXITCODE; if ($null -eq $code) { $code = if ($?) { 0 } else { 1 } }; exit $code`
}

func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

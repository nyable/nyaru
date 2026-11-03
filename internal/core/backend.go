package core

import (
	"context"
	"os/exec"
	"runtime"
	"strings"

	"github.com/nyable/nyaru/internal/models"
)

// Mode selects the backend used for reads.
type Mode string

const (
	ModeAuto  Mode = "auto"
	ModeSfsu  Mode = "sfsu"
	ModeScoop Mode = "scoop"
)

// ParseMode normalizes a configured mode and falls back to auto.
func ParseMode(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sfsu":
		return ModeSfsu
	case "scoop":
		return ModeScoop
	default:
		return ModeAuto
	}
}

type InstallOptions struct {
	Global        bool
	Arch          string
	NoCache       bool
	Independent   bool
	SkipHashCheck bool
	NoUpdateScoop bool
}

type UninstallOptions struct {
	Global bool
	Purge  bool
}

type UpdateOptions struct {
	Global        bool
	Force         bool
	Independent   bool
	NoCache       bool
	SkipHashCheck bool
	All           bool
}

type HoldOptions struct {
	Global bool
}

type CleanupOptions struct {
	Global bool
	All    bool
	Cache  bool
}

// Reader is the read-only side of a package manager backend.
type Reader interface {
	Search(ctx context.Context, query string) ([]models.AppInfo, error)
	List(ctx context.Context, query string) ([]models.AppInfo, error)
	Status(ctx context.Context) ([]models.AppInfo, error)
	Info(ctx context.Context, app string) (models.AppDetail, error)
	BucketList(ctx context.Context) ([]models.BucketResult, error)
	CacheList(ctx context.Context) ([]models.CacheResult, error)
	OldVersions(ctx context.Context) ([]models.OldVersion, error)
}

// Writer builds interactive scoop commands. Building and running are separate
// so the TUI can suspend itself with tea.ExecProcess.
type Writer interface {
	Install(apps []string, opts InstallOptions) Command
	Uninstall(apps []string, opts UninstallOptions) Command
	Update(apps []string, opts UpdateOptions) Command
	Hold(apps []string, opts HoldOptions) Command
	Unhold(apps []string, opts HoldOptions) Command
	BucketAdd(name, source string) Command
	BucketRemove(name string) Command
	CacheRemove(apps []string, all bool) Command
	Cleanup(apps []string, opts CleanupOptions) Command
}

// Backend is a complete package manager backend.
type Backend interface {
	Name() string
	Reader
	Writer
}

// Command is an external process description.
type Command struct {
	Name string
	Args []string
}

func (c Command) Exec() *exec.Cmd {
	return c.exec(nil)
}

func (c Command) ExecContext(ctx context.Context) *exec.Cmd {
	return c.exec(ctx)
}

// exec runs scoop through PowerShell on Windows. Scoop is distributed as a
// .cmd shim; PowerShell also gives us predictable argument quoting for bucket
// URLs and app names. Other commands run directly.
func (c Command) exec(ctx context.Context) *exec.Cmd {
	if runtime.GOOS == "windows" && strings.EqualFold(c.Name, "scoop") {
		if ps, err := exec.LookPath("powershell"); err == nil {
			return c.powerShellCmd(ctx, ps)
		}
		if ps, err := exec.LookPath("pwsh"); err == nil {
			return c.powerShellCmd(ctx, ps)
		}
	}
	if ctx != nil {
		return exec.CommandContext(ctx, c.Name, c.Args...)
	}
	return exec.Command(c.Name, c.Args...)
}

func (c Command) powerShellCmd(ctx context.Context, ps string) *exec.Cmd {
	args := []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScoopScript(c.Args)}
	var cmd *exec.Cmd
	if ctx != nil {
		cmd = exec.CommandContext(ctx, ps, args...)
	} else {
		cmd = exec.Command(ps, args...)
	}
	if env := powershellEnv(ps); env != nil {
		cmd.Env = env
	}
	return cmd
}

func (c Command) String() string {
	if len(c.Args) == 0 {
		return c.Name
	}
	return c.Name + " " + strings.Join(c.Args, " ")
}

// NewBackend resolves the configured mode once and returns a backend whose
// read path is explicit. Auto only falls back when sfsu is not installed;
// a broken sfsu installation is reported instead of being silently ignored.
func NewBackend(mode Mode, runner *Runner) Backend {
	switch mode {
	case ModeSfsu:
		return NewSfsuBackend(runner)
	case ModeScoop:
		return NewScoopBackend(runner)
	default:
		if runner.HasSfsu() {
			return NewSfsuBackend(runner)
		}
		return NewScoopBackend(runner)
	}
}

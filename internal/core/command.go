package core

import "strings"

func appendFlag(args []string, cond bool, flag string) []string {
	if cond {
		return append(args, flag)
	}
	return args
}

func appendValue(args []string, value, flag string) []string {
	if value != "" {
		return append(args, flag, value)
	}
	return args
}

func (b *ScoopBackend) Install(apps []string, opts InstallOptions) Command {
	args := []string{"install"}
	args = append(args, apps...)
	args = appendFlag(args, opts.Global, "--global")
	args = appendValue(args, opts.Arch, "--arch")
	args = appendFlag(args, opts.NoCache, "--no-cache")
	args = appendFlag(args, opts.Independent, "--independent")
	args = appendFlag(args, opts.SkipHashCheck, "--skip-hash-check")
	args = appendFlag(args, opts.NoUpdateScoop, "--no-update-scoop")
	return Command{Name: "scoop", Args: args}
}

func (b *ScoopBackend) Uninstall(apps []string, opts UninstallOptions) Command {
	args := []string{"uninstall"}
	args = append(args, apps...)
	args = appendFlag(args, opts.Global, "--global")
	args = appendFlag(args, opts.Purge, "--purge")
	return Command{Name: "scoop", Args: args}
}

func (b *ScoopBackend) Update(apps []string, opts UpdateOptions) Command {
	args := []string{"update"}
	if opts.All {
		args = append(args, "*")
	} else {
		args = append(args, apps...)
	}
	args = appendFlag(args, opts.Global, "--global")
	args = appendFlag(args, opts.Force, "--force")
	args = appendFlag(args, opts.Independent, "--independent")
	args = appendFlag(args, opts.NoCache, "--no-cache")
	args = appendFlag(args, opts.SkipHashCheck, "--skip-hash-check")
	return Command{Name: "scoop", Args: args}
}

func (b *ScoopBackend) Hold(apps []string, opts HoldOptions) Command {
	args := []string{"hold"}
	args = append(args, apps...)
	args = appendFlag(args, opts.Global, "--global")
	return Command{Name: "scoop", Args: args}
}

func (b *ScoopBackend) Unhold(apps []string, opts HoldOptions) Command {
	args := []string{"unhold"}
	args = append(args, apps...)
	args = appendFlag(args, opts.Global, "--global")
	return Command{Name: "scoop", Args: args}
}

func (b *ScoopBackend) BucketAdd(name, source string) Command {
	args := []string{"bucket", "add", name}
	if strings.TrimSpace(source) != "" {
		args = append(args, source)
	}
	return Command{Name: "scoop", Args: args}
}

func (b *ScoopBackend) BucketRemove(name string) Command {
	return Command{Name: "scoop", Args: []string{"bucket", "rm", name}}
}

func (b *ScoopBackend) CacheRemove(apps []string, all bool) Command {
	args := []string{"cache", "rm"}
	if all {
		args = append(args, "*")
	} else {
		args = append(args, apps...)
	}
	return Command{Name: "scoop", Args: args}
}

func (b *ScoopBackend) Cleanup(apps []string, opts CleanupOptions) Command {
	args := []string{"cleanup"}
	if opts.All {
		args = append(args, "*")
	} else {
		args = append(args, apps...)
	}
	args = appendFlag(args, opts.Global, "--global")
	args = appendFlag(args, opts.Cache, "--cache")
	return Command{Name: "scoop", Args: args}
}

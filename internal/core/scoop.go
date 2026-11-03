package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/nyable/nyaru/internal/core/parse"
	"github.com/nyable/nyaru/internal/models"
)

// ScoopBackend is the native scoop backend. Reads go through PowerShell
// ConvertTo-Json; writes are interactive scoop commands.
type ScoopBackend struct {
	runner *Runner
}

func NewScoopBackend(runner *Runner) *ScoopBackend {
	return &ScoopBackend{runner: runner}
}

func (b *ScoopBackend) Name() string { return "scoop" }

func (b *ScoopBackend) Search(ctx context.Context, query string) ([]models.AppInfo, error) {
	raw, err := b.runner.ScoopSearchJSON(ctx, query)
	if err != nil {
		if isEmptyResult(err, raw) {
			return []models.AppInfo{}, nil
		}
		return nil, err
	}
	apps, err := parse.List[models.AppInfo](raw)
	if err != nil {
		return nil, err
	}
	if apps == nil {
		apps = []models.AppInfo{}
	}
	parse.SortApps(apps)
	return apps, nil
}

func (b *ScoopBackend) List(ctx context.Context, query string) ([]models.AppInfo, error) {
	raw, err := b.runner.ScoopListJSON(ctx, query)
	if err != nil {
		if isEmptyResult(err, raw) {
			return []models.AppInfo{}, nil
		}
		return nil, err
	}
	apps, err := parse.List[models.AppInfo](raw)
	if err != nil {
		return nil, err
	}
	if apps == nil {
		apps = []models.AppInfo{}
	}
	for i := range apps {
		apps[i].Installed = true
		if apps[i].Info == "" {
			apps[i].Info = apps[i].Notes
		}
		if apps[i].Status == models.StatusUnknown {
			apps[i].Status = parse.ClassifyStatus(apps[i].Info)
		}
	}
	parse.SortApps(apps)
	return apps, nil
}

func (b *ScoopBackend) Status(ctx context.Context) ([]models.AppInfo, error) {
	raw, err := b.runner.ScoopStatusJSON(ctx)
	if err != nil {
		if isEmptyResult(err, raw) {
			return []models.AppInfo{}, nil
		}
		return nil, err
	}
	return parse.ScoopStatus(raw)
}

func (b *ScoopBackend) Info(ctx context.Context, app string) (models.AppDetail, error) {
	raw, err := b.runner.ScoopOutput(ctx, "info", app)
	if err != nil {
		if strings.TrimSpace(raw) == "" {
			return models.AppDetail{}, err
		}
	}
	detail := parse.Info(raw)
	if detail.Name == "" && detail.Version == "" && detail.Description == "" {
		if err != nil {
			return models.AppDetail{}, err
		}
		return models.AppDetail{}, fmt.Errorf("未找到 %s 的信息", app)
	}
	return detail, nil
}

func (b *ScoopBackend) BucketList(ctx context.Context) ([]models.BucketResult, error) {
	raw, err := b.runner.ScoopBucketListJSON(ctx)
	if err != nil {
		if isEmptyResult(err, raw) {
			return []models.BucketResult{}, nil
		}
		return nil, err
	}
	buckets, err := parse.List[models.BucketResult](raw)
	if err != nil {
		return nil, err
	}
	if buckets == nil {
		buckets = []models.BucketResult{}
	}
	parse.SortBuckets(buckets)
	return buckets, nil
}

func (b *ScoopBackend) CacheList(ctx context.Context) ([]models.CacheResult, error) {
	raw, err := b.runner.ScoopOutput(ctx, "cache", "show")
	if err != nil {
		if isEmptyResult(err, raw) {
			return []models.CacheResult{}, nil
		}
		return nil, err
	}
	return parse.CacheOutput(raw), nil
}

// OldVersions scans the apps directories directly because scoop does not
// expose a dry-run for cleanup. The current junction marks the live version;
// only the remaining directories are reported.
func (b *ScoopBackend) OldVersions(ctx context.Context) ([]models.OldVersion, error) {
	raw, err := b.runner.ScoopOldVersionsJSON(ctx)
	if err != nil {
		if isEmptyResult(err, raw) {
			return []models.OldVersion{}, nil
		}
		return nil, err
	}
	return parse.OldVersions(raw)
}

// isEmptyResult distinguishes "nothing to show" from a real failure. Scoop
// exits non-zero for empty search/list results.
func isEmptyResult(err error, output string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(output + " " + err.Error())
	for _, marker := range []string{
		"no matches found",
		"there aren't any apps installed",
		"there are no apps installed",
		"no apps installed",
		"no bucket found",
		"no buckets found",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

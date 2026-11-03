package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/nyable/nyaru/internal/core/parse"
	"github.com/nyable/nyaru/internal/models"
)

// SfsuBackend uses sfsu for fast read paths and delegates writes to scoop.
type SfsuBackend struct {
	*ScoopBackend
	runner *Runner
}

func NewSfsuBackend(runner *Runner) *SfsuBackend {
	return &SfsuBackend{
		ScoopBackend: NewScoopBackend(runner),
		runner:       runner,
	}
}

func (b *SfsuBackend) Name() string { return "sfsu" }

func (b *SfsuBackend) Search(ctx context.Context, query string) ([]models.AppInfo, error) {
	raw, err := b.runner.SfsuJSON(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	return parse.SfsuSearch(raw)
}

func (b *SfsuBackend) List(ctx context.Context, query string) ([]models.AppInfo, error) {
	raw, err := b.runner.SfsuJSON(ctx, "list")
	if err != nil {
		return nil, err
	}
	apps, err := parse.List[models.AppInfo](raw)
	if err != nil {
		return nil, err
	}
	filtered := make([]models.AppInfo, 0, len(apps))
	q := strings.ToLower(strings.TrimSpace(query))
	for i := range apps {
		apps[i].Installed = true
		if apps[i].Info == "" {
			apps[i].Info = apps[i].Notes
		}
		if apps[i].Status == models.StatusUnknown {
			apps[i].Status = parse.ClassifyStatus(apps[i].Notes)
		}
		if q == "" ||
			strings.Contains(strings.ToLower(apps[i].Name), q) ||
			strings.Contains(strings.ToLower(apps[i].FullName()), q) {
			filtered = append(filtered, apps[i])
		}
	}
	parse.SortApps(filtered)
	return filtered, nil
}

func (b *SfsuBackend) Status(ctx context.Context) ([]models.AppInfo, error) {
	raw, err := b.runner.SfsuJSON(ctx, "status")
	if err != nil {
		return nil, err
	}
	return parse.SfsuStatus(raw)
}

func (b *SfsuBackend) Info(ctx context.Context, app string) (models.AppDetail, error) {
	if !b.runner.HasSfsu() {
		return models.AppDetail{}, fmt.Errorf("sfsu 未安装，无法使用 sfsu 后端")
	}
	raw, err := b.runner.SfsuJSON(ctx, "info", app)
	if err == nil {
		detail := parse.Info(raw)
		if detail.Name != "" || detail.Version != "" || detail.Description != "" {
			return detail, nil
		}
	}
	// sfsu info can be unavailable for some manifests; fall back to scoop.
	return b.ScoopBackend.Info(ctx, app)
}

func (b *SfsuBackend) BucketList(ctx context.Context) ([]models.BucketResult, error) {
	raw, err := b.runner.SfsuJSON(ctx, "bucket", "list")
	if err != nil {
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

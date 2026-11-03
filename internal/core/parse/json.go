package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/nyable/nyaru/internal/models"
)

var ansiPattern = regexp.MustCompile(`\x1b(?:\[[0-9;?]*[ -/]*[@-~]|\][^\x07]*(?:\x07|\x1b\\))`)

// StripANSI removes common ANSI color sequences from command output.
func StripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func firstJSONIndex(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '{' || s[i] == '[' {
			return i
		}
	}
	return -1
}

// List decodes the first JSON value in raw into a slice. Scoop commands often
// print human-readable messages before their JSON payload, and a single object
// is converted into a one-element list.
func List[T any](raw string) ([]T, error) {
	raw = StripANSI(raw)
	raw = strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff"))
	var lastErr error
	for i := 0; i < len(raw); i++ {
		if raw[i] != '{' && raw[i] != '[' {
			continue
		}
		dec := json.NewDecoder(bytes.NewReader([]byte(raw[i:])))
		if raw[i] == '[' {
			var list []T
			if err := dec.Decode(&list); err == nil {
				return list, nil
			} else {
				lastErr = err
			}
		} else {
			var item T
			if err := dec.Decode(&item); err == nil {
				return []T{item}, nil
			} else {
				lastErr = err
			}
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, nil
}

func SortApps(apps []models.AppInfo) {
	sort.SliceStable(apps, func(i, j int) bool {
		if strings.EqualFold(apps[i].Name, apps[j].Name) {
			return apps[i].FullName() < apps[j].FullName()
		}
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})
}

// SfsuSearch flattens sfsu's bucket -> apps map into a single list.
func SfsuSearch(raw string) ([]models.AppInfo, error) {
	raw = StripANSI(raw)
	idx := firstJSONIndex(raw)
	if idx < 0 {
		return nil, nil
	}
	var result map[string][]models.AppInfo
	if err := json.Unmarshal([]byte(raw[idx:]), &result); err != nil {
		return nil, fmt.Errorf("解析 sfsu search 输出失败: %w", err)
	}
	apps := make([]models.AppInfo, 0)
	for bucket, list := range result {
		for i := range list {
			list[i].Bucket = bucket
		}
		apps = append(apps, list...)
	}
	SortApps(apps)
	return apps, nil
}

// SfsuStatus maps sfsu's status JSON into the unified model.
func SfsuStatus(raw string) ([]models.AppInfo, error) {
	raw = StripANSI(raw)
	idx := firstJSONIndex(raw)
	if idx < 0 {
		return nil, nil
	}
	var result struct {
		Packages []struct {
			Name                string `json:"name"`
			Current             string `json:"current"`
			Available           string `json:"available"`
			MissingDependencies []any  `json:"missing_dependencies"`
			Info                string `json:"info"`
		} `json:"packages"`
	}
	if err := json.Unmarshal([]byte(raw[idx:]), &result); err != nil {
		return nil, fmt.Errorf("解析 sfsu status 输出失败: %w", err)
	}
	apps := make([]models.AppInfo, 0, len(result.Packages))
	for _, p := range result.Packages {
		status := ClassifyStatus(p.Info)
		if status == models.StatusUnknown && len(p.MissingDependencies) > 0 {
			status = models.StatusMissingDeps
		}
		if status == models.StatusUnknown && p.Available != "" && p.Available != p.Current {
			status = models.StatusOutdated
		}
		apps = append(apps, models.AppInfo{
			Name:             p.Name,
			CurrentVersion:   p.Current,
			AvailableVersion: p.Available,
			Version:          p.Available,
			Installed:        true,
			Status:           status,
			Info:             p.Info,
		})
	}
	SortApps(apps)
	return apps, nil
}

// ScoopStatus maps scoop status's JSON fields, which contain spaces.
func ScoopStatus(raw string) ([]models.AppInfo, error) {
	type scoopStatus struct {
		Name                string `json:"Name"`
		InstalledVersion    string `json:"Installed Version"`
		LatestVersion       string `json:"Latest Version"`
		MissingDependencies string `json:"Missing Dependencies"`
		Info                string `json:"Info"`
		Global              bool   `json:"Global"`
	}
	rows, err := List[scoopStatus](raw)
	if err != nil {
		return nil, fmt.Errorf("解析 scoop status 输出失败: %w", err)
	}
	apps := make([]models.AppInfo, 0, len(rows))
	for _, row := range rows {
		status := ClassifyStatus(row.Info)
		if status == models.StatusUnknown && strings.TrimSpace(row.MissingDependencies) != "" {
			status = models.StatusMissingDeps
		}
		if status == models.StatusUnknown && row.LatestVersion != "" && row.LatestVersion != row.InstalledVersion {
			status = models.StatusOutdated
		}
		apps = append(apps, models.AppInfo{
			Name:             row.Name,
			CurrentVersion:   row.InstalledVersion,
			AvailableVersion: row.LatestVersion,
			Version:          row.LatestVersion,
			Installed:        true,
			Global:           row.Global,
			Status:           status,
			Info:             strings.TrimSpace(row.Info),
		})
	}
	SortApps(apps)
	return apps, nil
}

func ClassifyStatus(info string) models.AppStatus {
	lower := strings.ToLower(info)
	switch {
	case strings.Contains(lower, "install failed"):
		return models.StatusFailed
	case strings.Contains(lower, "held"):
		return models.StatusHeld
	case strings.Contains(lower, "deprecated"):
		return models.StatusDeprecated
	case strings.Contains(lower, "removed"):
		return models.StatusRemoved
	case strings.Contains(lower, "missing depend"):
		return models.StatusMissingDeps
	default:
		return models.StatusUnknown
	}
}

func SortBuckets(buckets []models.BucketResult) {
	sort.SliceStable(buckets, func(i, j int) bool {
		return strings.ToLower(buckets[i].Name) < strings.ToLower(buckets[j].Name)
	})
}

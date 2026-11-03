package parse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/utils"
)

// OldVersions decodes the filesystem scan performed by the scoop backend and
// fills in human-readable sizes. Entries without any obsolete version are
// dropped so the cleanup screen only shows actionable applications.
func OldVersions(raw string) ([]models.OldVersion, error) {
	entries, err := List[models.OldVersion](raw)
	if err != nil {
		return nil, fmt.Errorf("解析旧版本列表失败: %w", err)
	}
	result := make([]models.OldVersion, 0, len(entries))
	for _, entry := range entries {
		entry.Name = strings.TrimSpace(entry.Name)
		if entry.Name == "" {
			continue
		}
		var total int64
		versions := make([]models.OldVersionEntry, 0, len(entry.Versions))
		for _, version := range entry.Versions {
			version.Version = strings.TrimSpace(version.Version)
			if version.Version == "" || version.Version == "current" {
				continue
			}
			if version.Length < 0 {
				version.Length = 0
			}
			version.FormatSize = utils.SizeToHumanRead(version.Length)
			total += version.Length
			versions = append(versions, version)
		}
		if len(versions) == 0 {
			continue
		}
		sort.SliceStable(versions, func(i, j int) bool {
			return versions[i].Version > versions[j].Version
		})
		entry.Versions = versions
		entry.Length = total
		entry.FormatSize = utils.SizeToHumanRead(total)
		result = append(result, entry)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if strings.EqualFold(result[i].Name, result[j].Name) {
			return !result[i].Global && result[j].Global
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result, nil
}

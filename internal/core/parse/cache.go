package parse

import (
	"sort"
	"strings"

	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/utils"
)

// CacheOutput parses `scoop cache show` and groups rows by application,
// because `scoop cache rm <app>` removes every cached version of that app.
func CacheOutput(output string) []models.CacheResult {
	clean := StripANSI(strings.ReplaceAll(output, "\r\n", "\n"))
	lines := strings.Split(clean, "\n")

	type row struct {
		app     string
		version string
		length  int64
	}
	rows := make([]row, 0)

	nameIdx, versionIdx, sizeIdx := -1, -1, -1
	started := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "total:") {
			break
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if !started {
			if strings.Contains(lower, "name") && (strings.Contains(lower, "length") || strings.Contains(lower, "size")) {
				for i, f := range fields {
					switch strings.ToLower(f) {
					case "name":
						nameIdx = i
					case "version":
						versionIdx = i
					case "length", "size":
						sizeIdx = i
					}
				}
				if nameIdx >= 0 && versionIdx >= 0 && sizeIdx >= 0 {
					started = true
				}
				continue
			}
			if strings.Contains(line, "----") {
				started = true
				continue
			}
			continue
		}

		if strings.HasPrefix(fields[0], "---") || strings.EqualFold(fields[0], "name") {
			continue
		}
		if len(fields) <= maxInt(nameIdx, versionIdx) {
			continue
		}
		app := strings.Trim(fields[nameIdx], "\"', ")
		version := strings.Trim(fields[versionIdx], "\"', ")
		if app == "" || app == "Total:" {
			continue
		}
		length := parseRowSize(fields, sizeIdx)
		if length <= 0 {
			continue
		}
		rows = append(rows, row{app: app, version: version, length: length})
	}

	// Group by app and preserve version order.
	byApp := make(map[string]*models.CacheResult)
	for _, r := range rows {
		entry, ok := byApp[r.app]
		if !ok {
			entry = &models.CacheResult{App: r.app}
			byApp[r.app] = entry
		}
		entry.Length += r.length
		found := false
		for _, v := range entry.Versions {
			if v == r.version {
				found = true
				break
			}
		}
		if !found && r.version != "" {
			entry.Versions = append(entry.Versions, r.version)
		}
	}

	result := make([]models.CacheResult, 0, len(byApp))
	for _, entry := range byApp {
		sort.SliceStable(entry.Versions, func(i, j int) bool {
			return entry.Versions[i] > entry.Versions[j]
		})
		entry.FormatSize = utils.SizeToHumanRead(entry.Length)
		result = append(result, *entry)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.ToLower(result[i].App) < strings.ToLower(result[j].App)
	})
	for i := range result {
		result[i].Index = i
	}
	return result
}

func parseRowSize(fields []string, sizeIdx int) int64 {
	// The size column can be "1.90 MiB" (two fields).
	if sizeIdx >= 0 && sizeIdx < len(fields) {
		if sizeIdx+1 < len(fields) {
			if n := utils.HumanSizeToBytes(fields[sizeIdx] + " " + fields[sizeIdx+1]); n > 0 {
				return n
			}
		}
		if n := utils.HumanSizeToBytes(fields[sizeIdx]); n > 0 {
			return n
		}
	}
	// Fallback: scan from the end for a number + unit pair.
	for i := len(fields) - 2; i >= 0; i-- {
		if n := utils.HumanSizeToBytes(fields[i] + " " + fields[i+1]); n > 0 {
			return n
		}
	}
	return 0
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package parse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nyable/nyaru/internal/models"
)

// Info parses either sfsu's JSON info or scoop's PowerShell list-formatted
// text into the structured AppDetail model.
func Info(raw string) models.AppDetail {
	raw = strings.TrimSpace(strings.TrimPrefix(StripANSI(raw), "\ufeff"))
	if raw == "" {
		return models.AppDetail{}
	}
	if idx := firstJSONIndex(raw); idx >= 0 && raw[idx] == '{' {
		if detail, err := infoJSON(raw[idx:]); err == nil &&
			(detail.Name != "" || detail.Version != "" || detail.Description != "") {
			return detail
		}
	}
	return infoText(raw)
}

func infoJSON(raw string) (models.AppDetail, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return models.AppDetail{}, err
	}
	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok && v != nil {
				return fmt.Sprintf("%v", v)
			}
		}
		return ""
	}
	return models.AppDetail{
		Name:        get("name", "Name"),
		Description: get("description", "Description"),
		Version:     get("version", "Version"),
		Bucket:      get("bucket", "Bucket", "source", "Source"),
		Website:     get("website", "Website", "homepage", "Homepage"),
		License:     licenseValue(m["license"], m["License"]),
		Binaries:    binariesFromValue(m["bin"], m["binaries"], m["Binaries"]),
		Notes:       get("notes", "Notes"),
	}, nil
}

func licenseValue(values ...any) string {
	for _, value := range values {
		switch v := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(v) != "" {
				return v
			}
		case map[string]any:
			for _, key := range []string{"String", "License", "Object", "identifier"} {
				if nested, ok := v[key]; ok {
					if got := licenseValue(nested); got != "" {
						return got
					}
				}
			}
			if nested, ok := v["url"]; ok {
				if got := licenseValue(nested); got != "" {
					return got
				}
			}
		}
	}
	return ""
}

func binariesFromValue(values ...any) []string {
	for _, v := range values {
		switch val := v.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(val) != "" {
				return splitBinaries(val)
			}
		case []any:
			out := collectBinaries(val)
			if len(out) > 0 {
				return out
			}
		case []string:
			return val
		}
	}
	return nil
}

// collectBinaries flattens sfsu's bin arrays, which can contain nested
// [command, alias, args] entries as well as plain strings.
func collectBinaries(values []any) []string {
	out := make([]string, 0, len(values))
	for _, item := range values {
		switch v := item.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				out = append(out, v)
			}
		case []any:
			if len(v) > 0 {
				if s, ok := v[0].(string); ok {
					if strings.TrimSpace(s) != "" {
						out = append(out, s)
					}
					continue
				}
			}
			out = append(out, collectBinaries(v)...)
		}
	}
	return out
}

func splitBinaries(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == '|' || r == ','
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if v := strings.TrimSpace(f); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func infoText(raw string) models.AppDetail {
	data := make(map[string]string)
	var currentKey string
	var currentVal strings.Builder

	flush := func() {
		if currentKey != "" {
			data[currentKey] = strings.TrimSpace(currentVal.String())
		}
	}

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if idx := strings.Index(line, " : "); idx >= 0 {
			flush()
			currentKey = strings.TrimSpace(line[:idx])
			currentVal.Reset()
			currentVal.WriteString(strings.TrimSpace(line[idx+3:]))
			continue
		}
		if currentKey != "" && strings.HasPrefix(line, " ") {
			currentVal.WriteString("\n")
			currentVal.WriteString(strings.TrimSpace(line))
		}
	}
	flush()

	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := data[k]; ok {
				return v
			}
		}
		return ""
	}
	return models.AppDetail{
		Name:        get("Name"),
		Description: get("Description"),
		Version:     get("Version"),
		Bucket:      get("Source", "Bucket"),
		Website:     get("Website"),
		License:     get("License"),
		Binaries:    splitBinaries(get("Binaries")),
		Notes:       get("Notes"),
	}
}

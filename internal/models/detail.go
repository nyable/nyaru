package models

import "fmt"

// AppDetail is the structured result of `scoop info` or `sfsu info --json`.
type AppDetail struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Bucket      string   `json:"bucket,omitempty"`
	Website     string   `json:"website,omitempty"`
	License     string   `json:"license,omitempty"`
	Binaries    []string `json:"binaries,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

// Fields returns the detail lines that have a value, in display order.
func (d AppDetail) Fields() []DetailField {
	fields := []DetailField{
		{"名称", d.Name},
		{"版本", d.Version},
		{"来源", d.Bucket},
		{"主页", d.Website},
		{"许可证", d.License},
		{"可执行文件", joinBinaries(d.Binaries)},
	}
	out := make([]DetailField, 0, len(fields))
	for _, f := range fields {
		if f.Value != "" {
			out = append(out, f)
		}
	}
	return out
}

// DetailField is one label/value pair in the detail screen.
type DetailField struct {
	Label string
	Value string
}

func joinBinaries(bins []string) string {
	if len(bins) == 0 {
		return ""
	}
	out := ""
	for i, b := range bins {
		if i > 0 {
			out += " · "
		}
		out += b
	}
	return out
}

func (d AppDetail) String() string {
	out := ""
	for _, f := range d.Fields() {
		out += fmt.Sprintf("%s: %s\n", f.Label, f.Value)
	}
	if d.Description != "" {
		out += "\n" + d.Description + "\n"
	}
	if d.Notes != "" {
		out += "\n" + d.Notes + "\n"
	}
	return out
}

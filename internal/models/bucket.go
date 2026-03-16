package models

import (
	"encoding/json"
	"fmt"
)

type BucketUpdatedInfo struct {
	Value       string `json:"value"`
	DisplayHint int    `json:"DisplayHint"`
	DateTime    string `json:"DateTime"`
}

// UnmarshalJSON implements custom unmarshalling to handle both string and object formats
func (b *BucketUpdatedInfo) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first (sfsu format)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		b.DateTime = s
		b.Value = s
		return nil
	}

	// Try to unmarshal as object (Scoop format)
	type Alias BucketUpdatedInfo
	var aux Alias
	if err := json.Unmarshal(data, &aux); err == nil {
		*b = BucketUpdatedInfo(aux)
		return nil
	}

	return fmt.Errorf("failed to unmarshal BucketUpdatedInfo: %s", string(data))
}

type BucketResult struct {
	Index     int               `json:"-"`
	Name      string            `json:"name"`
	Source    string            `json:"source"`
	Updated   BucketUpdatedInfo `json:"updated"`
	Manifests int               `json:"manifests"`
	Selected  bool              `json:"-"` // Internal field for multiselect
}

// FilterValue implements the bubbletea list.Item interface
func (b BucketResult) FilterValue() string {
	return b.Name
}

// Title implements the bubbletea list.DefaultItem interface
func (b BucketResult) Title() string {
	return b.Name
}

// Description implements the bubbletea list.DefaultItem interface
func (b BucketResult) Description() string {
	selected := ""
	if b.Selected {
		selected = " [√]"
	}
	return fmt.Sprintf("%s (%d manifests) - %s%s", b.Source, b.Manifests, b.Updated.DateTime, selected)
}

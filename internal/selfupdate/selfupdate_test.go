package selfupdate

import (
	"strings"
	"testing"
)

func TestHelperScriptQuotesLogPath(t *testing.T) {
	script := helperScript(123, ApplyOptions{Global: true}, `C:\Users\O'Brien\AppData\Roaming\nyaru\self-update.log`)
	if !strings.Contains(script, `$log = 'C:\Users\O''Brien\AppData\Roaming\nyaru\self-update.log'`) {
		t.Fatalf("log path not quoted:\n%s", script)
	}
	if !strings.Contains(script, "scoop update nyaru --global") {
		t.Fatalf("global update missing:\n%s", script)
	}
}

func TestVersionLess(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"1.0.0", "1.0.1", true},
		{"1.0.1", "1.0.0", false},
		{"v0.9", "1.0", true},
		{"1.0", "1.0", false},
		{"1.2.10", "1.2.9", false},
	}
	for _, tt := range tests {
		if got := versionLess(tt.current, tt.latest); got != tt.want {
			t.Fatalf("versionLess(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

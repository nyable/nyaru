package parse

import (
	"strings"
	"testing"
)

func TestOldVersionsParsesSizesAndOrder(t *testing.T) {
	raw := `[
		{"name":"git","current":"2.47.0","global":false,"length":3145728,"versions":[
			{"version":"2.40.0","length":1048576},
			{"version":"2.45.0","length":2097152}
		]},
		{"name":"7zip","current":"24.09","global":true,"length":1048576,"versions":[
			{"version":"24.08","length":1048576}
		]}
	]`
	entries, err := OldVersions(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "7zip" || !entries[0].Global {
		t.Fatalf("expected sorted entries with global scope, got %+v", entries[0])
	}
	git := entries[1]
	if git.Length != 3145728 {
		t.Fatalf("total length = %d", git.Length)
	}
	if git.FormatSize != "3.0 MB" {
		t.Fatalf("format size = %q", git.FormatSize)
	}
	if len(git.Versions) != 2 || git.Versions[0].Version != "2.45.0" {
		t.Fatalf("unexpected versions: %+v", git.Versions)
	}
	if git.Versions[0].FormatSize == "" {
		t.Fatal("expected per-version size")
	}
	if !strings.Contains(git.Description(), "可释放 3.0 MB") {
		t.Fatalf("description missing size: %s", git.Description())
	}
}

func TestOldVersionsEmptyAndInvalid(t *testing.T) {
	for _, raw := range []string{"", "[]", "null"} {
		entries, err := OldVersions(raw)
		if err != nil {
			t.Fatalf("%q: %v", raw, err)
		}
		if len(entries) != 0 {
			t.Fatalf("%q: expected no entries, got %d", raw, len(entries))
		}
	}
	if _, err := OldVersions("{not-json"); err == nil {
		t.Fatal("expected parse error")
	}
}

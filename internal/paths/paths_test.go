package paths

import (
	"path/filepath"
	"testing"
)

func TestPlatformPaths(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(root, "appdata"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg"))

	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	wantDir := filepath.Join(root, "xdg", "nyaru")
	if dir != wantDir {
		t.Fatalf("config dir = %q, want %q", dir, wantDir)
	}
	file, err := ConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(file) != "config.yaml" {
		t.Fatalf("config file = %q", file)
	}
	log, err := SelfUpdateLog()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(log) != "self-update.log" {
		t.Fatalf("log file = %q", log)
	}
	legacy, err := LegacyConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	wantLegacy := filepath.Join(home, ".nyaru", "config.yaml")
	if legacy != wantLegacy {
		t.Fatalf("legacy config = %q, want %q", legacy, wantLegacy)
	}
}

package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMode(t *testing.T) {
	found := func(string) (string, error) { return "sfsu", nil }
	missing := func(string) (string, error) { return "", errors.New("not found") }

	tests := []struct {
		mode string
		look func(string) (string, error)
		want string
	}{
		{"sfsu", missing, "sfsu"},
		{"scoop", found, "scoop"},
		{"auto", found, "sfsu"},
		{"auto", missing, "scoop"},
		{"", missing, "scoop"},
		{"invalid", found, "sfsu"},
	}
	for _, tt := range tests {
		if got := ResolveMode(tt.mode, tt.look); got != tt.want {
			t.Fatalf("ResolveMode(%q) = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	cfg := &Config{Mode: "  SFS S ", Language: ""}
	cfg.normalize()
	if cfg.Mode != "auto" {
		t.Fatalf("invalid mode should normalize to auto, got %q", cfg.Mode)
	}
	if cfg.Language != "zh" {
		t.Fatalf("empty language should default to zh, got %q", cfg.Language)
	}
	if cfg.StatusTimeoutSeconds() != DefaultStatusTimeout {
		t.Fatalf("status timeout = %d", cfg.StatusTimeoutSeconds())
	}
}

func TestLoadMigratesLegacyConfig(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(root, "appdata"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg"))
	current = nil
	t.Cleanup(func() { current = nil })

	legacyDir := filepath.Join(home, ".nyaru")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := "mode: scoop\nconfirm_batch: false\nstatus_timeout: 9\n"
	if err := os.WriteFile(filepath.Join(legacyDir, "config.yaml"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != "scoop" || cfg.ConfirmBatch || cfg.StatusTimeoutSeconds() != 9 {
		t.Fatalf("unexpected migrated config: %+v", cfg)
	}
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if path == filepath.Join(legacyDir, "config.yaml") {
		t.Fatalf("config path was not migrated: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("migrated config missing: %v", err)
	}
}

func TestSettingsGetSet(t *testing.T) {
	cfg := DefaultConfig()
	if value, err := cfg.SettingValue("mode"); err != nil || value != "auto" {
		t.Fatalf("mode = %q, %v", value, err)
	}
	if err := cfg.SetSetting("mode", "SCOOP"); err != nil || cfg.Mode != "scoop" {
		t.Fatalf("set mode: %q, %v", cfg.Mode, err)
	}
	if err := cfg.SetSetting("confirm_batch", "off"); err != nil || cfg.ConfirmBatch {
		t.Fatalf("set confirm_batch: %v, %v", cfg.ConfirmBatch, err)
	}
	if err := cfg.SetSetting("status_timeout", "never"); err != nil || cfg.StatusTimeoutSeconds() != 0 {
		t.Fatalf("set status_timeout never: %d, %v", cfg.StatusTimeoutSeconds(), err)
	}
	if err := cfg.SetSetting("status_timeout", "12"); err != nil || cfg.StatusTimeoutSeconds() != 12 {
		t.Fatalf("set status_timeout: %d, %v", cfg.StatusTimeoutSeconds(), err)
	}
	for _, tc := range []struct{ key, value string }{
		{"mode", "bad"},
		{"confirm_batch", "maybe"},
		{"status_timeout", "-1"},
		{"status_timeout", "99999"},
		{"unknown", "x"},
	} {
		if err := cfg.SetSetting(tc.key, tc.value); err == nil {
			t.Fatalf("expected error for %s=%s", tc.key, tc.value)
		}
	}
}

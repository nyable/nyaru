package models

import "testing"

func TestAppInfoFullName(t *testing.T) {
	app := AppInfo{Name: "git", Bucket: "main"}
	if got := app.FullName(); got != "main/git" {
		t.Fatalf("FullName = %q", got)
	}
	app = AppInfo{Name: "git", Source: "extras"}
	if got := app.FullName(); got != "extras/git" {
		t.Fatalf("FullName = %q", got)
	}
	app = AppInfo{Name: "git"}
	if got := app.FullName(); got != "git" {
		t.Fatalf("FullName = %q", got)
	}
}

func TestAppInfoIDSeparatesGlobalInstall(t *testing.T) {
	local := AppInfo{Name: "git", Source: "main", Installed: true}
	global := AppInfo{Name: "git", Source: "main", Installed: true, Global: true}
	if local.ID() == global.ID() {
		t.Fatalf("local and global installs share ID %q", local.ID())
	}
}

func TestAppInfoOrigin(t *testing.T) {
	if got := (AppInfo{Name: "git", Bucket: "main", Source: "extras"}).Origin(); got != "main" {
		t.Fatalf("Origin = %q", got)
	}
	if got := (AppInfo{Name: "git", Source: "extras"}).Origin(); got != "extras" {
		t.Fatalf("Origin = %q", got)
	}
}

func TestAppInfoDisplayVersion(t *testing.T) {
	app := AppInfo{CurrentVersion: "1.0", AvailableVersion: "2.0"}
	if got := app.DisplayVersion(); got != "1.0 → 2.0" {
		t.Fatalf("DisplayVersion = %q", got)
	}
	app = AppInfo{AvailableVersion: "2.0"}
	if got := app.DisplayVersion(); got != "2.0" {
		t.Fatalf("DisplayVersion = %q", got)
	}
}

func TestAppInfoHasUpdate(t *testing.T) {
	if !(AppInfo{Status: StatusOutdated}).HasUpdate() {
		t.Fatal("outdated app should have update")
	}
	if (AppInfo{CurrentVersion: "1.0", AvailableVersion: "1.0"}).HasUpdate() {
		t.Fatal("same version should not have update")
	}
}

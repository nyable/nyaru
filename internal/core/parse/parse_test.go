package parse

import (
	"strings"
	"testing"

	"github.com/nyable/nyaru/internal/models"
)

func TestScoopStatusMapsVersionsAndState(t *testing.T) {
	raw := `Scoop is up to date.
[{"Name":"git","Installed Version":"2.40.0","Latest Version":"2.47.0","Missing Dependencies":"","Info":"","Global":false},
 {"Name":"held-app","Installed Version":"1.0","Latest Version":"1.1","Missing Dependencies":"","Info":"Held package"},
 {"Name":"broken","Installed Version":"1.0","Latest Version":"","Missing Dependencies":"dep1 | dep2","Info":"Install failed"},
 {"Name":"deps","Installed Version":"1.0","Latest Version":"1.0","Missing Dependencies":"dep1 | dep2","Info":"","Global":true}]`
	apps, err := ScoopStatus(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 4 {
		t.Fatalf("expected 4 apps, got %d", len(apps))
	}
	if apps[0].Name != "broken" {
		t.Fatalf("expected sorted by name, got %q", apps[0].Name)
	}
	byName := map[string]models.AppInfo{}
	for _, app := range apps {
		byName[app.Name] = app
	}
	if got := byName["git"]; got.CurrentVersion != "2.40.0" || got.AvailableVersion != "2.47.0" || got.Status != models.StatusOutdated {
		t.Fatalf("unexpected git mapping: %+v", got)
	}
	if got := byName["deps"]; !got.Global {
		t.Fatalf("expected global app, got %+v", got)
	}
	if got := byName["held-app"]; got.Status != models.StatusHeld {
		t.Fatalf("expected held status, got %+v", got)
	}
	if got := byName["broken"]; got.Status != models.StatusFailed {
		t.Fatalf("expected failed status, got %+v", got)
	}
	if got := byName["deps"]; got.Status != models.StatusMissingDeps {
		t.Fatalf("expected missing deps status, got %+v", got)
	}
}

func TestSfsuSearchFlattensBuckets(t *testing.T) {
	raw := `{
  "main": [{"name":"git","bucket":"main","version":"2.47.0","installed":true}],
  "extras": [{"name":"aria2","bucket":"extras","version":"1.37.0","installed":false}]
}`
	apps, err := SfsuSearch(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
	byName := map[string]models.AppInfo{}
	for _, app := range apps {
		byName[app.Name] = app
	}
	if byName["git"].Bucket != "main" || byName["aria2"].Bucket != "extras" {
		t.Fatalf("bucket not applied: %+v", apps)
	}
}

func TestSfsuStatusMapsPackageState(t *testing.T) {
	raw := `{"scoop":false,"buckets":[],"packages":[
		{"name":"git","current":"2.40.0","available":"2.47.0","missing_dependencies":[],"info":null},
		{"name":"held","current":"1.0","available":"1.0","missing_dependencies":[],"info":"Held package"},
		{"name":"deps","current":"1.0","available":"1.0","missing_dependencies":[{"name":"x"}],"info":null}
	]}`
	apps, err := SfsuStatus(raw)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]models.AppInfo{}
	for _, app := range apps {
		byName[app.Name] = app
	}
	if byName["git"].Status != models.StatusOutdated {
		t.Fatalf("expected outdated, got %+v", byName["git"])
	}
	if byName["held"].Status != models.StatusHeld {
		t.Fatalf("expected held, got %+v", byName["held"])
	}
	if byName["deps"].Status != models.StatusMissingDeps {
		t.Fatalf("expected missing deps, got %+v", byName["deps"])
	}
}

func TestListSkipsHostMessages(t *testing.T) {
	raw := "Results from local buckets...\n[{\"Name\":\"git\",\"Version\":\"2.47.0\",\"Source\":\"main\"}]"
	apps, err := List[models.AppInfo](raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 || apps[0].Name != "git" {
		t.Fatalf("unexpected list: %+v", apps)
	}
}

func TestCacheOutputGroupsByAppAndSkipsTotal(t *testing.T) {
	raw := `Name   Version Length
----   ------- ------
7zip   24.09   1.5 MB
7zip   24.08   1.0 MB
git    2.47.0  50.0 MB

Total: 3 files, 52.5 MB
`
	cache := CacheOutput(raw)
	if len(cache) != 2 {
		t.Fatalf("expected 2 apps, got %d: %+v", len(cache), cache)
	}
	byApp := map[string]models.CacheResult{}
	for _, c := range cache {
		byApp[c.App] = c
	}
	if _, ok := byApp["Total:"]; ok {
		t.Fatalf("Total summary row was parsed as an app: %+v", cache)
	}
	seven := byApp["7zip"]
	if len(seven.Versions) != 2 || seven.Versions[0] != "24.09" {
		t.Fatalf("unexpected versions: %+v", seven)
	}
	if seven.Length <= 0 || seven.FormatSize == "" {
		t.Fatalf("expected size, got %+v", seven)
	}
	if !strings.Contains(cache[0].App, "7zip") && !strings.Contains(cache[1].App, "7zip") {
		t.Fatalf("expected 7zip in sorted results: %+v", cache)
	}
}

func TestInfoJSON(t *testing.T) {
	raw := `{"name":"git","description":"distributed version control","version":"2.47.0","bucket":"main","website":"https://git-scm.com","license":"GPL-2.0-only","bin":["git.exe","gitk.exe"],"notes":"note"}`
	detail := Info(raw)
	if detail.Name != "git" || detail.Bucket != "main" || detail.Website != "https://git-scm.com" {
		t.Fatalf("unexpected detail: %+v", detail)
	}
	if len(detail.Binaries) != 2 || detail.Binaries[0] != "git.exe" {
		t.Fatalf("unexpected binaries: %+v", detail.Binaries)
	}
}

func TestInfoJSONNestedBin(t *testing.T) {
	raw := `{"name":"git","version":"2.47.0","license":{"String":"GPL-2.0-only"},"bin":[["git.exe","git"],["gitk.exe"]]}`
	detail := Info(raw)
	if len(detail.Binaries) != 2 || detail.Binaries[0] != "git.exe" || detail.Binaries[1] != "gitk.exe" {
		t.Fatalf("unexpected binaries: %+v", detail.Binaries)
	}
	if detail.License != "GPL-2.0-only" {
		t.Fatalf("unexpected license: %q", detail.License)
	}
}

func TestInfoText(t *testing.T) {
	raw := `Name        : git
Description : distributed version control
Version     : 2.47.0
Source      : main
Website     : https://git-scm.com
License     : GPL-2.0-only
Binaries    : git.exe | gitk.exe
Notes       : first line
              second line`
	detail := Info(raw)
	if detail.Name != "git" || detail.Bucket != "main" || detail.Version != "2.47.0" {
		t.Fatalf("unexpected detail: %+v", detail)
	}
	if len(detail.Binaries) != 2 {
		t.Fatalf("unexpected binaries: %+v", detail.Binaries)
	}
	if !strings.Contains(detail.Notes, "second line") {
		t.Fatalf("notes not parsed: %q", detail.Notes)
	}
}

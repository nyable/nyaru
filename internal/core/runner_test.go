package core

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestEnvWithReplacesExistingKey(t *testing.T) {
	got := envWith([]string{"A=1", "NYARU_QUERY=old", "B=2"}, "NYARU_QUERY=new")
	want := []string{"A=1", "B=2", "NYARU_QUERY=new"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("envWith = %#v, want %#v", got, want)
	}
}

func TestEnvWithoutRemovesKey(t *testing.T) {
	got := envWithout([]string{"A=1", "PSModulePath=C:\\Program Files\\PowerShell\\7\\Modules", "B=2"}, "PSModulePath")
	want := []string{"A=1", "B=2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("envWithout = %#v, want %#v", got, want)
	}
}

func TestIsWindowsPowerShell(t *testing.T) {
	for _, name := range []string{"powershell", "powershell.exe", "PowerShell.EXE"} {
		if !isWindowsPowerShell(name) {
			t.Fatalf("%q should be Windows PowerShell", name)
		}
	}
	for _, name := range []string{"pwsh", "pwsh.exe"} {
		if isWindowsPowerShell(name) {
			t.Fatalf("%q should not be Windows PowerShell", name)
		}
	}
}

func TestCommandString(t *testing.T) {
	cmd := Command{Name: "scoop", Args: []string{"install", "git"}}
	if got := cmd.String(); got != "scoop install git" {
		t.Fatalf("String = %q", got)
	}
}

func TestPSScoopScriptQuotesArguments(t *testing.T) {
	script := psScoopScript([]string{"bucket", "add", "x", "https://example.com/a?b&c"})
	want := "& scoop 'bucket' 'add' 'x' 'https://example.com/a?b&c'"
	if !strings.Contains(script, want) {
		t.Fatalf("script missing quoted args: %s", script)
	}
}

func TestPowerShellCmdUsesExecutionPolicy(t *testing.T) {
	cmd := (Command{Name: "scoop", Args: []string{"install", "git"}}).powerShellCmd(context.Background(), "powershell")
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "-ExecutionPolicy Bypass") {
		t.Fatalf("missing execution policy: %v", cmd.Args)
	}
	if !strings.Contains(joined, "& scoop 'install' 'git'") {
		t.Fatalf("missing scoop command: %v", cmd.Args)
	}
}

func TestIsEmptyResult(t *testing.T) {
	if !isEmptyResult(errors.New("No bucket found"), "") {
		t.Fatal("expected empty bucket result")
	}
	if isEmptyResult(errors.New("access denied"), "") {
		t.Fatal("unexpected empty result")
	}
}

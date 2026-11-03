package cli

import "testing"

func TestUpdateCommandSyntax(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"update"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Use != "update [app...]" {
		t.Fatalf("update Use = %q", cmd.Use)
	}
	if cmd.Flags().Lookup("all") == nil {
		t.Fatal("update is missing --all")
	}
}

func TestListCommandSyntax(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Use != "list [query]" {
		t.Fatalf("list Use = %q", cmd.Use)
	}
}

func TestInstallHasNoGetAlias(t *testing.T) {
	for _, alias := range installCmd.Aliases {
		if alias == "get" {
			t.Fatal("install should not advertise an unregistered get alias")
		}
	}
}

func TestSettingsCommandSyntax(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"settings"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Use != "settings" {
		t.Fatalf("settings Use = %q", cmd.Use)
	}
	for _, sub := range []string{"list", "get", "set", "reset", "path"} {
		child, _, err := rootCmd.Find([]string{"settings", sub})
		if err != nil {
			t.Fatalf("settings %s: %v", sub, err)
		}
		if child == cmd {
			t.Fatalf("settings %s not registered", sub)
		}
	}
}

func TestSelfUpdateCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"self-update"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Flags().Lookup("check") == nil {
		t.Fatal("self-update is missing --check")
	}
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMakeAliasStrUsesRuntimeEnvironmentOnly(t *testing.T) {
	alias := makeAliasStr()

	required := []string{
		"commitron comment --diff",
		"COMMITRON_DIFF=$(git diff --cached)",
	}
	for _, want := range required {
		if !strings.Contains(alias, want) {
			t.Fatalf("makeAliasStr() missing %q in:\n%s", want, alias)
		}
	}

	forbidden := []string{
		" -ak ",
		" --access_key ",
		" -sk ",
		" --secret_key ",
		" -endpoint ",
		" --endpoint ",
		"Enter Access Key",
		"Enter Secret Key",
		"Enter Doubao Endpoint",
	}
	for _, bad := range forbidden {
		if strings.Contains(alias, bad) {
			t.Fatalf("makeAliasStr() contains %q in:\n%s", bad, alias)
		}
	}
}

func TestWriteGitConfigWithAliasRestrictsPermissions(t *testing.T) {
	gitConfigPath := filepath.Join(t.TempDir(), "gitconfig")
	original := []byte("[user]\n    name = Test User\n")
	alias := makeAliasStr()

	if err := writeGitConfigWithAlias(gitConfigPath, original, alias); err != nil {
		t.Fatalf("writeGitConfigWithAlias() error = %v", err)
	}

	got, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", gitConfigPath, err)
	}
	if !strings.HasPrefix(string(got), string(original)) {
		t.Fatalf("written config = %q, want prefix %q", got, original)
	}
	if !strings.Contains(string(got), "commitron comment --diff") {
		t.Fatalf("written config missing commitron alias in:\n%s", got)
	}

	info, err := os.Stat(gitConfigPath)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", gitConfigPath, err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0o600 {
		t.Fatalf("git config mode = %v, want 0600", gotMode)
	}
}

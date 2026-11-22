package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteSSHKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	ssh := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(ssh, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"id_a", "id_a.pub", "id_b", "id_b.pub"} {
		if err := os.WriteFile(filepath.Join(ssh, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, directive := completeSSHKeys(nil, nil, "")
	want := []string{filepath.Join(ssh, "id_a"), filepath.Join(ssh, "id_b")}
	if !slices.Equal(got, want) || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("completeSSHKeys = %q, %v, want %q", got, directive, want)
	}

	got, directive = completeSSHKeys(nil, nil, filepath.Join(ssh, "id_b"))
	if !slices.Equal(got, want[1:]) {
		t.Errorf("prefix filter = %q", got)
	}

	got, directive = completeSSHKeys(nil, nil, "/elsewhere/")
	if got != nil || directive != cobra.ShellCompDirectiveDefault {
		t.Errorf("unknown paths should fall back to file completion, got %q, %v", got, directive)
	}
}

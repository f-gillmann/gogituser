package keys

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func touch(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSSHKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	ssh := filepath.Join(home, ".ssh")
	touch(t,
		filepath.Join(ssh, "id_ed25519"), filepath.Join(ssh, "id_ed25519.pub"),
		filepath.Join(ssh, "work", "id_work"), filepath.Join(ssh, "work", "id_work.pub"),
		filepath.Join(ssh, "work", "deep", "id_deep"), filepath.Join(ssh, "work", "deep", "id_deep.pub"),
		filepath.Join(ssh, "no_pub_key"),
		filepath.Join(ssh, "orphan.pub"),
		filepath.Join(ssh, "config"),
	)

	want := []SSHKey{
		{Name: "id_ed25519", Path: "~/.ssh/id_ed25519"},
		{Name: "work/id_work", Path: "~/.ssh/work/id_work"},
	}
	if got := SSHKeys(); !slices.Equal(got, want) {
		t.Errorf("SSHKeys = %+v, want %+v", got, want)
	}
}

func TestCompletePath(t *testing.T) {
	dir := t.TempDir()
	touch(t,
		filepath.Join(dir, "id_a"), filepath.Join(dir, "id_a.pub"), filepath.Join(dir, "id_b"),
		filepath.Join(dir, ".hidden"), filepath.Join(dir, "sub", "key"),
	)

	tests := map[string][]string{
		dir + "/":     {dir + "/id_a", dir + "/id_b", dir + "/sub/"},
		dir + "/id_":  {dir + "/id_a", dir + "/id_b"},
		dir + "/s":    {dir + "/sub/"},
		dir + "/.":    {dir + "/.hidden"},
		dir + "/nope": nil,
		"no-slash":    nil,
	}
	for typed, want := range tests {
		if got := CompletePath(typed); !slices.Equal(got, want) {
			t.Errorf("CompletePath(%q) = %q, want %q", typed, got, want)
		}
	}
}

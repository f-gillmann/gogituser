package config

import (
	"os"
	"path/filepath"
	"testing"
)

func fakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func TestHomePaths(t *testing.T) {
	home := fakeHome(t)

	if got := ExpandHome("~/.ssh/id"); got != filepath.Join(home, ".ssh/id") {
		t.Errorf("ExpandHome = %q", got)
	}
	if got := ContractHome(filepath.Join(home, ".ssh/id")); got != "~/.ssh/id" {
		t.Errorf("ContractHome = %q", got)
	}
	for _, path := range []string{"/etc/key", home + "other/id", "relative/id"} {
		if got := ContractHome(path); got != path {
			t.Errorf("ContractHome(%q) = %q, want it unchanged", path, got)
		}
	}
}

func TestNormalized(t *testing.T) {
	home := fakeHome(t)
	got := Profile{
		Name:       " work ",
		GitName:    "  Me ",
		GitEmail:   " me@work.test ",
		SigningKey: " ABCD ",
		SSHKey:     " " + filepath.Join(home, ".ssh/id_work") + " ",
	}.Normalized()

	want := Profile{Name: "work", GitName: "Me", GitEmail: "me@work.test", SigningKey: "ABCD", SSHKey: "~/.ssh/id_work"}
	if got != want {
		t.Errorf("Normalized = %+v, want %+v", got, want)
	}
}

func TestCheckKeyPath(t *testing.T) {
	home := fakeHome(t)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"~/.ssh/id":                   "",
		filepath.Join(sshDir, "id"):   "",
		"id":                          "use a full path",
		"./id":                        "use a full path",
		"~/.ssh/missing":              "file not found",
		"~/.ssh":                      "that is a folder",
		filepath.Join(sshDir, "nope"): "file not found",
	}
	for path, want := range tests {
		err := CheckKeyPath(path)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != want {
			t.Errorf("CheckKeyPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil || len(cfg.Profiles) != 0 {
		t.Fatalf("Load without a file = %+v, %v", cfg, err)
	}

	cfg.Upsert(Profile{Name: "work", GitName: "Me", GitEmail: "me@work.test"})
	cfg.Upsert(Profile{Name: "home", GitName: "Me", GitEmail: "me@home.test", SSHKey: "~/.ssh/id"})
	cfg.Upsert(Profile{Name: "work", GitName: "Me", GitEmail: "new@work.test"})
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Profiles) != 2 || loaded.Find("work").GitEmail != "new@work.test" {
		t.Fatalf("loaded = %+v", loaded.Profiles)
	}
	if !loaded.Remove("home") || loaded.Find("home") != nil || loaded.Remove("home") {
		t.Error("Remove did not remove exactly once")
	}
}

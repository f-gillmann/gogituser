// Package gittest has helpers for tests that run real git commands.
package gittest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Isolate keeps the user's own git config and pager out of a test.
func Isolate(t testing.TB) {
	t.Helper()
	global := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(global, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_PAGER", "cat")
	t.Setenv("MANPAGER", "cat")
}

// Run runs a command in dir and fails the test when it fails.
func Run(t testing.TB, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// NewRepo creates an empty repo and returns its path.
func NewRepo(t testing.TB) string {
	t.Helper()
	repo := t.TempDir()
	Run(t, repo, "git", "init", "-q")
	return repo
}

// LocalConfig reads a key from the repo's own config, "" when unset.
func LocalConfig(t testing.TB, repo, key string) string {
	t.Helper()
	out, _ := exec.Command("git", "-C", repo, "config", "--local", "--get", key).Output()
	return strings.TrimSpace(string(out))
}

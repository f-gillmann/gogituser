package cmd

import (
	"errors"
	"path/filepath"
	"testing"

	"gogituser/internal/config"
	"gogituser/internal/gitconfig"
	"gogituser/internal/gittest"
)

func TestRunGitAsClone(t *testing.T) {
	gittest.Isolate(t)
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	gittest.Run(t, dir, "git", "init", "-q", "--bare", "src.git")
	p := config.Profile{Name: "work", GitName: "Me", GitEmail: "me@work.test", SSHKey: "~/.ssh/id work"}

	t.Run("trailing slash", func(t *testing.T) {
		t.Chdir(dir)
		if err := runGitAs(p, []string{"clone", "-q", filepath.Join(dir, "src.git") + "/", "a"}); err != nil {
			t.Fatal(err)
		}
		if got := gittest.LocalConfig(t, filepath.Join(dir, "a"), "user.email"); got != "me@work.test" {
			t.Errorf("user.email = %q", got)
		}
		wantSSH := "ssh -i '" + config.ExpandHome("~/.ssh/id work") + "' -o IdentitiesOnly=yes"
		if got := gittest.LocalConfig(t, filepath.Join(dir, "a"), "core.sshCommand"); got != wantSSH {
			t.Errorf("core.sshCommand = %q, want %q", got, wantSSH)
		}
	})

	t.Run("global option before clone", func(t *testing.T) {
		if err := runGitAs(p, []string{"-C", dir, "clone", "-q", filepath.Join(dir, "src.git"), "b"}); err != nil {
			t.Fatal(err)
		}
		if got := gittest.LocalConfig(t, filepath.Join(dir, "b"), gitconfig.ActiveProfileKey); got != "work" {
			t.Errorf("ggu.profile = %q", got)
		}
	})

	t.Run("help does not panic", func(t *testing.T) {
		err := runGitAs(p, []string{"clone", "--help"})
		var exitCode ExitCodeError
		if err != nil && !errors.As(err, &exitCode) {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("git exit code is passed on", func(t *testing.T) {
		err := runGitAs(p, []string{"clone", "-q", filepath.Join(dir, "missing.git"), "c"})
		var exitCode ExitCodeError
		if !errors.As(err, &exitCode) || exitCode == 0 {
			t.Errorf("err = %v, want a non-zero ExitCodeError", err)
		}
	})
}

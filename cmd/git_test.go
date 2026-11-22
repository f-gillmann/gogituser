package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gogituser/config"
)

// isolateGit keeps the user's own git config out of tests that run git.
func isolateGit(t *testing.T) {
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

func run(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func localConfig(t *testing.T, repo, key string) string {
	t.Helper()
	out, _ := exec.Command("git", "-C", repo, "config", "--local", "--get", key).Output()
	return strings.TrimSpace(string(out))
}

func TestSubcommandIndex(t *testing.T) {
	tests := []struct {
		args []string
		want int
	}{
		{[]string{"clone", "url"}, 0},
		{[]string{"-C", "dir", "clone", "url"}, 2},
		{[]string{"-c", "a.b=c", "--no-pager", "log"}, 3},
		{[]string{"--no-pager"}, -1},
		{nil, -1},
	}
	for _, tt := range tests {
		if got := subcommandIndex(tt.args); got != tt.want {
			t.Errorf("subcommandIndex(%q) = %d, want %d", tt.args, got, tt.want)
		}
	}
}

func TestWithCloneConfig(t *testing.T) {
	p := config.Profile{Name: "work", GitName: "Me", GitEmail: "me@work.test"}

	args, cloning := withCloneConfig(p, []string{"-C", "dir", "clone", "url"})
	if !cloning {
		t.Fatal("clone not detected after -C")
	}
	if args[2] != "clone" || args[3] != "--config" || args[len(args)-1] != "url" {
		t.Errorf("--config not inserted right after clone: %q", args)
	}
	for _, want := range []string{"user.name=Me", "user.email=me@work.test", "ggu.profile=work"} {
		if !slices.Contains(args, want) {
			t.Errorf("missing %q in %q", want, args)
		}
	}
	if slices.Contains(args, "commit.gpgsign=") {
		t.Errorf("empty settings must not be passed to clone: %q", args)
	}

	original := []string{"log", "--oneline"}
	if got, cloning := withCloneConfig(p, original); cloning || !slices.Equal(got, original) {
		t.Errorf("non-clone args changed: %q", got)
	}
}

func TestProfileOverridesClearLeftovers(t *testing.T) {
	isolateGit(t)

	plain := profileOverrides(config.Profile{Name: "p", GitName: "P", GitEmail: "p@x"})
	for _, want := range []string{"commit.gpgsign=false", "tag.gpgsign=false", "core.sshCommand=ssh"} {
		if !slices.Contains(plain, want) {
			t.Errorf("missing %q in %q", want, plain)
		}
	}
	if slices.ContainsFunc(plain, func(s string) bool { return strings.HasPrefix(s, "user.signingkey=") }) {
		t.Errorf("empty signing key must not be passed: %q", plain)
	}

	run(t, "", "git", "config", "--global", "core.sshCommand", "ssh -v")
	if overrides := profileOverrides(config.Profile{}); !slices.Contains(overrides, "core.sshCommand=ssh -v") {
		t.Errorf("global core.sshCommand not kept: %q", overrides)
	}
}

func TestRunGitAsClone(t *testing.T) {
	isolateGit(t)
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q", "--bare", "src.git")
	p := config.Profile{Name: "work", GitName: "Me", GitEmail: "me@work.test", SSHKey: "~/.ssh/id work"}

	t.Run("trailing slash", func(t *testing.T) {
		t.Chdir(dir)
		if err := runGitAs(p, []string{"clone", "-q", filepath.Join(dir, "src.git") + "/", "a"}); err != nil {
			t.Fatal(err)
		}
		if got := localConfig(t, filepath.Join(dir, "a"), "user.email"); got != "me@work.test" {
			t.Errorf("user.email = %q", got)
		}
		wantSSH := "ssh -i '" + config.ExpandHome("~/.ssh/id work") + "' -o IdentitiesOnly=yes"
		if got := localConfig(t, filepath.Join(dir, "a"), "core.sshCommand"); got != wantSSH {
			t.Errorf("core.sshCommand = %q, want %q", got, wantSSH)
		}
	})

	t.Run("global option before clone", func(t *testing.T) {
		if err := runGitAs(p, []string{"-C", dir, "clone", "-q", filepath.Join(dir, "src.git"), "b"}); err != nil {
			t.Fatal(err)
		}
		if got := localConfig(t, filepath.Join(dir, "b"), activeProfileKey); got != "work" {
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

package gitconfig

import (
	"slices"
	"strings"
	"testing"

	"gogituser/internal/config"
	"gogituser/internal/gittest"
)

func TestApplySwitchesCleanly(t *testing.T) {
	gittest.Isolate(t)
	repo := gittest.NewRepo(t)

	keyed := config.Profile{Name: "work", GitName: "W", GitEmail: "w@x", SigningKey: "ABCD", SSHKey: "/keys/id"}
	if err := Apply(keyed, repo); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"user.signingkey": "ABCD",
		"commit.gpgsign":  "true",
		"tag.gpgsign":     "true",
		"core.sshCommand": "ssh -i '/keys/id' -o IdentitiesOnly=yes",
	} {
		if got := gittest.LocalConfig(t, repo, key); got != want {
			t.Errorf("after work: %s = %q, want %q", key, got, want)
		}
	}
	if got := ActiveProfile(repo); got != "work" {
		t.Errorf("ActiveProfile = %q", got)
	}

	plain := config.Profile{Name: "plain", GitName: "P", GitEmail: "p@x"}
	if err := Apply(plain, repo); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"user.signingkey", "commit.gpgsign", "tag.gpgsign", "core.sshCommand"} {
		if got := gittest.LocalConfig(t, repo, key); got != "" {
			t.Errorf("after plain: %s = %q, want it unset", key, got)
		}
	}
	if got := gittest.LocalConfig(t, repo, "user.email"); got != "p@x" {
		t.Errorf("user.email = %q", got)
	}
}

func TestApplyOutsideRepo(t *testing.T) {
	gittest.Isolate(t)
	err := Apply(config.Profile{Name: "x", GitName: "X", GitEmail: "x@x"}, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "not inside a git repository") {
		t.Errorf("err = %v", err)
	}
	if got := ActiveProfile(t.TempDir()); got != "" {
		t.Errorf("ActiveProfile outside a repo = %q", got)
	}
}

func TestOriginUsesHTTPS(t *testing.T) {
	gittest.Isolate(t)
	repo := gittest.NewRepo(t)
	if OriginUsesHTTPS(repo) {
		t.Error("no origin should not count as HTTPS")
	}

	gittest.Run(t, repo, "git", "remote", "add", "origin", "https://github.com/me/repo")
	if !OriginUsesHTTPS(repo) {
		t.Error("https origin not detected")
	}

	gittest.Run(t, repo, "git", "remote", "set-url", "origin", "git@github.com:me/repo.git")
	if OriginUsesHTTPS(repo) {
		t.Error("ssh origin reported as HTTPS")
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("/a b/it's"); got != `'/a b/it'\''s'` {
		t.Errorf("shellQuote = %s", got)
	}
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

	args, cloning := WithCloneConfig(p, []string{"-C", "dir", "clone", "url"})
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
	if got, cloning := WithCloneConfig(p, original); cloning || !slices.Equal(got, original) {
		t.Errorf("non-clone args changed: %q", got)
	}
}

func TestOverridesClearLeftovers(t *testing.T) {
	gittest.Isolate(t)

	plain := Overrides(config.Profile{Name: "p", GitName: "P", GitEmail: "p@x"})
	for _, want := range []string{"commit.gpgsign=false", "tag.gpgsign=false", "core.sshCommand=ssh"} {
		if !slices.Contains(plain, want) {
			t.Errorf("missing %q in %q", want, plain)
		}
	}
	if slices.ContainsFunc(plain, func(s string) bool { return strings.HasPrefix(s, "user.signingkey=") }) {
		t.Errorf("empty signing key must not be passed: %q", plain)
	}

	gittest.Run(t, "", "git", "config", "--global", "core.sshCommand", "ssh -v")
	if overrides := Overrides(config.Profile{}); !slices.Contains(overrides, "core.sshCommand=ssh -v") {
		t.Errorf("global core.sshCommand not kept: %q", overrides)
	}
}

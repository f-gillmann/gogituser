package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"gogituser/config"
)

func TestApplyGitConfigSwitchesCleanly(t *testing.T) {
	isolateGit(t)
	repo := t.TempDir()
	run(t, repo, "git", "init", "-q")
	t.Chdir(repo)

	keyed := config.Profile{Name: "work", GitName: "W", GitEmail: "w@x", SigningKey: "ABCD", SSHKey: "/keys/id"}
	if err := applyGitConfig(keyed, ""); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"user.signingkey": "ABCD",
		"commit.gpgsign":  "true",
		"tag.gpgsign":     "true",
		"core.sshCommand": "ssh -i '/keys/id' -o IdentitiesOnly=yes",
	} {
		if got := localConfig(t, repo, key); got != want {
			t.Errorf("after work: %s = %q, want %q", key, got, want)
		}
	}
	if got := activeProfile(); got != "work" {
		t.Errorf("activeProfile = %q", got)
	}

	plain := config.Profile{Name: "plain", GitName: "P", GitEmail: "p@x"}
	if err := applyGitConfig(plain, ""); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"user.signingkey", "commit.gpgsign", "tag.gpgsign", "core.sshCommand"} {
		if got := localConfig(t, repo, key); got != "" {
			t.Errorf("after plain: %s = %q, want it unset", key, got)
		}
	}
	if got := localConfig(t, repo, "user.email"); got != "p@x" {
		t.Errorf("user.email = %q", got)
	}
}

func TestApplyGitConfigOutsideRepo(t *testing.T) {
	isolateGit(t)
	t.Chdir(t.TempDir())
	err := applyGitConfig(config.Profile{Name: "x", GitName: "X", GitEmail: "x@x"}, "")
	if err == nil || !strings.Contains(err.Error(), "not inside a git repository") {
		t.Errorf("err = %v", err)
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("/a b/it's"); got != `'/a b/it'\''s'` {
		t.Errorf("shellQuote = %s", got)
	}
}

func newFlagCommand(t *testing.T, args ...string) (*cobra.Command, *profileFlags) {
	t.Helper()
	cmd := &cobra.Command{}
	var flags profileFlags
	flags.register(cmd)
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	return cmd, &flags
}

func TestProfileFlagsApplyOnlyPassedFlags(t *testing.T) {
	existing := config.Profile{Name: "w", GitName: "W", GitEmail: "old@x", SigningKey: "ABCD", SSHKey: "~/.ssh/id"}

	cmd, flags := newFlagCommand(t, "--git-email", "new@x", "--signing-key", "")
	p := existing
	if !flags.applyTo(cmd, &p) {
		t.Fatal("applyTo reported no change")
	}
	want := config.Profile{Name: "w", GitName: "W", GitEmail: "new@x", SigningKey: "", SSHKey: "~/.ssh/id"}
	if p != want {
		t.Errorf("got %+v, want %+v", p, want)
	}

	cmd, flags = newFlagCommand(t)
	p = existing
	if flags.applyTo(cmd, &p) || p != existing {
		t.Errorf("no flags changed the profile: %+v", p)
	}
}

func TestCheckProfile(t *testing.T) {
	if err := checkProfile(config.Profile{}); err == nil ||
		err.Error() != "missing a profile name, --git-name, --git-email" {
		t.Errorf("err = %v", err)
	}
	if err := checkProfile(config.Profile{Name: "w", GitName: "W", GitEmail: "w@x", SSHKey: "relative"}); err == nil ||
		!strings.Contains(err.Error(), "use a full path") {
		t.Errorf("err = %v", err)
	}
	if err := checkProfile(config.Profile{Name: "w", GitName: "W", GitEmail: "w@x"}); err != nil {
		t.Errorf("valid profile rejected: %v", err)
	}
}

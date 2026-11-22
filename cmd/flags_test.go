package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/f-gillmann/gogituser/internal/config"
)

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

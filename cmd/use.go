package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
)

var useCmd = &cobra.Command{
	Use:               "use [profile]",
	Short:             "Apply a profile to the current git repository, or pick one from a list",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeProfiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			p, err := loadProfile(args[0])
			if err != nil {
				return err
			}
			return applyGitConfig(*p, "")
		}

		if err := requireGitRepo(gitIn("")); err != nil {
			return err
		}
		if !interactive() {
			return fmt.Errorf("no terminal for the picker, pass a profile name")
		}
		p, ok, err := pickProfile()
		if err != nil || !ok {
			return err
		}
		return applyGitConfig(p, "")
	},
}

type gitSetting struct {
	key, value string
	// cleared gives the value "ggu git" passes when the profile leaves this setting empty,
	// so a leftover from another profile in the repo config cannot take effect.
	cleared func() string
}

type gitRunner func(args ...string) *exec.Cmd

// activeProfileKey records in the repo's git config which profile was applied.
const activeProfileKey = "ggu.profile"

// gitConfigKeyMissing is the exit code of "git config --unset" when the key is not set.
const gitConfigKeyMissing = 5

func loadProfile(name string) (*config.Profile, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	p := cfg.Find(name)
	if p == nil {
		return nil, fmt.Errorf("unknown profile %q", name)
	}
	return p, nil
}

func pickProfile() (config.Profile, bool, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Profile{}, false, err
	}
	if len(cfg.Profiles) == 0 {
		return config.Profile{}, false, fmt.Errorf("no profiles found, run 'ggu add' to create one")
	}
	return tui.Pick(cfg.Profiles)
}

func requireGitRepo(git gitRunner) error {
	if out, err := git("rev-parse", "--git-dir").CombinedOutput(); err != nil {
		return fmt.Errorf("not inside a git repository: %s", out)
	}
	return nil
}

// profileSettings returns every git setting a profile controls. An empty value means
// the setting should be unset so no leftover from a previous profile stays active.
func profileSettings(p config.Profile) []gitSetting {
	sshCommand := ""
	if p.SSHKey != "" {
		sshCommand = fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", shellQuote(config.ExpandHome(p.SSHKey)))
	}
	sign := ""
	if p.SigningKey != "" {
		sign = "true"
	}
	return []gitSetting{
		{key: "user.name", value: p.GitName},
		{key: "user.email", value: p.GitEmail},
		{key: "user.signingkey", value: p.SigningKey},
		{key: "commit.gpgsign", value: sign, cleared: signingOff},
		{key: "tag.gpgsign", value: sign, cleared: signingOff},
		{key: "core.sshCommand", value: sshCommand, cleared: globalSSHCommand},
		{key: activeProfileKey, value: p.Name},
	}
}

func signingOff() string { return "false" }

// globalSSHCommand is what ssh would be without any repo config. An empty
// core.sshCommand is no option, git would try to run a program named "".
func globalSSHCommand() string {
	out, err := exec.Command("git", "config", "--global", "--get", "core.sshCommand").Output()
	if command := strings.TrimSpace(string(out)); err == nil && command != "" {
		return command
	}
	return "ssh"
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func writeSetting(git gitRunner, s gitSetting) error {
	if s.value != "" {
		if err := git("config", s.key, s.value).Run(); err != nil {
			return fmt.Errorf("git config %s: %w", s.key, err)
		}
		return nil
	}

	err := git("config", "--unset", s.key).Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == gitConfigKeyMissing {
		return nil
	}
	if err != nil {
		return fmt.Errorf("git config --unset %s: %w", s.key, err)
	}
	return nil
}

// activeProfile returns the profile name applied to the current repo, or "" outside a repo.
func activeProfile() string {
	out, err := gitIn("")("config", "--local", "--get", activeProfileKey).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitIn(dir string) gitRunner {
	return func(args ...string) *exec.Cmd {
		if dir != "" {
			args = append([]string{"-C", dir}, args...)
		}
		return exec.Command("git", args...)
	}
}

// applyGitConfig writes the profile's identity into git config.
// dir is the target repo; pass "" to use the current working directory.
func applyGitConfig(p config.Profile, dir string) error {
	git := gitIn(dir)

	if err := requireGitRepo(git); err != nil {
		return err
	}

	for _, s := range profileSettings(p) {
		if err := writeSetting(git, s); err != nil {
			return err
		}
	}

	fmt.Println(tui.UsingProfile(p.Name))
	fmt.Println(tui.ProfileDetails(p))
	return nil
}

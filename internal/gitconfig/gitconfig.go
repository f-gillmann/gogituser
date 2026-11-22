// Package gitconfig turns a profile into git settings and writes them into repos.
package gitconfig

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"gogituser/internal/config"
)

// ActiveProfileKey records in the repo's git config which profile was applied.
const ActiveProfileKey = "ggu.profile"

// keyMissingExitCode is the exit code of "git config --unset" when the key is not set.
const keyMissingExitCode = 5

type setting struct {
	key, value string
	// cleared gives the value a one-off git run gets when the profile leaves this setting
	// empty, so a leftover from another profile in the repo config cannot take effect.
	cleared func() string
}

// settings returns every git setting a profile controls. An empty value means the
// setting gets unset, so nothing from a previous profile stays active.
func settings(p config.Profile) []setting {
	sshCommand := ""
	if p.SSHKey != "" {
		sshCommand = fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", shellQuote(config.ExpandHome(p.SSHKey)))
	}
	sign := ""
	if p.SigningKey != "" {
		sign = "true"
	}
	return []setting{
		{key: "user.name", value: p.GitName},
		{key: "user.email", value: p.GitEmail},
		{key: "user.signingkey", value: p.SigningKey},
		{key: "commit.gpgsign", value: sign, cleared: signingOff},
		{key: "tag.gpgsign", value: sign, cleared: signingOff},
		{key: "core.sshCommand", value: sshCommand, cleared: globalSSHCommand},
		{key: ActiveProfileKey, value: p.Name},
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

type runner func(args ...string) *exec.Cmd

// in runs git in dir, or in the current directory when dir is "".
func in(dir string) runner {
	return func(args ...string) *exec.Cmd {
		if dir != "" {
			args = append([]string{"-C", dir}, args...)
		}
		return exec.Command("git", args...)
	}
}

// RequireRepo fails when dir ("" for the current directory) is not inside a git repo.
func RequireRepo(dir string) error {
	if out, err := in(dir)("rev-parse", "--git-dir").CombinedOutput(); err != nil {
		return fmt.Errorf("not inside a git repository: %s", out)
	}
	return nil
}

// Apply writes the profile into the repo config of dir ("" for the current directory).
func Apply(p config.Profile, dir string) error {
	if err := RequireRepo(dir); err != nil {
		return err
	}
	git := in(dir)
	for _, s := range settings(p) {
		if err := write(git, s); err != nil {
			return err
		}
	}
	return nil
}

func write(git runner, s setting) error {
	if s.value != "" {
		if err := git("config", s.key, s.value).Run(); err != nil {
			return fmt.Errorf("git config %s: %w", s.key, err)
		}
		return nil
	}

	err := git("config", "--unset", s.key).Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == keyMissingExitCode {
		return nil
	}
	if err != nil {
		return fmt.Errorf("git config --unset %s: %w", s.key, err)
	}
	return nil
}

// ActiveProfile returns the profile applied to the repo in dir, or "" outside a repo.
func ActiveProfile(dir string) string {
	out, err := in(dir)("config", "--local", "--get", ActiveProfileKey).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// OriginUsesHTTPS reports whether pushes and pulls of origin bypass ssh, and with it the profile's SSH key.
func OriginUsesHTTPS(dir string) bool {
	out, err := in(dir)("remote", "get-url", "origin").Output()
	if err != nil {
		return false
	}
	url := strings.TrimSpace(string(out))
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")
}

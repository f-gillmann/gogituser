package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gogituser/config"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage git remotes using a profile's SSH host alias",
}

var remoteSetURLCmd = &cobra.Command{
	Use:               "set-url <profile> [remote]",
	Short:             "Rewrite an existing remote URL to use the profile's SSH host alias",
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: completeProfiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := loadProfileWithHost(args[0])
		if err != nil {
			return err
		}

		remote := remoteArg(args, 1)

		if err := requireGitRepo(); err != nil {
			return err
		}

		currentURL, err := getRemoteURL(remote)
		if err != nil {
			return err
		}

		newURL, err := rewriteHost(currentURL, p.SSHHost)
		if err != nil {
			return err
		}

		if err := exec.Command("git", "remote", "set-url", remote, newURL).Run(); err != nil {
			return fmt.Errorf("git remote set-url: %w", err)
		}

		fmt.Printf("Remote %q updated:\n  %s\n  → %s\n", remote, currentURL, newURL)
		return nil
	},
}

var remoteAddCmd = &cobra.Command{
	Use:               "add <profile> <url> [remote]",
	Short:             "Add a new remote, rewriting the URL to use the profile's SSH host alias",
	Args:              cobra.RangeArgs(2, 3),
	ValidArgsFunction: completeProfiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := loadProfileWithHost(args[0])
		if err != nil {
			return err
		}

		rawURL := args[1]
		remote := remoteArg(args, 2)

		if err := requireGitRepo(); err != nil {
			return err
		}

		newURL, err := rewriteHost(rawURL, p.SSHHost)
		if err != nil {
			return err
		}

		if err := exec.Command("git", "remote", "add", remote, newURL).Run(); err != nil {
			return fmt.Errorf("git remote add: %w", err)
		}

		fmt.Printf("Remote %q added: %s\n", remote, newURL)
		return nil
	},
}

func init() {
	remoteCmd.AddCommand(remoteSetURLCmd, remoteAddCmd)
	rootCmd.AddCommand(remoteCmd)
}

// ---- helpers -------------------------------------------------------------

func loadProfileWithHost(name string) (*config.Profile, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	p := cfg.Find(name)
	if p == nil {
		return nil, fmt.Errorf("unknown profile %q", name)
	}
	if p.SSHHost == "" {
		return nil, fmt.Errorf("profile %q has no ssh_host set — run 'gogituser update %s' to add one", p.Name, p.Name)
	}
	if err := warnIfSSHHostMissing(p.SSHHost); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: %s\n", err)
	}
	return p, nil
}

func remoteArg(args []string, idx int) string {
	if idx < len(args) {
		return args[idx]
	}
	return "origin"
}

func requireGitRepo() error {
	if out, err := exec.Command("git", "rev-parse", "--git-dir").CombinedOutput(); err != nil {
		return fmt.Errorf("not inside a git repository: %s", out)
	}
	return nil
}

func getRemoteURL(remote string) (string, error) {
	out, err := exec.Command("git", "remote", "get-url", remote).Output()
	if err != nil {
		return "", fmt.Errorf("remote %q not found", remote)
	}
	return strings.TrimSpace(string(out)), nil
}

// rewriteHost replaces the hostname in an SSH remote URL with newHost.
func rewriteHost(remoteURL, newHost string) (string, error) {
	if strings.Contains(remoteURL, "://") {
		if !strings.HasPrefix(remoteURL, "ssh://") {
			return "", fmt.Errorf("remote %q uses a non-SSH URL — SSH host aliases only work with SSH remotes", remoteURL)
		}
		rest := strings.TrimPrefix(remoteURL, "ssh://")
		if at := strings.Index(rest, "@"); at >= 0 {
			rest = rest[at+1:]
		}
		slash := strings.Index(rest, "/")
		if slash < 0 {
			return "", fmt.Errorf("cannot parse remote URL %q", remoteURL)
		}
		return "ssh://git@" + newHost + rest[slash:], nil
	}

	// SCP-style: [user@]host:path
	colon := strings.Index(remoteURL, ":")
	if colon < 0 {
		return "", fmt.Errorf("cannot parse remote URL %q", remoteURL)
	}
	return "git@" + newHost + ":" + remoteURL[colon+1:], nil
}

// warnIfSSHHostMissing checks ~/.ssh/config for a matching Host entry.
func warnIfSSHHostMissing(alias string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	f, err := os.Open(filepath.Join(home, ".ssh", "config"))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("~/.ssh/config does not exist — create it with a Host %s entry for the alias to work", alias)
		}
		return nil
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && strings.EqualFold(fields[0], "Host") && fields[1] == alias {
			return nil
		}
	}
	return fmt.Errorf("no 'Host %s' entry found in ~/.ssh/config", alias)
}

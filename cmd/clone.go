package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gogituser/config"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <profile> <repo> [directory]",
	Short: "Clone a repository using a profile's SSH key",
	Args:  cobra.RangeArgs(2, 3),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeProfiles(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		p := cfg.Find(args[0])
		if p == nil {
			return fmt.Errorf("unknown profile %q", args[0])
		}

		repoURL := args[1]
		dir := cloneDir(repoURL)
		if len(args) == 3 {
			dir = args[2]
		}

		sshCmd := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", p.SSHKey)
		gitCmd := exec.Command("git", "clone", repoURL, dir)
		gitCmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr

		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}

		return applyGitConfig(*p, dir)
	},
}

// cloneDir derives the default destination directory from a repo URL,
// matching git's own behaviour (last path component, sans .git).
func cloneDir(repoURL string) string {
	base := repoURL
	if i := strings.LastIndexAny(base, ":/"); i >= 0 {
		base = base[i+1:]
	}
	base = strings.TrimSuffix(base, ".git")
	return filepath.Base(base)
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"gogituser/config"
)

var useCmd = &cobra.Command{
	Use:               "use <profile>",
	Short:             "Apply a profile to the current git repository",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProfiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		return applyProfile(args[0])
	},
}

func applyProfile(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	p := cfg.Find(name)
	if p == nil {
		return fmt.Errorf("unknown profile %q", name)
	}
	return applyGitConfig(*p, "")
}

// applyGitConfig writes the profile's identity into git config.
// dir is the target repo; pass "" to use the current working directory.
func applyGitConfig(p config.Profile, dir string) error {
	checkArgs := []string{"rev-parse", "--git-dir"}
	configPrefix := []string{"config"}
	if dir != "" {
		checkArgs = append([]string{"-C", dir}, checkArgs...)
		configPrefix = append([]string{"-C", dir}, configPrefix...)
	}

	if out, err := exec.Command("git", checkArgs...).CombinedOutput(); err != nil {
		return fmt.Errorf("not inside a git repository: %s", out)
	}

	settings := [][]string{
		{"user.name", p.GitName},
		{"user.email", p.Email},
		{"user.signingkey", p.SigningKey},
		{"core.sshCommand", fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", p.SSHKey)},
	}
	for _, kv := range settings {
		args := append(configPrefix, kv...)
		if err := exec.Command("git", args...).Run(); err != nil {
			return fmt.Errorf("git config %s: %w", kv[0], err)
		}
	}

	fmt.Printf("Git identity set for profile %q:\n", p.Name)
	fmt.Printf("  Name        : %s\n", p.GitName)
	fmt.Printf("  Email       : %s\n", p.Email)
	fmt.Printf("  Signing Key : %s\n", p.SigningKey)
	fmt.Printf("  SSH Key     : %s\n", p.SSHKey)
	return nil
}

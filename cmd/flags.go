package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"gogituser/internal/config"
	"gogituser/internal/keys"
)

const (
	flagGitName    = "git-name"
	flagGitEmail   = "git-email"
	flagSigningKey = "signing-key"
	flagSSHKey     = "ssh-key"
)

// profileFlags lets add and update work without the form, e.g. in scripts.
type profileFlags struct {
	gitName, gitEmail, signingKey, sshKey string
}

func (f *profileFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.gitName, flagGitName, "", "name on commits")
	cmd.Flags().StringVar(&f.gitEmail, flagGitEmail, "", "email on commits")
	cmd.Flags().StringVar(&f.signingKey, flagSigningKey, "", "GPG key ID or fingerprint, empty to clear")
	cmd.Flags().StringVar(&f.sshKey, flagSSHKey, "", "full path to the SSH key, empty to clear")

	completions := map[string]cobra.CompletionFunc{
		flagGitName:    completeGitNames,
		flagGitEmail:   completeGitEmails,
		flagSigningKey: completeSigningKeys,
		flagSSHKey:     completeSSHKeys,
	}
	for flag, complete := range completions {
		_ = cmd.RegisterFlagCompletionFunc(flag, complete)
	}
}

// applyTo copies only the flags that were passed, so an empty --ssh-key clears a field
// while leaving out a flag keeps the current value. It reports whether any flag was passed.
func (f *profileFlags) applyTo(cmd *cobra.Command, p *config.Profile) bool {
	fields := []struct {
		flag   string
		value  string
		target *string
	}{
		{flagGitName, f.gitName, &p.GitName},
		{flagGitEmail, f.gitEmail, &p.GitEmail},
		{flagSigningKey, f.signingKey, &p.SigningKey},
		{flagSSHKey, f.sshKey, &p.SSHKey},
	}

	changed := false
	for _, field := range fields {
		if cmd.Flags().Changed(field.flag) {
			*field.target = field.value
			changed = true
		}
	}
	return changed
}

// missingFields lists what a profile still needs, named the way the user would pass it.
func missingFields(p config.Profile) []string {
	var missing []string
	if strings.TrimSpace(p.Name) == "" {
		missing = append(missing, "a profile name")
	}
	if strings.TrimSpace(p.GitName) == "" {
		missing = append(missing, "--"+flagGitName)
	}
	if strings.TrimSpace(p.GitEmail) == "" {
		missing = append(missing, "--"+flagGitEmail)
	}
	return missing
}

// prepareProfile is the single place where form and flag input become what gets saved.
func prepareProfile(p config.Profile) config.Profile {
	p = p.Normalized()
	p.SigningKey = keys.ResolveSigningKey(p.SigningKey)
	return p
}

func checkProfile(p config.Profile) error {
	if missing := missingFields(p); len(missing) > 0 {
		return fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	if p.SSHKey == "" {
		return nil
	}
	if err := config.CheckKeyPath(p.SSHKey); err != nil {
		return fmt.Errorf("--%s %s: %w", flagSSHKey, p.SSHKey, err)
	}
	return nil
}

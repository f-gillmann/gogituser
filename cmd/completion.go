package cmd

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"gogituser/internal/config"
	"gogituser/internal/keys"
)

func completeProfiles(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	names := make([]string, len(cfg.Profiles))
	for i, p := range cfg.Profiles {
		names[i] = p.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeGitNames offers the names on your GPG keys, since commits usually use one of them.
func completeGitNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return uidParts(func(name, _ string) string { return name }), cobra.ShellCompDirectiveNoFileComp
}

// completeGitEmails offers the emails on your GPG keys, which GitHub needs for verified commits.
func completeGitEmails(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return uidParts(func(_, email string) string { return email }), cobra.ShellCompDirectiveNoFileComp
}

func uidParts(pick func(name, email string) string) []string {
	var parts []string
	for _, k := range keys.ListGPG() {
		for _, uid := range k.UIDs {
			if part := pick(keys.SplitUID(uid)); part != "" && !slices.Contains(parts, part) {
				parts = append(parts, part)
			}
		}
	}
	return parts
}

// completeSigningKeys offers key IDs, with the key's identity shown next to each.
func completeSigningKeys(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	var completions []string
	for _, k := range keys.ListGPG() {
		completion := k.KeyID()
		if len(k.UIDs) > 0 {
			completion += "\t" + k.UIDs[0]
		}
		completions = append(completions, completion)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeSSHKeys offers the key pairs in ~/.ssh and falls back to normal file
// completion for anything else. Paths are full, since a "~" from completion
// would not be expanded by every shell.
func completeSSHKeys(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string
	for _, k := range keys.SSHKeys() {
		if path := config.ExpandHome(k.Path); strings.HasPrefix(path, toComplete) {
			completions = append(completions, path)
		}
	}
	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

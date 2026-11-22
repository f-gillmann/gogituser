package cmd

import (
	"github.com/spf13/cobra"

	"gogituser/internal/config"
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

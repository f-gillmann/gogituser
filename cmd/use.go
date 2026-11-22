package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/f-gillmann/gogituser/internal/config"
	"github.com/f-gillmann/gogituser/internal/gitconfig"
	"github.com/f-gillmann/gogituser/internal/terminal"
	"github.com/f-gillmann/gogituser/internal/tui"
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
			return applyProfile(*p)
		}

		if err := gitconfig.RequireRepo(""); err != nil {
			return err
		}
		if !terminal.Interactive() {
			return withHint(errors.New("can't show the picker without a terminal"),
				"run 'ggu use <profile>', 'ggu list' shows your profiles")
		}
		p, ok, err := pickProfile()
		if err != nil || !ok {
			return err
		}
		return applyProfile(p)
	},
}

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

func applyProfile(p config.Profile) error {
	if err := gitconfig.Apply(p, ""); err != nil {
		return err
	}
	fmt.Println(tui.UsingProfile(p.Name))
	fmt.Println(tui.ProfileDetails(p))
	printSSHKeyHint(p)
	return nil
}

// printSSHKeyHint points out that an HTTPS origin never uses the profile's SSH key.
func printSSHKeyHint(p config.Profile) {
	if p.SSHKey != "" && gitconfig.OriginUsesHTTPS("") {
		fmt.Println()
		fmt.Println(tui.Hint("origin uses HTTPS, so the SSH key is not used"))
	}
}

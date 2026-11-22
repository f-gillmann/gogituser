package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/f-gillmann/gogituser/internal/config"
	"github.com/f-gillmann/gogituser/internal/terminal"
	"github.com/f-gillmann/gogituser/internal/tui"
)

var addFlags profileFlags

var addCmd = &cobra.Command{
	Use:   "add [profile]",
	Short: "Add a new profile, with flags or an interactive form",
	Example: `  ggu add
  ggu add work --git-name "Your Name" --git-email you@work.com --ssh-key ~/.ssh/id_work`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var p config.Profile
		if len(args) == 1 {
			p.Name = args[0]
		}
		addFlags.applyTo(cmd, &p)

		if missing := missingFields(p); len(missing) > 0 {
			if !terminal.Interactive() {
				return withHint(fmt.Errorf("can't show the form without a terminal, missing %s", strings.Join(missing, ", ")),
					"run 'ggu add --help' to see all flags")
			}
			var ok bool
			p, ok, err = tui.RunForm(p, false)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		p = prepareProfile(p)
		if err := checkProfile(p); err != nil {
			return err
		}
		if cfg.Find(p.Name) != nil {
			return fmt.Errorf("profile %q already exists, use 'ggu update %s' to change it", p.Name, p.Name)
		}

		cfg.Upsert(p)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Profile %q saved.\n", p.Name)
		return nil
	},
}

func init() {
	addFlags.register(addCmd)
}

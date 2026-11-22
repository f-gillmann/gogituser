package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
)

var updateFlags profileFlags

var updateCmd = &cobra.Command{
	Use:   "update <profile>",
	Short: "Update a profile, with flags or an interactive form",
	Example: `  ggu update work
  ggu update work --git-email new@work.com
  ggu update work --signing-key ""`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProfiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		existing := cfg.Find(args[0])
		if existing == nil {
			return fmt.Errorf("unknown profile %q", args[0])
		}

		p := *existing
		if !updateFlags.applyTo(cmd, &p) {
			if !interactive() {
				return fmt.Errorf("no terminal for the form, pass the fields to change, like --%s", flagGitEmail)
			}
			var ok bool
			p, ok, err = tui.RunForm(p, true)
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
		cfg.Upsert(p)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Profile %q updated.\n", p.Name)
		return nil
	},
}

func init() {
	updateFlags.register(updateCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
)

var updateCmd = &cobra.Command{
	Use:               "update <profile>",
	Short:             "Update an existing profile",
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

		p, ok, err := tui.RunForm(*existing, true)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}

		cfg.Upsert(p)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Profile %q updated.\n", p.Name)
		return nil
	},
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
)

var deleteCmd = &cobra.Command{
	Use:               "delete <profile>",
	Short:             "Delete a profile",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProfiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		name := args[0]
		if cfg.Find(name) == nil {
			return fmt.Errorf("unknown profile %q", name)
		}

		ok, err := tui.Confirm(fmt.Sprintf("Delete profile %q?", name))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}

		cfg.Delete(name)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Profile %q deleted.\n", name)
		return nil
	},
}

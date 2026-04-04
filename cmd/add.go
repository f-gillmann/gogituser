package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new profile",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		p, ok, err := tui.RunForm(config.Profile{}, false)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}

		if cfg.Find(p.Name) != nil {
			return fmt.Errorf("profile %q already exists — use 'gogituser update %s' to modify it", p.Name, p.Name)
		}

		cfg.Upsert(p)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Profile %q saved.\n", p.Name)
		return nil
	},
}

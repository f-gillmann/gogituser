package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles found. Run 'ggu add' to create one.")
			return nil
		}
		fmt.Println(tui.ProfileList(cfg.Profiles, activeProfile()))
		return nil
	},
}

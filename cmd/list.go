package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/internal/config"
	"gogituser/internal/gitconfig"
	"gogituser/internal/tui"
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
		fmt.Println(tui.ProfileList(cfg.Profiles, gitconfig.ActiveProfile("")))
		return nil
	},
}

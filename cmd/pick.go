package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
)

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Interactively pick a profile and apply it",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Profiles) == 0 {
			return fmt.Errorf("no profiles found — run 'gogituser add' to create one")
		}

		p, ok, err := tui.Pick(cfg.Profiles)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		return applyGitConfig(p, "")
	},
}

func init() {
	rootCmd.AddCommand(pickCmd)
}

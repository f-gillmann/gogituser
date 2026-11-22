package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/internal/config"
	"gogituser/internal/terminal"
	"gogituser/internal/tui"
)

var removeYes bool

var removeCmd = &cobra.Command{
	Use:               "remove <profile>",
	Short:             "Remove a profile",
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

		if !removeYes {
			if !terminal.Interactive() {
				return withHint(errors.New("can't ask for confirmation without a terminal"),
					fmt.Sprintf("run 'ggu remove %s --yes' to remove it anyway", name))
			}
			ok, err := tui.Confirm(fmt.Sprintf("Remove profile %q?", name))
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		cfg.Remove(name)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Profile %q removed.\n", name)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "remove without asking")
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gogituser/config"
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
			fmt.Println("No profiles found. Run 'gogituser add' to create one.")
			return nil
		}
		for _, p := range cfg.Profiles {
			fmt.Printf("%s\n  name:        %s\n  email:       %s\n  signing key: %s\n  ssh key:     %s\n\n",
				p.Name, p.GitName, p.Email, p.SigningKey, p.SSHKey)
		}
		return nil
	},
}

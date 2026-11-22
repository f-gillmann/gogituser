package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "ggu",
	Short:         "Manage and apply git identity profiles",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// ExitCodeError carries a child process exit code that should be passed on without extra output.
type ExitCodeError int

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit status %d", int(e))
}

// withHint adds a line telling the user what to run next.
func withHint(err error, hint string) error {
	return fmt.Errorf("%w\n%s", err, hint)
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(useCmd, listCmd, addCmd, updateCmd, removeCmd)
}

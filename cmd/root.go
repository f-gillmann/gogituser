package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is set at build time from the git tag, see the justfile and .goreleaser.yaml.
var version = "dev"

// moduleVersion covers "go install", which sets no version but records the module's tag in the binary.
func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if version == "dev" && ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

var rootCmd = &cobra.Command{
	Use:           "ggu",
	Short:         "Manage and apply git identity profiles",
	Version:       moduleVersion(),
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

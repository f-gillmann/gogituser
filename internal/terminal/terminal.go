// Package terminal tells whether ggu can show its interactive TUI.
package terminal

import (
	"os"

	"github.com/mattn/go-isatty"
)

// Interactive reports whether the TUI can run, which needs a terminal on both ends.
func Interactive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

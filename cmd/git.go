package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"slices"

	"github.com/spf13/cobra"

	"gogituser/internal/config"
	"gogituser/internal/gitconfig"
	"gogituser/internal/tui"
)

var gitCmd = &cobra.Command{
	Use:                "git <profile> <git args...>",
	Short:              "Run any git command as a profile",
	DisableFlagParsing: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeProfiles(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 || args[0] == "-h" || args[0] == "--help" {
			return cmd.Help()
		}
		p, err := loadProfile(args[0])
		if err != nil {
			return err
		}
		return runGitAs(*p, args[1:])
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}

func runGitAs(p config.Profile, args []string) error {
	args, cloning := gitconfig.WithCloneConfig(p, args)

	// git handles Ctrl+C itself. ggu has to outlive it, otherwise the shell takes the
	// terminal back while git's pager or editor is still running.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)

	git := exec.Command("git", append(gitconfig.Overrides(p), args...)...)
	git.Stdin, git.Stdout, git.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := git.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExitCodeError(exitErr.ExitCode())
		}
		return err
	}

	if cloning && !asksForHelp(args) {
		fmt.Println(tui.UsingProfile(p.Name))
	}
	return nil
}

func asksForHelp(args []string) bool {
	return slices.Contains(args, "--help") || slices.Contains(args, "-h")
}

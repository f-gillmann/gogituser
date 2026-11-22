package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"gogituser/config"
	"gogituser/tui"
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
	args, cloning := withCloneConfig(p, args)

	// git handles Ctrl+C itself. ggu has to outlive it, otherwise the shell takes the
	// terminal back while git's pager or editor is still running.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)

	git := exec.Command("git", append(profileOverrides(p), args...)...)
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

// profileOverrides passes the profile to a single git run. Settings the profile leaves empty
// are passed in their cleared form, so leftovers from another profile in the repo stay inactive.
func profileOverrides(p config.Profile) []string {
	var overrides []string
	for _, s := range profileSettings(p) {
		value := s.value
		if value == "" && s.cleared != nil {
			value = s.cleared()
		}
		if value != "" {
			overrides = append(overrides, "-c", s.key+"="+value)
		}
	}
	return overrides
}

// withCloneConfig lets git itself write the profile into a repo it clones,
// so ggu never has to work out where the clone ends up.
func withCloneConfig(p config.Profile, args []string) ([]string, bool) {
	i := subcommandIndex(args)
	if i < 0 || args[i] != "clone" {
		return args, false
	}

	var cloneConfig []string
	for _, s := range profileSettings(p) {
		if s.value != "" {
			cloneConfig = append(cloneConfig, "--config", s.key+"="+s.value)
		}
	}
	return slices.Insert(slices.Clone(args), i+1, cloneConfig...), true
}

// subcommandIndex finds git's subcommand, skipping global options like "-C <dir>".
func subcommandIndex(args []string) int {
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "-C" || arg == "-c":
			i++
		case strings.HasPrefix(arg, "-"):
		default:
			return i
		}
	}
	return -1
}

func asksForHelp(args []string) bool {
	return slices.Contains(args, "--help") || slices.Contains(args, "-h")
}

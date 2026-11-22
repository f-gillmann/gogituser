package gitconfig

import (
	"slices"
	"strings"

	"gogituser/internal/config"
)

// Overrides passes the profile to a single git run as "-c" options. Settings the profile
// leaves empty are passed in their cleared form, so leftovers in the repo stay inactive.
func Overrides(p config.Profile) []string {
	var overrides []string
	for _, s := range settings(p) {
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

// WithCloneConfig lets git itself write the profile into a repo it clones, so nobody
// has to work out where the clone ends up. It reports whether args are a clone.
func WithCloneConfig(p config.Profile, args []string) ([]string, bool) {
	i := subcommandIndex(args)
	if i < 0 || args[i] != "clone" {
		return args, false
	}

	var cloneConfig []string
	for _, s := range settings(p) {
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

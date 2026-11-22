package tui

import (
	"os"
	"path/filepath"
	"strings"

	"gogituser/config"
)

// SelectOption is a label/value pair for a select field.
type SelectOption struct {
	Label string // displayed in the list
	Value string // stored when selected
	Note  string // shown faint after the label
	key   *gpgKey
}

// SSHKeyOptions returns the private keys in ~/.ssh/ and its direct subfolders that
// have a matching .pub file. Only file names are checked, key contents are never read.
func SSHKeyOptions() []SelectOption {
	sshDir := config.ExpandHome(sshDirPrefix)
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil
	}

	opts := keyPairsIn(sshDir, "")
	for _, e := range entries {
		if e.IsDir() {
			opts = append(opts, keyPairsIn(filepath.Join(sshDir, e.Name()), e.Name()+"/")...)
		}
	}
	return opts
}

// keyPairsIn lists key pairs in dir. subPath is dir relative to ~/.ssh/, like "work/".
func keyPairsIn(dir, subPath string) []SelectOption {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var opts []SelectOption
	for _, e := range entries {
		name, isPub := strings.CutSuffix(e.Name(), ".pub")
		if e.IsDir() || !isPub {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			continue
		}
		opts = append(opts, SelectOption{
			Label: subPath + name,
			Value: sshDirPrefix + subPath + name,
		})
	}
	return opts
}

const sshDirPrefix = "~/.ssh/"

// completeKeyPath suggests entries of the folder being typed in, keeping the path as typed.
func completeKeyPath(typed string) []string {
	slash := strings.LastIndex(typed, "/")
	if slash < 0 {
		return nil
	}
	dirPart, base := typed[:slash+1], typed[slash+1:]

	entries, err := os.ReadDir(config.ExpandHome(dirPart))
	if err != nil {
		return nil
	}

	var suggestions []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, base) || strings.HasSuffix(name, ".pub") {
			continue
		}
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		if e.IsDir() {
			name += "/"
		}
		suggestions = append(suggestions, dirPart+name)
	}
	return suggestions
}

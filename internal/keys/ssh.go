package keys

import (
	"os"
	"path/filepath"
	"strings"

	"gogituser/internal/config"
)

// SSHDir is where SSH keys are looked for, written the way profiles store paths.
const SSHDir = "~/.ssh/"

type SSHKey struct {
	Name string // path relative to ~/.ssh/, like "work/id_work"
	Path string // path as stored in a profile, like "~/.ssh/work/id_work"
}

// SSHKeys returns the private keys in ~/.ssh/ and its direct subfolders that have
// a matching .pub file. Only file names are checked, key contents are never read.
func SSHKeys() []SSHKey {
	sshDir := config.ExpandHome(SSHDir)
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil
	}

	found := sshKeysIn(sshDir, "")
	for _, e := range entries {
		if e.IsDir() {
			found = append(found, sshKeysIn(filepath.Join(sshDir, e.Name()), e.Name()+"/")...)
		}
	}
	return found
}

// sshKeysIn lists key pairs in dir. subPath is dir relative to ~/.ssh/, like "work/".
func sshKeysIn(dir, subPath string) []SSHKey {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var found []SSHKey
	for _, e := range entries {
		name, isPub := strings.CutSuffix(e.Name(), ".pub")
		if e.IsDir() || !isPub {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			continue
		}
		found = append(found, SSHKey{Name: subPath + name, Path: SSHDir + subPath + name})
	}
	return found
}

// CompletePath suggests entries of the folder being typed in, keeping the path as typed.
func CompletePath(typed string) []string {
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

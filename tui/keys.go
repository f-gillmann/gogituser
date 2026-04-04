package tui

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SelectOption is a label/value pair for a select field.
type SelectOption struct {
	Label string // displayed in the list
	Value string // stored when selected
}

// GPGKeyOptions returns the fingerprints of all secret keys in the keyring.
func GPGKeyOptions() []SelectOption {
	out, err := exec.Command("gpg", "--list-secret-keys", "--with-colons").Output()
	if err != nil || len(out) == 0 {
		return nil
	}

	type gpgKey struct{ fpr, uid string }
	var keys []gpgKey
	var cur *gpgKey

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		f := strings.Split(scanner.Text(), ":")
		if len(f) < 2 {
			continue
		}
		switch f[0] {
		case "sec":
			keys = append(keys, gpgKey{})
			cur = &keys[len(keys)-1]
		case "fpr":
			if cur != nil && cur.fpr == "" && len(f) >= 10 {
				cur.fpr = f[9]
			}
		case "uid":
			if cur != nil && cur.uid == "" && len(f) >= 10 && f[1] != "r" {
				cur.uid = f[9]
			}
		}
	}

	opts := make([]SelectOption, 0, len(keys))
	for _, k := range keys {
		if k.fpr == "" {
			continue
		}
		label := k.fpr
		if len(k.fpr) > 8 {
			label = k.fpr[:8] + "..."
		}
		if k.uid != "" {
			label += "  " + k.uid
		}
		opts = append(opts, SelectOption{Label: label, Value: k.fpr})
	}
	return opts
}

// SSHKeyOptions returns paths to private key files found in ~/.ssh/.
func SSHKeyOptions() []SelectOption {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	entries, err := os.ReadDir(filepath.Join(home, ".ssh"))
	if err != nil {
		return nil
	}

	skip := map[string]bool{
		"known_hosts": true, "known_hosts.old": true,
		"config": true, "authorized_keys": true, "environment": true,
	}

	var opts []SelectOption
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".pub") || skip[e.Name()] {
			continue
		}
		opts = append(opts, SelectOption{
			Label: e.Name(),
			Value: "~/.ssh/" + e.Name(),
		})
	}
	return opts
}

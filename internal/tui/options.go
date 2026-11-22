package tui

import "gogituser/internal/keys"

// SelectOption is a label/value pair for a select field.
type SelectOption struct {
	Label string // displayed in the list
	Value string // stored when selected
	Note  string // shown faint after the label
	key   *keys.GPGKey
}

// gpgKeyOptions lists the keyring's secret keys, labeled the way gpg and GitHub show them.
func gpgKeyOptions() []SelectOption {
	var opts []SelectOption
	for _, k := range keys.ListGPG() {
		if k.Fingerprint == "" {
			continue
		}
		label := k.KeyID()
		if len(k.UIDs) > 0 {
			label += "  " + k.UIDs[0]
		}
		opts = append(opts, SelectOption{Label: label, Value: k.Fingerprint, key: &k})
	}
	return opts
}

func sshKeyOptions() []SelectOption {
	var opts []SelectOption
	for _, k := range keys.SSHKeys() {
		opts = append(opts, SelectOption{Label: k.Name, Value: k.Path})
	}
	return opts
}

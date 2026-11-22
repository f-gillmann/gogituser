package tui

import (
	"bufio"
	"bytes"
	"os/exec"
	"slices"
	"strings"
)

type gpgKey struct {
	fingerprint string
	subkeys     []string
	uids        []string
}

// listGPGKeys returns the secret keys in the keyring with their non-revoked user IDs.
func listGPGKeys() []gpgKey {
	out, err := exec.Command("gpg", "--list-secret-keys", "--with-colons").Output()
	if err != nil {
		return nil
	}
	return parseGPGKeys(out)
}

// parseGPGKeys reads gpg's --with-colons output. A "fpr" record belongs to the
// "sec" (primary key) or "ssb" (subkey) record right before it.
func parseGPGKeys(out []byte) []gpgKey {
	var keys []gpgKey
	var lastRecord string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		f := strings.Split(scanner.Text(), ":")
		if len(f) < 10 {
			continue
		}
		last := len(keys) - 1
		switch {
		case f[0] == "sec":
			keys = append(keys, gpgKey{})
		case last < 0:
		case f[0] == "fpr" && lastRecord == "sec":
			keys[last].fingerprint = f[9]
		case f[0] == "fpr" && lastRecord == "ssb":
			keys[last].subkeys = append(keys[last].subkeys, f[9])
		case f[0] == "uid" && f[1] != "r":
			keys[last].uids = append(keys[last].uids, f[9])
		}
		if f[0] != "fpr" && f[0] != "grp" {
			lastRecord = f[0]
		}
	}
	return keys
}

func (k gpgKey) keyID() string { return shortKeyID(k.fingerprint) }

func (k gpgKey) hasEmail(email string) bool {
	for _, uid := range k.uids {
		if strings.EqualFold(uidEmail(uid), strings.TrimSpace(email)) {
			return true
		}
	}
	return false
}

// uidFor prefers the user ID with the given email, so the shown identity matches the profile.
func (k gpgKey) uidFor(email string) string {
	for _, uid := range k.uids {
		if strings.EqualFold(uidEmail(uid), strings.TrimSpace(email)) {
			return uid
		}
	}
	if len(k.uids) > 0 {
		return k.uids[0]
	}
	return ""
}

func uidEmail(uid string) string {
	start, end := strings.LastIndex(uid, "<"), strings.LastIndex(uid, ">")
	if start < 0 || end < start {
		return ""
	}
	return uid[start+1 : end]
}

// minKeyIDLength keeps very short inputs from matching the tail of some random fingerprint.
const minKeyIDLength = 8

// findGPGKey matches a signing key the way people write it: a fingerprint, a long or
// short key ID, with or without "0x", for the primary key or one of its subkeys.
func findGPGKey(keys []gpgKey, id string) (gpgKey, bool) {
	id = normalizeKeyID(id)
	if len(id) < minKeyIDLength {
		return gpgKey{}, false
	}
	for _, k := range keys {
		if k.fingerprint == "" {
			continue
		}
		if strings.HasSuffix(k.fingerprint, id) || slices.ContainsFunc(k.subkeys, func(sub string) bool {
			return strings.HasSuffix(sub, id)
		}) {
			return k, true
		}
	}
	return gpgKey{}, false
}

func normalizeKeyID(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	id = strings.TrimPrefix(id, "0X")
	return strings.TrimSuffix(id, "!")
}

// ResolveSigningKey turns any way of naming a key in the keyring into its full primary
// fingerprint, so profiles store one stable value. Keys pinned with "!" stay as written,
// since that tells gpg to use exactly that subkey.
func ResolveSigningKey(id string) string {
	if strings.HasSuffix(strings.TrimSpace(id), "!") {
		return id
	}
	if key, ok := findGPGKey(listGPGKeys(), id); ok {
		return key.fingerprint
	}
	return id
}

// GPGKeyOptions returns the keyring's secret keys, labeled the way gpg and GitHub show them.
func GPGKeyOptions() []SelectOption {
	var opts []SelectOption
	for _, k := range listGPGKeys() {
		if k.fingerprint == "" {
			continue
		}
		label := k.keyID()
		if len(k.uids) > 0 {
			label += "  " + k.uids[0]
		}
		opts = append(opts, SelectOption{Label: label, Value: k.fingerprint, key: &k})
	}
	return opts
}

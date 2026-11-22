// Package keys finds GPG and SSH keys on this machine. It only reads key
// metadata and file names, never private key material.
package keys

import (
	"bufio"
	"bytes"
	"os/exec"
	"slices"
	"strings"
)

// keyIDLength is the long key ID that gpg and GitHub show, the tail of the fingerprint.
const keyIDLength = 16

// minKeyIDLength keeps very short inputs from matching the tail of some random fingerprint.
const minKeyIDLength = 8

type GPGKey struct {
	Fingerprint string
	Subkeys     []string
	UIDs        []string
}

// ListGPG returns the secret keys in the keyring with their non-revoked user IDs.
func ListGPG() []GPGKey {
	out, err := exec.Command("gpg", "--list-secret-keys", "--with-colons").Output()
	if err != nil {
		return nil
	}
	return parseGPG(out)
}

// parseGPG reads gpg's --with-colons output. A "fpr" record belongs to the
// "sec" (primary key) or "ssb" (subkey) record right before it.
func parseGPG(out []byte) []GPGKey {
	var keys []GPGKey
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
			keys = append(keys, GPGKey{})
		case last < 0:
		case f[0] == "fpr" && lastRecord == "sec":
			keys[last].Fingerprint = f[9]
		case f[0] == "fpr" && lastRecord == "ssb":
			keys[last].Subkeys = append(keys[last].Subkeys, f[9])
		case f[0] == "uid" && f[1] != "r":
			keys[last].UIDs = append(keys[last].UIDs, f[9])
		}
		if f[0] != "fpr" && f[0] != "grp" {
			lastRecord = f[0]
		}
	}
	return keys
}

func (k GPGKey) KeyID() string { return ShortKeyID(k.Fingerprint) }

func (k GPGKey) HasEmail(email string) bool {
	return k.uidWithEmail(email) != ""
}

// UIDFor prefers the user ID with the given email, so the shown identity matches the profile.
func (k GPGKey) UIDFor(email string) string {
	if uid := k.uidWithEmail(email); uid != "" {
		return uid
	}
	if len(k.UIDs) > 0 {
		return k.UIDs[0]
	}
	return ""
}

func (k GPGKey) uidWithEmail(email string) string {
	for _, uid := range k.UIDs {
		if _, uidEmail := SplitUID(uid); strings.EqualFold(uidEmail, strings.TrimSpace(email)) {
			return uid
		}
	}
	return ""
}

// SplitUID splits a user ID like "Name (comment) <mail@example.com>" into name and email.
func SplitUID(uid string) (name, email string) {
	start, end := strings.LastIndex(uid, "<"), strings.LastIndex(uid, ">")
	if start < 0 || end < start {
		return strings.TrimSpace(uid), ""
	}
	name = strings.TrimSpace(uid[:start])
	if comment := strings.Index(name, " ("); comment >= 0 {
		name = name[:comment]
	}
	return name, uid[start+1 : end]
}

// FindGPG matches a signing key the way people write it: a fingerprint, a long or
// short key ID, with or without "0x", for the primary key or one of its subkeys.
func FindGPG(keys []GPGKey, id string) (GPGKey, bool) {
	id = normalizeKeyID(id)
	if len(id) < minKeyIDLength {
		return GPGKey{}, false
	}
	for _, k := range keys {
		if k.Fingerprint == "" {
			continue
		}
		if strings.HasSuffix(k.Fingerprint, id) || slices.ContainsFunc(k.Subkeys, func(sub string) bool {
			return strings.HasSuffix(sub, id)
		}) {
			return k, true
		}
	}
	return GPGKey{}, false
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
	if key, ok := FindGPG(ListGPG(), id); ok {
		return key.Fingerprint
	}
	return id
}

func ShortKeyID(fingerprint string) string {
	if len(fingerprint) <= keyIDLength {
		return fingerprint
	}
	return fingerprint[len(fingerprint)-keyIDLength:]
}

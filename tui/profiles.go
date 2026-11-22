package tui

import (
	"slices"
	"strings"

	"gogituser/config"
)

// keyIDLength is the long key ID that gpg and GitHub show, the tail of the fingerprint.
const keyIDLength = 16

var detailLabel = dimStyle.Width(9)

// ProfileList renders all profiles and marks the one active in the current repo.
func ProfileList(profiles []config.Profile, active string) string {
	keys := listGPGKeys()
	blocks := make([]string, len(profiles))
	for i, p := range profiles {
		title := accentStyle.Render(p.Name)
		if p.Name == active {
			title += "  " + checkStyle.Render("✓ active")
		}
		blocks[i] = title + "\n" + profileDetails(p, keys)
	}
	return strings.Join(blocks, "\n\n")
}

// ProfileDetails renders the identity lines of a profile, indented under its name.
func ProfileDetails(p config.Profile) string {
	return profileDetails(p, listGPGKeys())
}

func profileDetails(p config.Profile, keys []gpgKey) string {
	email := p.GitEmail
	var signing string
	if p.SigningKey != "" {
		key, found := findGPGKey(keys, p.SigningKey)
		signing = signingValue(p, key, found)
		if found && !key.hasEmail(p.GitEmail) {
			email += "  " + warnStyle.Render("differs from signing key")
		}
	}

	lines := []string{
		detailRow("Name", p.GitName),
		detailRow("Email", email),
		detailRow("Signing", signing),
		detailRow("SSH", p.SSHKey),
	}
	return strings.Join(slices.DeleteFunc(lines, func(l string) bool { return l == "" }), "\n")
}

func detailRow(label, value string) string {
	if value == "" {
		return ""
	}
	return "  " + detailLabel.Render(label) + value
}

func signingValue(p config.Profile, key gpgKey, found bool) string {
	if !found {
		return shortKeyID(p.SigningKey) + "  " + warnStyle.Render("not found in keyring")
	}
	value := key.keyID()
	if uid := key.uidFor(p.GitEmail); uid != "" {
		value += "  " + dimStyle.Render(uid)
	}
	return value
}

func UsingProfile(name string) string {
	return checkStyle.Render("✓") + " Using profile " + accentStyle.Render(name)
}

func shortKeyID(fingerprint string) string {
	if len(fingerprint) <= keyIDLength {
		return fingerprint
	}
	return fingerprint[len(fingerprint)-keyIDLength:]
}

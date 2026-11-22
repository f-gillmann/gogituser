package tui

import (
	"slices"
	"strings"

	"github.com/f-gillmann/gogituser/internal/config"
	"github.com/f-gillmann/gogituser/internal/keys"
)

var detailLabel = dimStyle.Width(9)

// ProfileList renders all profiles and marks the one active in the current repo.
func ProfileList(profiles []config.Profile, active string) string {
	keyring := keys.ListGPG()
	blocks := make([]string, len(profiles))
	for i, p := range profiles {
		title := accentStyle.Render(p.Name)
		if p.Name == active {
			title += "  " + checkStyle.Render("✓ active")
		}
		blocks[i] = title + "\n" + profileDetails(p, keyring)
	}
	return strings.Join(blocks, "\n\n")
}

// ProfileDetails renders the identity lines of a profile, indented under its name.
func ProfileDetails(p config.Profile) string {
	return profileDetails(p, keys.ListGPG())
}

func profileDetails(p config.Profile, keyring []keys.GPGKey) string {
	email := p.GitEmail
	var signing string
	if p.SigningKey != "" {
		key, found := keys.FindGPG(keyring, p.SigningKey)
		signing = signingValue(p, key, found)
		if found && !key.HasEmail(p.GitEmail) {
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

func signingValue(p config.Profile, key keys.GPGKey, found bool) string {
	if !found {
		return keys.ShortKeyID(p.SigningKey) + "  " + warnStyle.Render("not found in keyring")
	}
	value := key.KeyID()
	if uid := key.UIDFor(p.GitEmail); uid != "" {
		value += "  " + dimStyle.Render(uid)
	}
	return value
}

func UsingProfile(name string) string {
	return checkStyle.Render("✓") + " Using profile " + accentStyle.Render(name)
}

// Hint renders a short note about something that works differently than expected.
func Hint(text string) string {
	return hintStyle.Render(text)
}

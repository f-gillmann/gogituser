package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ANSI palette indices, so every color follows the user's terminal theme.
const (
	ansiRed     = lipgloss.Color("1")
	ansiGreen   = lipgloss.Color("2")
	ansiBlue    = lipgloss.Color("4")
	ansiMagenta = lipgloss.Color("5")
	ansiCyan    = lipgloss.Color("6")
)

var (
	accentStyle   = lipgloss.NewStyle().Bold(true).Foreground(ansiMagenta)
	errorStyle    = lipgloss.NewStyle().Bold(true).Foreground(ansiRed)
	checkStyle    = lipgloss.NewStyle().Foreground(ansiGreen)
	warnStyle     = lipgloss.NewStyle().Foreground(ansiRed)
	normalStyle   = lipgloss.NewStyle()
	dimStyle      = lipgloss.NewStyle().Faint(true)
	dropdownArrow = lipgloss.NewStyle().Foreground(ansiCyan)
	keyStyle      = lipgloss.NewStyle().Foreground(ansiCyan)
	customStyle   = lipgloss.NewStyle().Foreground(ansiBlue).Italic(true)
)

type keyHint struct {
	key, action string
}

func renderHints(hints ...keyHint) string {
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = keyStyle.Render(h.key) + " " + dimStyle.Render(h.action)
	}
	return strings.Join(parts, "   ")
}

func newTextInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.PlaceholderStyle = dimStyle
	ti.Prompt = "  "
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(ansiMagenta)
	return ti
}

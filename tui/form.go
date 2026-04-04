package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gogituser/config"
)

// formField is the common interface for all fields in the form.
// update returns (cmd, advance) where advance=true means move to next field.
type formField interface {
	update(msg tea.Msg) (tea.Cmd, bool)
	view(focused bool) string
	value() string
	setValue(string)
	focus() tea.Cmd
	blur()
	isExpanded(focused bool) bool
}

// textField wraps a bubbles textinput.
type textField struct{ input textinput.Model }

func newTextField(placeholder, initial string) *textField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(initial)
	return &textField{input: ti}
}

func (f *textField) update(msg tea.Msg) (tea.Cmd, bool) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "enter" {
		return nil, true
	}
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return cmd, false
}
func (f *textField) view(bool) string     { return f.input.View() }
func (f *textField) value() string        { return f.input.Value() }
func (f *textField) setValue(v string)    { f.input.SetValue(v) }
func (f *textField) focus() tea.Cmd       { return f.input.Focus() }
func (f *textField) blur()                { f.input.Blur() }
func (f *textField) isExpanded(bool) bool { return false }

// ---- field indices -------------------------------------------------------

const (
	fieldName = iota
	fieldGitName
	fieldEmail
	fieldSigningKey
	fieldSSHKey
	fieldSSHHost
	fieldCount
)

var fieldLabels = [fieldCount]string{
	"Profile name",
	"Git name",
	"Email",
	"Signing key",
	"SSH key",
	"SSH host",
}

// ---- form model ----------------------------------------------------------

type formModel struct {
	fields    [fieldCount]formField
	focus     int
	submitted bool
	cancelled bool
	nameFixed bool
}

func newFormModel(initial config.Profile, nameFixed bool) formModel {
	m := formModel{nameFixed: nameFixed}

	m.fields[fieldName] = newTextField("profile name", initial.Name)
	m.fields[fieldGitName] = newTextField("full name", initial.GitName)
	m.fields[fieldEmail] = newTextField("email address", initial.Email)
	m.fields[fieldSigningKey] = newSelectField(GPGKeyOptions(), initial.SigningKey)
	m.fields[fieldSSHKey] = newSelectField(SSHKeyOptions(), initial.SSHKey)
	m.fields[fieldSSHHost] = newTextField("e.g. github-personal (optional)", initial.SSHHost)

	start := fieldName
	if nameFixed {
		start = fieldGitName
	}
	m.fields[start].focus()
	m.focus = start
	return m
}

func (m formModel) Init() tea.Cmd { return textinput.Blink }

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "shift+tab":
			start := fieldName
			if m.nameFixed {
				start = fieldGitName
			}
			if m.focus > start {
				m.fields[m.focus].blur()
				m.focus--
				return m, m.fields[m.focus].focus()
			}
			return m, nil

		case "tab":
			return m.advanceFocus()
		}
	}

	// Delegate to the current field.
	cmd, advance := m.fields[m.focus].update(msg)
	if advance {
		next, nextCmd := m.advanceFocus()
		return next, nextCmd
	}
	return m, cmd
}

func (m formModel) advanceFocus() (formModel, tea.Cmd) {
	next := m.focus + 1
	if next >= fieldCount {
		// Require a non-empty name before submitting.
		if m.fields[fieldName].value() == "" {
			return m, nil
		}
		m.submitted = true
		return m, tea.Quit
	}
	m.fields[m.focus].blur()
	m.focus = next
	return m, m.fields[m.focus].focus()
}

// ---- styles --------------------------------------------------------------

var (
	labelStyle  = lipgloss.NewStyle().Width(14).Foreground(lipgloss.Color("240"))
	activeLabel = lipgloss.NewStyle().Width(14).Bold(true)
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	indentStyle = lipgloss.NewStyle().PaddingLeft(2)
)

func (m formModel) View() string {
	var b strings.Builder
	b.WriteString("\n")

	for i, field := range m.fields {
		focused := i == m.focus
		label := fieldLabels[i]

		// Name field: show "(fixed)" when not editable.
		if i == fieldName && m.nameFixed {
			b.WriteString(labelStyle.Render(label+": ") + field.value() + " " + hintStyle.Render("(fixed)") + "\n")
			continue
		}

		if field.isExpanded(focused) {
			// Multi-line: label on its own line, list indented below.
			b.WriteString(activeLabel.Render(label+":") + "\n")
			b.WriteString(indentStyle.Render(field.view(focused)) + "\n")
		} else {
			lbl := labelStyle.Render(label + ": ")
			if focused {
				lbl = activeLabel.Render(label + ": ")
			}
			b.WriteString(lbl + field.view(focused) + "\n")
		}
	}

	b.WriteString("\n" + hintStyle.Render("Tab: next • Shift+Tab: previous • Enter: select/next • Ctrl+C: cancel") + "\n")
	return b.String()
}

func (m formModel) toProfile() config.Profile {
	return config.Profile{
		Name:       m.fields[fieldName].value(),
		GitName:    m.fields[fieldGitName].value(),
		Email:      m.fields[fieldEmail].value(),
		SigningKey: m.fields[fieldSigningKey].value(),
		SSHKey:     m.fields[fieldSSHKey].value(),
		SSHHost:    m.fields[fieldSSHHost].value(),
	}
}

// RunForm shows an input form pre-filled with initial. nameFixed=true locks
// the name field (used for update). Returns profile + submitted bool.
func RunForm(initial config.Profile, nameFixed bool) (config.Profile, bool, error) {
	m, err := tea.NewProgram(newFormModel(initial, nameFixed)).Run()
	if err != nil {
		return config.Profile{}, false, err
	}
	fm := m.(formModel)
	if fm.cancelled || !fm.submitted {
		return config.Profile{}, false, nil
	}
	return fm.toProfile(), true, nil
}

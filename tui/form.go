package tui

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gogituser/config"
)

// navigation tells the form where focus should go after a field handled a key.
type navigation int

const (
	navStay navigation = iota
	navPrev
	navNext
)

type formField interface {
	update(msg tea.Msg) (tea.Cmd, navigation)
	view(focused bool) string
	value() string
	setValue(string)
	focus() tea.Cmd
	blur()
	// keyHints replaces the form's default hints while the field needs its own keys, nil otherwise.
	keyHints() []keyHint
	capturesTab() bool
}

// textField wraps a bubbles textinput.
type textField struct{ input textinput.Model }

func newTextField(placeholder, initial string) *textField {
	ti := newTextInput(placeholder)
	ti.SetValue(initial)
	return &textField{input: ti}
}

func (f *textField) update(msg tea.Msg) (tea.Cmd, navigation) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "up":
			return nil, navPrev
		case "down":
			return nil, navNext
		}
	}
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return cmd, navStay
}
func (f *textField) view(bool) string    { return f.input.View() }
func (f *textField) value() string       { return f.input.Value() }
func (f *textField) setValue(v string)   { f.input.SetValue(v) }
func (f *textField) focus() tea.Cmd      { return f.input.Focus() }
func (f *textField) blur()               { f.input.Blur() }
func (f *textField) keyHints() []keyHint { return nil }
func (f *textField) capturesTab() bool   { return false }

// ---- field indices -------------------------------------------------------

const (
	fieldName = iota
	fieldGitName
	fieldGitEmail
	fieldSigningKey
	fieldSSHKey
	fieldCount
)

// saveRow is the focus index of the Save row below the last field.
const saveRow = fieldCount

var fieldLabels = [fieldCount]string{
	"Profile Name",
	"Git Name",
	"Git Email",
	"Signing Key",
	"SSH Key",
}

var requiredFields = []int{fieldName, fieldGitName, fieldGitEmail}

// ---- form model ----------------------------------------------------------

type formModel struct {
	fields      [fieldCount]formField
	focus       int
	submitted   bool
	cancelled   bool
	nameFixed   bool
	showMissing bool
}

func newFormModel(initial config.Profile, nameFixed bool) formModel {
	m := formModel{nameFixed: nameFixed}

	m.fields[fieldName] = newTextField("profile name", initial.Name)
	m.fields[fieldGitName] = newTextField("your name", initial.GitName)
	m.fields[fieldGitEmail] = newTextField("email address", initial.GitEmail)
	m.fields[fieldSigningKey] = newSelectField(GPGKeyOptions(), ResolveSigningKey(initial.SigningKey), nil)
	m.fields[fieldSSHKey] = newSelectField(SSHKeyOptions(), initial.SSHKey, &customEntry{
		label:       "Other path",
		placeholder: "path to key file",
		prefill:     sshDirPrefix,
		complete:    completeKeyPath,
		validate:    config.CheckKeyPath,
	})

	m.focus = m.firstEditable()
	m.fields[m.focus].focus()
	return m
}

func (m formModel) Init() tea.Cmd { return textinput.Blink }

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "ctrl+s":
			return m.submit()
		case "shift+tab":
			return m.focusPrev()
		case "tab":
			if m.focus == saveRow || !m.fields[m.focus].capturesTab() {
				return m.focusNext()
			}
		}
	}

	if m.focus == saveRow {
		return m.updateSaveRow(msg)
	}

	cmd, nav := m.fields[m.focus].update(msg)
	switch nav {
	case navPrev:
		return m.focusPrev()
	case navNext:
		return m.focusNext()
	}
	return m, cmd
}

func (m formModel) updateSaveRow(msg tea.Msg) (formModel, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k.String() {
	case "enter":
		return m.submit()
	case "up", "k":
		return m.focusPrev()
	}
	return m, nil
}

func (m formModel) submit() (formModel, tea.Cmd) {
	if missing := m.firstMissing(); missing >= 0 {
		m.showMissing = true
		return m.focusField(missing)
	}
	m.submitted = true
	return m, tea.Quit
}

func (m formModel) firstEditable() int {
	if m.nameFixed {
		return fieldGitName
	}
	return fieldName
}

func (m formModel) focusPrev() (formModel, tea.Cmd) {
	if m.focus <= m.firstEditable() {
		return m, nil
	}
	return m.focusField(m.focus - 1)
}

func (m formModel) focusNext() (formModel, tea.Cmd) {
	if m.focus == saveRow {
		return m, nil
	}
	return m.focusField(m.focus + 1)
}

func (m formModel) focusField(i int) (formModel, tea.Cmd) {
	if m.focus != saveRow {
		m.fields[m.focus].blur()
	}
	m.focus = i
	if m.focus == saveRow {
		return m, nil
	}
	return m, m.fields[m.focus].focus()
}

func (m formModel) firstMissing() int {
	for _, i := range requiredFields {
		if m.isMissing(i) {
			return i
		}
	}
	return -1
}

func (m formModel) isMissing(i int) bool {
	return slices.Contains(requiredFields, i) && strings.TrimSpace(m.fields[i].value()) == ""
}

// ---- styles --------------------------------------------------------------

var (
	rowStyle     = lipgloss.NewStyle().PaddingLeft(2)
	focusedRow   = lipgloss.NewStyle().PaddingLeft(1).BorderLeft(true).BorderStyle(lipgloss.ThickBorder()).BorderForeground(ansiMagenta)
	labelStyle   = normalStyle.Width(14)
	activeLabel  = accentStyle.Width(14)
	missingLabel = errorStyle.Width(14)
)

func (m formModel) View() string {
	if signing, ok := m.fields[fieldSigningKey].(*selectField); ok {
		signing.matchEmail = m.fields[fieldGitEmail].value()
	}

	var b strings.Builder
	b.WriteString("\n")

	for i, field := range m.fields {
		focused := i == m.focus
		label := fieldLabels[i]

		// Name field: show "(fixed)" when not editable.
		if i == fieldName && m.nameFixed {
			b.WriteString(rowStyle.Render(labelStyle.Render(label+": ")+"  "+field.value()+" "+dimStyle.Render("(fixed)")) + "\n")
			continue
		}

		lbl := labelStyle.Render(label + ": ")
		switch {
		case m.showMissing && m.isMissing(i):
			lbl = missingLabel.Render(label + ": ")
		case focused:
			lbl = activeLabel.Render(label + ": ")
		}
		b.WriteString(renderRow(lipgloss.JoinHorizontal(lipgloss.Top, lbl, field.view(focused)), focused) + "\n")
	}

	saveFocused := m.focus == saveRow
	saveButton := normalStyle.Render("Save")
	if saveFocused {
		saveButton = accentStyle.Render("Save")
	}
	b.WriteString("\n" + renderRow(saveButton, saveFocused) + "\n")

	b.WriteString("\n" + rowStyle.Render(renderHints(m.keyHints()...)) + "\n")
	return b.String()
}

// renderRow adds a left gutter that turns into a colored bar on the focused row.
func renderRow(content string, focused bool) string {
	if focused {
		return focusedRow.Render(content)
	}
	return rowStyle.Render(content)
}

func (m formModel) keyHints() []keyHint {
	if m.focus == saveRow {
		return []keyHint{{"↑", "move"}, {"Enter", "save"}, {"Ctrl+C", "cancel"}}
	}
	if hints := m.fields[m.focus].keyHints(); hints != nil {
		return hints
	}
	return withFormHints(keyHint{"↑↓", "move"})
}

func withFormHints(hints ...keyHint) []keyHint {
	return append(hints, keyHint{"Ctrl+S", "save"}, keyHint{"Ctrl+C", "cancel"})
}

func (m formModel) toProfile() config.Profile {
	return config.Profile{
		Name:       m.fields[fieldName].value(),
		GitName:    m.fields[fieldGitName].value(),
		GitEmail:   m.fields[fieldGitEmail].value(),
		SigningKey: m.fields[fieldSigningKey].value(),
		SSHKey:     m.fields[fieldSSHKey].value(),
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

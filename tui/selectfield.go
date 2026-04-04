package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type selectMode int

const (
	selectModePick selectMode = iota
	selectModeType
)

const customLabel = "Custom..."

var (
	pickCursor = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	pickNormal = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	pickDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	pickCustom = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Italic(true)
)

// selectField is a form field with a predefined list of options plus a
// "Custom..." entry that falls back to free-text input.
type selectField struct {
	options []SelectOption
	cursor  int
	mode    selectMode
	input   textinput.Model
}

func newSelectField(options []SelectOption, initial string) *selectField {
	ti := textinput.New()
	ti.Placeholder = "enter value"

	cursor := len(options) // default: Custom
	for i, o := range options {
		if o.Value == initial {
			cursor = i
			break
		}
	}
	if cursor == len(options) && initial != "" {
		ti.SetValue(initial)
	}

	return &selectField{options: options, cursor: cursor, input: ti}
}

func (f *selectField) update(msg tea.Msg) (tea.Cmd, bool) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		if f.mode == selectModeType {
			var cmd tea.Cmd
			f.input, cmd = f.input.Update(msg)
			return cmd, false
		}
		return nil, false
	}

	if f.mode == selectModeType {
		switch key.String() {
		case "enter":
			return nil, true
		case "esc":
			f.mode = selectModePick
			f.input.Blur()
			return nil, false
		default:
			var cmd tea.Cmd
			f.input, cmd = f.input.Update(msg)
			return cmd, false
		}
	}

	// pick mode
	total := len(f.options) + 1
	switch key.String() {
	case "up", "k":
		if f.cursor > 0 {
			f.cursor--
		}
	case "down", "j":
		if f.cursor < total-1 {
			f.cursor++
		}
	case "enter":
		if f.cursor == len(f.options) {
			f.mode = selectModeType
			return f.input.Focus(), false
		}
		return nil, true
	}
	return nil, false
}

func (f *selectField) value() string {
	if f.mode == selectModeType {
		return f.input.Value()
	}
	if f.cursor < len(f.options) {
		return f.options[f.cursor].Value
	}
	return f.input.Value() // custom, not yet entered type mode
}

func (f *selectField) setValue(v string) {
	for i, o := range f.options {
		if o.Value == v {
			f.cursor = i
			f.mode = selectModePick
			return
		}
	}
	f.cursor = len(f.options)
	f.input.SetValue(v)
}

func (f *selectField) focus() tea.Cmd { return nil }
func (f *selectField) blur()          { f.input.Blur(); f.mode = selectModePick }

// isExpanded returns true when the field should render as a multi-line list.
func (f *selectField) isExpanded(focused bool) bool {
	return focused && f.mode == selectModePick
}

func (f *selectField) view(focused bool) string {
	if !focused {
		v := f.value()
		if v == "" {
			return pickDim.Render("(not set)")
		}
		return v
	}
	if f.mode == selectModeType {
		return f.input.View() + "  " + pickDim.Render("Esc: back to list")
	}

	s := ""
	for i, o := range f.options {
		if i == f.cursor {
			s += pickCursor.Render("▶ "+o.Label) + "\n"
		} else {
			s += pickNormal.Render("  "+o.Label) + "\n"
		}
	}
	customIdx := len(f.options)
	if customIdx == f.cursor {
		s += pickCursor.Render("▶ " + customLabel)
	} else {
		s += pickCustom.Render("  " + customLabel)
	}
	return s
}

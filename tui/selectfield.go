package tui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type selectMode int

const (
	selectModeClosed selectMode = iota
	selectModeOpen
	selectModeType
)

const (
	noneLabel         = "None"
	maxVisibleOptions = 5
)

// customEntry configures the extra row that opens a free-text input.
// complete and validate are optional. Without a customEntry there is no such row.
type customEntry struct {
	label       string
	placeholder string
	prefill     string
	complete    func(typed string) []string
	validate    func(value string) error
}

// selectField is a form field with a "None" entry, the found options, and a
// custom row that falls back to free-text input. Enter opens the list,
// arrows move the cursor, and Enter picks the row under it.
type selectField struct {
	options  []SelectOption
	cursor   int
	selected int
	offset   int
	mode     selectMode
	input    textinput.Model
	custom   *customEntry
	err      error
	// matchEmail flags GPG keys that do not carry this email, since GitHub would show their commits as unverified.
	matchEmail string
}

func newSelectField(found []SelectOption, initial string, custom *customEntry) *selectField {
	input := newTextInput("")
	if custom != nil {
		input.Placeholder = custom.placeholder
		input.ShowSuggestions = custom.complete != nil
		input.CompletionStyle = dimStyle
	}

	f := &selectField{
		options: append([]SelectOption{{Label: noneLabel}}, found...),
		input:   input,
		custom:  custom,
	}
	f.setValue(initial)
	return f
}

func (f *selectField) rowCount() int {
	if f.custom == nil {
		return len(f.options)
	}
	return len(f.options) + 1
}

// customRow is the index of the custom row, or -1 when the field has none.
func (f *selectField) customRow() int {
	if f.custom == nil {
		return -1
	}
	return len(f.options)
}

func (f *selectField) scrollToCursor() {
	if f.cursor < f.offset {
		f.offset = f.cursor
	}
	if f.cursor >= f.offset+maxVisibleOptions {
		f.offset = f.cursor - maxVisibleOptions + 1
	}
}

func (f *selectField) update(msg tea.Msg) (tea.Cmd, navigation) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		if f.mode == selectModeType {
			var cmd tea.Cmd
			f.input, cmd = f.input.Update(msg)
			return cmd, navStay
		}
		return nil, navStay
	}

	switch f.mode {
	case selectModeOpen:
		return f.updateOpen(key)
	case selectModeType:
		return f.updateTyping(key)
	}
	return f.updateClosed(key)
}

func (f *selectField) updateClosed(key tea.KeyMsg) (tea.Cmd, navigation) {
	switch key.String() {
	case "enter", "right":
		f.sortMatchingFirst()
		f.mode = selectModeOpen
		f.cursor = f.selected
		f.scrollToCursor()
	case "up":
		return nil, navPrev
	case "down":
		return nil, navNext
	}
	return nil, navStay
}

func (f *selectField) updateTyping(key tea.KeyMsg) (tea.Cmd, navigation) {
	switch key.String() {
	case "enter":
		if f.err = f.checkTyped(); f.err != nil {
			return nil, navStay
		}
		f.selected = f.customRow()
		f.input.Blur()
		f.mode = selectModeClosed
		return nil, navStay
	case "esc":
		f.err = nil
		f.mode = selectModeOpen
		f.input.Blur()
		return nil, navStay
	}

	f.err = nil
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(key)
	f.refreshSuggestions()
	return cmd, navStay
}

func (f *selectField) startTyping() tea.Cmd {
	f.mode = selectModeType
	if f.input.Value() == "" && f.custom.prefill != "" {
		f.input.SetValue(f.custom.prefill)
		f.input.CursorEnd()
	}
	f.refreshSuggestions()
	return f.input.Focus()
}

func (f *selectField) refreshSuggestions() {
	if f.custom.complete != nil {
		f.input.SetSuggestions(f.custom.complete(f.input.Value()))
	}
}

func (f *selectField) checkTyped() error {
	if f.custom.validate == nil {
		return nil
	}
	return f.custom.validate(f.input.Value())
}

// capturesTab lets the custom input use Tab for completion instead of moving focus.
func (f *selectField) capturesTab() bool {
	return f.mode == selectModeType && f.custom.complete != nil
}

// updateOpen moves the cursor without changing the value until Enter picks a row.
func (f *selectField) updateOpen(key tea.KeyMsg) (tea.Cmd, navigation) {
	switch key.String() {
	case "up", "k":
		f.cursor = max(f.cursor-1, 0)
		f.scrollToCursor()
	case "down", "j":
		f.cursor = min(f.cursor+1, f.rowCount()-1)
		f.scrollToCursor()
	case "esc", "left":
		f.mode = selectModeClosed
	case "enter":
		if f.cursor == f.customRow() {
			return f.startTyping(), navStay
		}
		f.selected = f.cursor
		f.mode = selectModeClosed
	}
	return nil, navStay
}

func (f *selectField) value() string {
	if f.selected == f.customRow() || f.hasValidTyped() {
		return f.input.Value()
	}
	return f.options[f.selected].Value
}

func (f *selectField) hasValidTyped() bool {
	return f.mode == selectModeType && f.input.Value() != f.custom.prefill && f.checkTyped() == nil
}

// setValue picks the matching option. An unknown value goes into the custom input,
// or becomes its own "(not found)" row when the field has no custom input.
func (f *selectField) setValue(v string) {
	f.selected = f.indexOf(v)
	if f.selected < 0 {
		if f.custom != nil {
			f.selected = f.customRow()
			f.input.SetValue(v)
		} else {
			f.options = append(f.options, SelectOption{Label: v, Value: v, Note: "(not found)"})
			f.selected = len(f.options) - 1
		}
	}
	f.cursor = f.selected
	f.scrollToCursor()
}

// sortMatchingFirst moves keys carrying matchEmail right below "None", keeping the current pick.
func (f *selectField) sortMatchingFirst() {
	if f.matchEmail == "" || f.selected == f.customRow() {
		return
	}
	picked := f.options[f.selected].Value
	slices.SortStableFunc(f.options[1:], func(a, b SelectOption) int {
		return cmp.Compare(f.matchRank(a), f.matchRank(b))
	})
	f.selected = f.indexOf(picked)
}

func (f *selectField) matchRank(o SelectOption) int {
	if o.key != nil && o.key.hasEmail(f.matchEmail) {
		return 0
	}
	return 1
}

func (f *selectField) indexOf(v string) int {
	for i, o := range f.options {
		if o.Value == v {
			return i
		}
	}
	return -1
}

func (f *selectField) focus() tea.Cmd { return nil }

func (f *selectField) blur() {
	if f.hasValidTyped() {
		f.selected = f.customRow()
	}
	f.err = nil
	f.input.Blur()
	f.mode = selectModeClosed
}

func (f *selectField) keyHints() []keyHint {
	switch f.mode {
	case selectModeOpen:
		return []keyHint{{"↑↓", "move"}, {"Enter", "pick"}, {"Esc/←", "close"}}
	case selectModeType:
		if f.custom.complete != nil {
			return []keyHint{{"Tab", "complete"}, {"Enter", "pick"}, {"Esc", "back to list"}}
		}
		return []keyHint{{"Enter", "pick"}, {"Esc", "back to list"}}
	}
	return withFormHints(keyHint{"↑↓", "move"}, keyHint{"Enter/→", "open"})
}

func (f *selectField) view(focused bool) string {
	if focused && f.mode == selectModeType {
		if f.err != nil {
			return f.input.View() + "  " + errorStyle.Render(f.err.Error())
		}
		return f.input.View()
	}

	open := focused && f.mode == selectModeOpen
	arrow := "›"
	if open {
		arrow = "⌄"
	}
	header := "  " + f.displayValue() + " " + dropdownArrow.Render(arrow)
	if open {
		return header + "\n" + f.listView()
	}
	return header
}

// displayValue shows the picked option's label, which is easier to read than its stored value.
func (f *selectField) displayValue() string {
	v := f.value()
	switch {
	case v == "":
		return dimStyle.Render("(optional)")
	case f.selected < len(f.options):
		return f.options[f.selected].Label + f.noteFor(f.options[f.selected])
	}
	return v
}

func (f *selectField) noteFor(o SelectOption) string {
	note := o.Note
	if note == "" && o.key != nil && f.matchEmail != "" && !o.key.hasEmail(f.matchEmail) {
		note = "(other email)"
	}
	if note == "" {
		return ""
	}
	return "  " + dimStyle.Render(note)
}

// listView is indented under the header so it reads as the content of the opened field.
func (f *selectField) listView() string {
	const indent = "  "
	var lines []string
	end := min(f.offset+maxVisibleOptions, f.rowCount())
	for i := f.offset; i < end; i++ {
		lines = append(lines, indent+f.renderRow(i))
	}
	return strings.Join(lines, "\n")
}

func (f *selectField) renderRow(i int) string {
	var label, note string
	if i == f.customRow() {
		label = f.custom.label
	} else {
		label, note = f.options[i].Label, f.noteFor(f.options[i])
	}

	var row string
	switch {
	case i == f.cursor:
		row = accentStyle.Render("▶ " + label)
	case i == f.customRow():
		row = customStyle.Render("  " + label)
	default:
		row = normalStyle.Render("  " + label)
	}
	row += note
	if i == f.selected {
		row += checkStyle.Render(" ✓")
	}
	if i == f.cursor {
		row += f.position()
	}
	return row
}

// position is only shown when the list scrolls.
func (f *selectField) position() string {
	if f.rowCount() <= maxVisibleOptions {
		return ""
	}
	return dimStyle.Render(fmt.Sprintf("  (%d/%d)", f.cursor+1, f.rowCount()))
}

package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gogituser/config"
)

var (
	keyUp    = tea.KeyMsg{Type: tea.KeyUp}
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
	keyLeft  = tea.KeyMsg{Type: tea.KeyLeft}
	keyRight = tea.KeyMsg{Type: tea.KeyRight}
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
	keyCtrlS = tea.KeyMsg{Type: tea.KeyCtrlS}
)

func runes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func pressAll(f *selectField, keys ...tea.KeyMsg) {
	for _, k := range keys {
		f.update(k)
	}
}

var twoKeys = []SelectOption{{Label: "id_a", Value: "~/.ssh/id_a"}, {Label: "id_b", Value: "~/.ssh/id_b"}}

func TestSelectFieldPickingNeedsEnter(t *testing.T) {
	f := newSelectField(twoKeys, "", nil)

	pressAll(f, keyRight, keyDown, keyDown)
	if f.value() != "" {
		t.Errorf("moving the cursor changed the value to %q", f.value())
	}
	pressAll(f, keyEsc)
	if f.mode != selectModeClosed || f.value() != "" {
		t.Errorf("Esc should close without a change, value %q", f.value())
	}

	pressAll(f, keyEnter, keyDown, keyDown, keyDown, keyEnter)
	if f.value() != "~/.ssh/id_b" || f.mode != selectModeClosed {
		t.Errorf("Enter should pick the last row and close, value %q", f.value())
	}

	pressAll(f, keyRight, keyUp, keyLeft)
	if f.value() != "~/.ssh/id_b" {
		t.Errorf("← should close without a change, value %q", f.value())
	}
}

func TestSelectFieldUnknownValue(t *testing.T) {
	f := newSelectField(twoKeys, "~/.ssh/gone", nil)
	if f.value() != "~/.ssh/gone" || f.options[f.selected].Note != "(not found)" {
		t.Errorf("unknown value should become a (not found) row, got %+v", f.options[f.selected])
	}

	custom := newSelectField(twoKeys, "/somewhere/else", &customEntry{label: "Other path"})
	if custom.selected != custom.customRow() || custom.value() != "/somewhere/else" {
		t.Errorf("unknown value should go into the custom input, got %q", custom.value())
	}
}

func TestSelectFieldCustomInput(t *testing.T) {
	invalid := errors.New("invalid")
	f := newSelectField(twoKeys, "", &customEntry{
		label:   "Other path",
		prefill: "/k/",
		validate: func(v string) error {
			if v != "/k/good" {
				return invalid
			}
			return nil
		},
	})

	pressAll(f, keyEnter, keyDown, keyDown, keyDown, keyEnter)
	if f.mode != selectModeType || f.input.Value() != "/k/" {
		t.Fatalf("custom input should open prefilled, mode %v value %q", f.mode, f.input.Value())
	}

	pressAll(f, runes("bad"), keyEnter)
	if !errors.Is(f.err, invalid) || f.mode != selectModeType {
		t.Errorf("invalid input should stay open with an error, err %v", f.err)
	}
	f.blur()
	if f.value() != "" {
		t.Errorf("leaving with invalid input should keep the old value, got %q", f.value())
	}

	pressAll(f, keyEnter, keyDown, keyDown, keyDown, keyEnter)
	for range len("bad") {
		f.update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	pressAll(f, runes("good"), keyEnter)
	if f.value() != "/k/good" || f.mode != selectModeClosed {
		t.Errorf("valid input should be picked, value %q", f.value())
	}
}

func TestSelectFieldSortsMatchingEmailFirst(t *testing.T) {
	keys := parseGPGKeys([]byte(gpgColons))
	options := []SelectOption{
		{Label: "personal", Value: keys[0].fingerprint, key: &keys[0]},
		{Label: "other", Value: keys[1].fingerprint, key: &keys[1]},
	}
	f := newSelectField(options, keys[0].fingerprint, nil)
	f.matchEmail = "other@example.test"

	pressAll(f, keyEnter)
	if f.options[1].Label != "other" {
		t.Errorf("matching key not first: %q", f.options[1].Label)
	}
	if f.value() != keys[0].fingerprint {
		t.Errorf("sorting changed the pick to %q", f.value())
	}
	if f.noteFor(f.options[2]) == "" || f.noteFor(f.options[1]) != "" {
		t.Error("only the key without the email should get a note")
	}
}

func newTestForm(t *testing.T, initial config.Profile, editing bool) tea.Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GNUPGHOME", t.TempDir())
	return newFormModel(initial, editing)
}

func send(m tea.Model, msgs ...tea.Msg) tea.Model {
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

func TestFormRequiresFieldsBeforeSaving(t *testing.T) {
	m := newTestForm(t, config.Profile{}, false)

	m = send(m, runes("work"), keyCtrlS)
	form := m.(formModel)
	if form.submitted || form.focus != fieldGitName || !form.showMissing {
		t.Fatalf("Ctrl+S with missing fields should jump to Git Name, focus %d", form.focus)
	}

	m = send(m, runes("Me"), keyEnter)
	if m.(formModel).focus != fieldGitName {
		t.Error("Enter in a text field must not move focus")
	}

	m = send(m, keyDown, runes("me@x"), keyDown, keyDown, keyDown)
	if form := m.(formModel); form.focus != saveRow || form.submitted {
		t.Fatalf("arrows should reach the Save row without saving, focus %d", form.focus)
	}

	m = send(m, keyEnter)
	form = m.(formModel)
	if !form.submitted {
		t.Fatal("Enter on Save should submit")
	}
	if p := form.toProfile(); p.Name != "work" || p.GitName != "Me" || p.GitEmail != "me@x" {
		t.Errorf("profile = %+v", p)
	}
}

func TestFormEditKeepsNameFixed(t *testing.T) {
	m := newTestForm(t, config.Profile{Name: "work", GitName: "  "}, true)
	m = send(m, keyUp)
	if m.(formModel).focus != fieldGitName {
		t.Error("the fixed profile name must not get focus")
	}
}

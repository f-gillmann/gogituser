package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gogituser/internal/config"
)

type profileItem struct{ p config.Profile }

func (i profileItem) FilterValue() string { return i.p.Name }

type compactDelegate struct{}

func (d compactDelegate) Height() int                             { return 1 }
func (d compactDelegate) Spacing() int                            { return 0 }
func (d compactDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d compactDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item := listItem.(profileItem).p
	cursor := "  "
	nameStyle := normalStyle
	if index == m.Index() {
		cursor = accentStyle.Render("▶ ")
		nameStyle = accentStyle
	}
	_, _ = fmt.Fprintf(w, "%s%s  %s <%s>",
		cursor,
		nameStyle.Render(item.Name),
		dimStyle.Render(item.GitName),
		dimStyle.Render(item.GitEmail),
	)
}

type pickerModel struct {
	list     list.Model
	selected *config.Profile
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(profileItem); ok {
				p := item.p
				m.selected = &p
			}
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string { return m.list.View() }

// Pick shows an interactive profile picker. Returns the selected profile and
// true if a selection was made.
func Pick(profiles []config.Profile) (config.Profile, bool, error) {
	items := make([]list.Item, len(profiles))
	for i, p := range profiles {
		items[i] = profileItem{p}
	}

	l := list.New(items, compactDelegate{}, 60, len(profiles)+4)
	l.Title = "Select a profile  (Enter to apply, q to quit)"
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	m, err := tea.NewProgram(pickerModel{list: l}).Run()
	if err != nil {
		return config.Profile{}, false, err
	}
	pm := m.(pickerModel)
	if pm.selected == nil {
		return config.Profile{}, false, nil
	}
	return *pm.selected, true, nil
}

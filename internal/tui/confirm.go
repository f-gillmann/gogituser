package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmModel struct {
	prompt   string
	choice   bool
	answered bool
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.choice = true
			m.answered = true
			return m, tea.Quit
		case "n", "N", "q", "ctrl+c", "esc":
			m.choice = false
			m.answered = true
			return m, tea.Quit
		}
	}
	return m, nil
}

var promptStyle = lipgloss.NewStyle().Bold(true)

func (m confirmModel) View() string {
	if m.answered {
		return ""
	}
	return promptStyle.Render(m.prompt) + " [y/N] "
}

// Confirm shows a yes/no prompt and returns the user's choice.
func Confirm(prompt string) (bool, error) {
	m, err := tea.NewProgram(confirmModel{prompt: prompt}).Run()
	if err != nil {
		return false, err
	}
	return m.(confirmModel).choice, nil
}

package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func newModel() *model {
	return &model{
		choices: []string{
			"✈️  New Trip",
			"📔 View Journal",
			"💰 Expenses",
			"🛑 Quit",
		},
	}
}
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "w":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "s":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", "":
			selected := m.choices[m.cursor]
			if selected == "🛑 Quit" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Render("Nomadic – Your Travel Journal Companion")

	title += fmt.Sprintf("\n")
	for i, choice := range m.choices {
		cursor := ""
		if m.cursor == i {
			cursor = "👉"
		}
		title += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	return title

}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}

package bubble

import (
	"clx/categories"
	"clx/cli"
	"fmt"
	"os"

	"clx/bubble/list"
	"clx/favorites"
	"clx/settings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle()

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func Run(config *settings.Config, cat *categories.Categories) {
	cli.ClearScreen()

	m := model{list: list.New(list.NewDefaultDelegate(), config, cat, favorites.New(), 0, 0)}

	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

package bubble

import (
	"clx/bubble/list"
	"clx/categories"
	"clx/cli"
	"clx/favorites"
	"clx/file"
	"clx/settings"
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var docStyle = lipgloss.NewStyle()

type model struct {
	list *list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true

	return v
}

func Run(config *settings.Config, cat *categories.Categories) {
	cli.ClearScreen(context.Background())

	m := model{list: list.New(list.NewDefaultDelegate(), config, cat, favorites.New(file.PathToFavoritesFile()), 0, 0)}

	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

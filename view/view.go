package view

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/view/list"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	list *list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && msg.Mod == tea.ModCtrl && msg.Code == 'c' {
		if cmd := m.list.CancelFetch(); cmd != nil {
			return m, cmd
		}

		return m, tea.Quit
	}

	var cmd tea.Cmd

	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.list.View())
	v.AltScreen = true

	return v
}

func Run(config *settings.Config, cat *categories.Categories) {
	m := model{list: list.New(list.NewDefaultDelegate(), config, cat, favorites.New(settings.FavoritesPath()), 0, 0)}

	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

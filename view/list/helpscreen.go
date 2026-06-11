package list

import (
	"github.com/bensadeh/circumflex/help"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) updateHelpScreen(msg tea.Msg) (*Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keymap.Quit, m.keymap.Help) {
			m.state = stateBrowsing

			return m, nil
		}

	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		m.resizeHelpViewport(msg.Width, msg.Height)

		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) resizeHelpViewport(width, height int) {
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height - headerAndFooterHeight)

	content := lipgloss.NewStyle().
		Width(width).
		SetString(help.MainMenuHelpScreen(width, m.keymap.MainMenuSections()))

	m.viewport.SetContent(content.String())
}

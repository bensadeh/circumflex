package list

import (
	"clx/help"

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
			m.state = StateBrowsing

			return m, nil
		}

	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)

		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - headerAndFooterHeight)

		content := lipgloss.NewStyle().
			Width(msg.Width).
			SetString(help.MainMenuHelpScreen(msg.Width, m.keymap.MainMenuBindings()))

		m.viewport.SetContent(content.String())

		return m, nil
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

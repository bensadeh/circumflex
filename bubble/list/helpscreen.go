package list

import (
	"clx/help"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) updateHelpScreen(msg tea.Msg) (*Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" || k == "i" || k == "?" {
			m.isOnHelpScreen = false

			return m, nil
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.setSize(msg.Width-h, msg.Height-v)

		headerHeight := lipgloss.Height("")
		footerHeight := lipgloss.Height("")
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - verticalMarginHeight)

		m.width = msg.Width
		m.height = msg.Height

		content := lipgloss.NewStyle().
			Width(msg.Width).
			AlignHorizontal(lipgloss.Center).
			SetString(help.GetHelpScreen(m.config.EnableNerdFonts))

		m.viewport.SetContent(content.String())

		return m, nil

	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

package view

import (
	"github.com/bensadeh/circumflex/help"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) updateHelpScreen(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keymap.Quit, m.keymap.Help) {
		m.state = stateBrowsing

		return nil
	}

	var cmd tea.Cmd

	m.helpViewport, cmd = m.helpViewport.Update(msg)

	return cmd
}

func (m *model) resizeHelpViewport(width, height int) {
	m.helpViewport.SetWidth(width)
	m.helpViewport.SetHeight(height - headerAndFooterHeight)

	content := lipgloss.NewStyle().
		Width(width).
		SetString(help.MainMenuHelpScreen(width, m.keymap.MainMenuSections()))

	m.helpViewport.SetContent(content.String())
}

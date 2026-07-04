package view

import (
	"fmt"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) updateHelpScreen(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keymap.Quit, m.keymap.Help) {
		m.screen = screenList

		return nil
	}

	var cmd tea.Cmd

	m.helpViewport, cmd = m.helpViewport.Update(msg)

	return cmd
}

// resizeHelpViewport lays help out for the pane it renders in: the detail
// pane in the wide layout, the full screen otherwise.
func (m *model) resizeHelpViewport() {
	width := m.detailWidth()

	m.helpViewport.SetWidth(width)
	m.helpViewport.SetHeight(m.height - headerAndFooterHeight)

	content := lipgloss.NewStyle().
		Width(width).
		SetString(help.MainMenuHelpScreen(width, m.keymap.MainMenuSections()))

	m.helpViewport.SetContent(content.String())
}

// helpView frames the help viewport with the same header and footer rules in
// both layouts; detailWidth is the full screen when the terminal is narrow.
func (m *model) helpView() string {
	width := m.detailWidth()

	return fmt.Sprintf("%s\n%s\n%s\n%s",
		header.HelpHeader("Keyboard Shortcuts", width),
		m.helpViewport.View(),
		m.bottomBar(width),
		help.Footer(width))
}

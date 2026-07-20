package comments

import (
	"github.com/bensadeh/circumflex/view/message"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleGlobalKeys dispatches the keys that work the same in both Read and
// Navigate modes. The bool reports whether the key was consumed.
func (m *Model) handleGlobalKeys(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keymap.Quit):
		return m.quitCmd(), true
	case key.Matches(msg, m.keymap.Help):
		m.showHelp = true

		return nil, true
	case key.Matches(msg, m.keymap.OpenLink):
		return m.openStoryInBrowser(), true
	case key.Matches(msg, m.keymap.OpenComments):
		return m.openCommentsInBrowser(), true
	case key.Matches(msg, m.keymap.NextStory):
		return message.OpenAdjacentStoryCmd(1), true
	case key.Matches(msg, m.keymap.PrevStory):
		return message.OpenAdjacentStoryCmd(-1), true
	case key.Matches(msg, m.keymap.ToggleWide):
		return message.ToggleWideLayoutCmd(), true
	}

	return nil, false
}

// quitCmd closes the view — or, for a thread reached through a link inside
// an article, steps back to the page behind it, parse in hand, exactly like
// the reader's walk-back.
func (m *Model) quitCmd() tea.Cmd {
	if n := len(m.linkTrail); n > 0 {
		return message.RestorePageCmd(m.linkTrail[n-1], m.linkTrail[:n-1])
	}

	return func() tea.Msg { return message.DetailQuit{} }
}

// handleLinkModeKey consumes the selector's own keys first — esc leaves it
// before it can clear a search or quit the view — and lets everything else
// (paging, o, J/K, z) fall through. The bool reports whether the key was
// consumed.
func (m *Model) handleLinkModeKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keymap.LinkMode):
		m.exitLinkMode()

	case key.Matches(msg, m.keymap.NextLink):
		m.moveLink(1)

	case key.Matches(msg, m.keymap.PrevLink):
		m.moveLink(-1)

	// o falls through to the story bindings below — the selector only claims
	// enter, and only for links a view can open in place.
	case key.Matches(msg, m.keymap.OpenSelected):
		if m.currentLink < 0 {
			return nil, true
		}

		if l := m.links[m.currentLink]; l.Viewable {
			return message.OpenReaderLinkCmd(l.URL, m.nextTrail()), true
		}

		return nil, true

	case key.Matches(msg, m.keymap.NavigateMode):
		m.exitLinkMode()
		m.enterNavigateMode()

	case key.Matches(msg, m.keymap.Search):
		m.exitLinkMode()
		m.expandAll()
		m.StartSearchPrompt()

	case key.Matches(msg, m.keymap.Quit):
		m.exitLinkMode()

	default:
		return nil, false
	}

	return nil, true
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if m.showHelp {
		if key.Matches(msg, m.keymap.Quit, m.keymap.Help) {
			m.showHelp = false
		}

		// The layout toggle stays live so the z the help screen documents
		// works right where it is read; the app resizes this view around it.
		if key.Matches(msg, m.keymap.ToggleWide) {
			return message.ToggleWideLayoutCmd()
		}

		return nil
	}

	if m.linkMode {
		if cmd, handled := m.handleLinkModeKey(msg); handled {
			return cmd
		}
	}

	if m.handleSearchKeys(msg) {
		return nil
	}

	// Navigate mode layers like the selector and search: the back keys
	// leave it before they can quit the view.
	if m.mode == modeNavigate && key.Matches(msg, m.keymap.Quit) {
		m.exitNavigateMode()

		return nil
	}

	if cmd, handled := m.handleGlobalKeys(msg); handled {
		return cmd
	}

	switch {
	case key.Matches(msg, m.keymap.LinkMode):
		m.enterLinkMode()

	case key.Matches(msg, m.keymap.NavigateMode):
		m.toggleNavigateMode()

	case key.Matches(msg, m.keymap.GotoTop):
		m.gotoTop()

	case key.Matches(msg, m.keymap.GotoBottom):
		m.gotoBottom()

	case key.Matches(msg, m.keymap.NextTopLevel):
		m.jumpToTopLevel(1)

	case key.Matches(msg, m.keymap.PrevTopLevel):
		m.jumpToTopLevel(-1)

	case key.Matches(msg, m.keymap.Collapse):
		if m.mode == modeRead {
			m.collapseLevel()
		} else {
			m.setCollapsed(true)
		}

	case key.Matches(msg, m.keymap.Expand):
		if m.mode == modeRead {
			m.expandLevel()
		} else {
			m.setCollapsed(false)
		}

	case key.Matches(msg, m.keymap.ToggleCollapse):
		if m.mode == modeRead {
			m.toggleCollapseAll()
		} else {
			m.toggleCollapse()
		}

	case m.mode == modeRead && key.Matches(msg, m.keymap.HalfPageDown):
		m.HalfPageDown()

	case m.mode == modeRead && key.Matches(msg, m.keymap.HalfPageUp):
		m.HalfPageUp()

	case m.mode == modeRead && key.Matches(msg, m.keymap.PageDown):
		m.PageDown()

	case m.mode == modeRead && key.Matches(msg, m.keymap.PageUp):
		m.PageUp()

	case m.mode == modeNavigate:
		return m.handleNavigateKeys(msg)

	default:
		return m.Forward(msg)
	}

	return nil
}

func (m *Model) handleNavigateKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keymap.NextComment):
		m.navigateComment(1)
	case key.Matches(msg, m.keymap.PrevComment):
		m.navigateComment(-1)
	case key.Matches(msg, m.keymap.HalfPageDown):
		m.HalfPageDown()
		m.snapFocusToVisible(1)
	case key.Matches(msg, m.keymap.HalfPageUp):
		m.HalfPageUp()
		m.snapFocusToVisible(-1)
	case key.Matches(msg, m.keymap.PageDown):
		m.PageDown()
		m.snapFocusToVisible(1)
	case key.Matches(msg, m.keymap.PageUp):
		m.PageUp()
		m.snapFocusToVisible(-1)
	default:
		return m.Forward(msg)
	}

	return nil
}

func (m *Model) toggleNavigateMode() {
	if m.mode == modeNavigate {
		m.exitNavigateMode()

		return
	}

	m.enterNavigateMode()
}

func (m *Model) enterNavigateMode() {
	if m.mode == modeNavigate {
		return
	}

	m.mode = modeNavigate

	// Disable viewport j/k so our navigate bindings take over.
	m.Viewport.KeyMap.Up.SetEnabled(false)
	m.Viewport.KeyMap.Down.SetEnabled(false)

	if m.focusedIdx < 0 && len(m.visible) > 0 {
		m.focusedIdx = m.findCommentAtScroll()
	}

	m.syncDecorations()
}

func (m *Model) exitNavigateMode() {
	if m.mode != modeNavigate {
		return
	}

	m.mode = modeRead

	m.Viewport.KeyMap.Up.SetEnabled(true)
	m.Viewport.KeyMap.Down.SetEnabled(true)

	m.focusedIdx = -1

	// Sync expandedDepth to the actual collapse state so the depth
	// indicator matches what's on screen without changing the view.
	m.syncExpandedDepth()

	m.syncDecorations()
}

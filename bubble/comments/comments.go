package comments

import (
	"clx/bubble/list/message"
	"clx/item"
	"clx/settings"
	"clx/style"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

// Model is the Bubble Tea model for the native comment view.
type Model struct {
	viewport viewport.Model
	keymap   KeyMap
	mode     Mode

	story       *item.Story
	flat        []FlatComment
	visible     []int // indices into flat
	focusedIdx  int   // index into visible (-1 = no focus, scroll mode)
	config      *settings.Config
	lastVisited int64

	width  int
	height int
}

// Reserve space for the mode indicator line at the bottom.
const footerHeight = 1

// New creates a new comment view model.
func New(story *item.Story, lastVisited int64, config *settings.Config, width, height int) *Model {
	km := defaultKeyMap()

	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height-footerHeight),
	)

	// Viewport handles j/k in scroll mode (toggled off in navigate mode).
	// h/l are always handled by us (collapse/expand), so disable them on viewport.
	vp.KeyMap = viewport.DefaultKeyMap()
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	flat := flatten(story)
	visible := computeVisible(flat)

	m := Model{
		viewport:    vp,
		keymap:      km,
		mode:        ModeScroll,
		story:       story,
		flat:        flat,
		visible:     visible,
		focusedIdx:  -1, // no focus in scroll mode
		config:      config,
		lastVisited: lastVisited,
		width:       width,
		height:      height,
	}

	m.rebuildContent()

	return &m
}

// Init returns nil; no initial commands needed.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the comment view.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keymap.Quit) {
			return func() tea.Msg { return message.CommentViewQuitMsg{} }
		}

		if key.Matches(msg, m.keymap.ToggleMode) {
			m.toggleMode()

			return nil
		}

		if key.Matches(msg, m.keymap.GotoTop) {
			m.gotoTop()

			return nil
		}

		if key.Matches(msg, m.keymap.GotoBottom) {
			m.gotoBottom()

			return nil
		}

		if m.mode == ModeScroll && key.Matches(msg, m.keymap.NextTopLevel) {
			m.jumpToTopLevel(1)

			return nil
		}

		if m.mode == ModeScroll && key.Matches(msg, m.keymap.PrevTopLevel) {
			m.jumpToTopLevel(-1)

			return nil
		}

		if key.Matches(msg, m.keymap.Collapse) {
			if m.mode == ModeScroll {
				m.collapseAll()
			} else {
				m.collapse()
			}

			return nil
		}

		if key.Matches(msg, m.keymap.Expand) {
			if m.mode == ModeScroll {
				m.expandAll()
			} else {
				m.expand()
			}

			return nil
		}

		if m.mode == ModeNavigate {
			return m.handleNavigateKeys(msg)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - footerHeight)
		m.rebuildContent()

		return nil
	}

	// In scroll mode (or for unhandled keys in navigate mode),
	// delegate to the viewport.
	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)

	return cmd
}

func (m *Model) handleNavigateKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keymap.NextComment):
		m.navigateComment(1)
	case key.Matches(msg, m.keymap.PrevComment):
		m.navigateComment(-1)
	default:
		// Unhandled key — let viewport process it (pgup/pgdn/etc).
		var cmd tea.Cmd

		m.viewport, cmd = m.viewport.Update(msg)

		return cmd
	}

	return nil
}

// View renders the comment view.
func (m *Model) View() string {
	modeIndicator := m.modeIndicator()

	return m.viewport.View() + "\n" + modeIndicator
}

func (m *Model) modeIndicator() string {
	switch m.mode {
	case ModeScroll:
		return style.Bold("SCROLL") + style.Faint("  j/k: scroll  n/N: next/prev thread  h/l: collapse/expand all  g/G: top/bottom  tab: navigate mode")
	case ModeNavigate:
		return style.Bold("NAVIGATE") + style.Faint("  j/k: comments  h/l: collapse/expand  g/G: top/bottom  tab: scroll mode")
	}

	return ""
}

func (m *Model) toggleMode() {
	switch m.mode {
	case ModeScroll:
		m.mode = ModeNavigate

		// Disable viewport j/k so our navigate bindings take over.
		m.viewport.KeyMap.Up.SetEnabled(false)
		m.viewport.KeyMap.Down.SetEnabled(false)

		// Set focus to the comment nearest to the current scroll position.
		if m.focusedIdx < 0 && len(m.visible) > 0 {
			m.focusedIdx = m.findCommentAtScroll()
		}

		m.rebuildContent()

	case ModeNavigate:
		m.mode = ModeScroll

		// Re-enable viewport j/k.
		m.viewport.KeyMap.Up.SetEnabled(true)
		m.viewport.KeyMap.Down.SetEnabled(true)

		m.focusedIdx = -1
		m.rebuildContent()
	}
}

// findCommentAtScroll returns the visible index of the comment closest to
// the current viewport scroll position.
func (m *Model) findCommentAtScroll() int {
	yOffset := m.viewport.YOffset()
	best := 0

	for vi, flatIdx := range m.visible {
		fc := m.flat[flatIdx]
		if fc.StartLine <= yOffset {
			best = vi
		} else {
			break
		}
	}

	return best
}

func (m *Model) rebuildContent() {
	content := renderFromFlat(m.story, m.flat, m.visible, m.focusedIdx, m.config, m.width, m.lastVisited)
	m.viewport.SetContent(content)
}

func (m *Model) collapse() {
	if len(m.visible) == 0 || m.focusedIdx < 0 {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	fc := &m.flat[flatIdx]

	if fc.ChildCount == 0 || fc.Collapsed {
		return
	}

	fc.Collapsed = true
	m.visible = computeVisible(m.flat)

	if m.focusedIdx >= len(m.visible) {
		m.focusedIdx = len(m.visible) - 1
	}

	m.rebuildContent()
	m.scrollToFocused()
}

func (m *Model) expand() {
	if len(m.visible) == 0 || m.focusedIdx < 0 {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	fc := &m.flat[flatIdx]

	if !fc.Collapsed {
		return
	}

	fc.Collapsed = false
	m.visible = computeVisible(m.flat)
	m.rebuildContent()
	m.scrollToFocused()
}

func (m *Model) navigateComment(direction int) {
	if len(m.visible) == 0 {
		return
	}

	newIdx := m.focusedIdx + direction
	if newIdx < 0 || newIdx >= len(m.visible) {
		return
	}

	m.focusedIdx = newIdx
	m.rebuildContent()
	m.scrollToFocused()
}

func (m *Model) jumpToTopLevel(direction int) {
	yOffset := m.viewport.YOffset()

	if direction > 0 {
		for _, flatIdx := range m.visible {
			fc := m.flat[flatIdx]
			if fc.Depth == 0 && fc.StartLine > yOffset+1 {
				m.viewport.SetYOffset(max(0, fc.StartLine-1))

				return
			}
		}
	} else {
		for i := len(m.visible) - 1; i >= 0; i-- {
			fc := m.flat[m.visible[i]]
			if fc.Depth == 0 && fc.StartLine < yOffset {
				m.viewport.SetYOffset(max(0, fc.StartLine-1))

				return
			}
		}
	}
}

func (m *Model) collapseAll() {
	for i := range m.flat {
		if m.flat[i].Depth == 0 && m.flat[i].ChildCount > 0 {
			m.flat[i].Collapsed = true
		}
	}

	m.visible = computeVisible(m.flat)
	m.rebuildContent()
}

func (m *Model) expandAll() {
	for i := range m.flat {
		if m.flat[i].Depth == 0 && m.flat[i].Collapsed {
			m.flat[i].Collapsed = false
		}
	}

	m.visible = computeVisible(m.flat)
	m.rebuildContent()
}

func (m *Model) gotoTop() {
	if m.mode == ModeNavigate && len(m.visible) > 0 {
		m.focusedIdx = 0
		m.rebuildContent()
	}

	m.viewport.GotoTop()
}

func (m *Model) gotoBottom() {
	if m.mode == ModeNavigate && len(m.visible) > 0 {
		m.focusedIdx = len(m.visible) - 1
		m.rebuildContent()
	}

	m.viewport.GotoBottom()
}

func (m *Model) scrollToFocused() {
	if len(m.visible) == 0 || m.focusedIdx < 0 || m.focusedIdx >= len(m.visible) {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	fc := m.flat[flatIdx]

	top := m.viewport.YOffset()
	bottom := top + m.viewport.VisibleLineCount()

	// Only scroll if the focused comment is outside the visible area.
	if fc.StartLine < top {
		// Scrolling up — put comment a few lines below the top.
		m.viewport.SetYOffset(max(0, fc.StartLine-2))
	} else if fc.StartLine+fc.LineCount > bottom {
		// Scrolling down — put the comment's start near the bottom.
		m.viewport.SetYOffset(fc.StartLine - m.viewport.VisibleLineCount() + fc.LineCount + 2)
	}
}

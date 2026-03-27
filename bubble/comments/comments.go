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

	// In scroll mode, viewport handles j/k/h/l. We'll toggle these
	// on/off when switching modes.
	vp.KeyMap = viewport.DefaultKeyMap()

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
	case key.Matches(msg, m.keymap.Collapse):
		m.collapse()
	case key.Matches(msg, m.keymap.Expand):
		m.expand()
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
		return style.Bold("SCROLL") + style.Faint("  j/k: scroll  tab: navigate mode")
	case ModeNavigate:
		return style.Bold("NAVIGATE") + style.Faint("  j/k: comments  h: collapse  l: expand  tab: scroll mode")
	}

	return ""
}

func (m *Model) toggleMode() {
	switch m.mode {
	case ModeScroll:
		m.mode = ModeNavigate

		// Disable viewport j/k/h/l so our navigate bindings take over.
		m.viewport.KeyMap.Up.SetEnabled(false)
		m.viewport.KeyMap.Down.SetEnabled(false)
		m.viewport.KeyMap.Left.SetEnabled(false)
		m.viewport.KeyMap.Right.SetEnabled(false)

		// Set focus to the comment nearest to the current scroll position.
		if m.focusedIdx < 0 && len(m.visible) > 0 {
			m.focusedIdx = m.findCommentAtScroll()
		}

		m.rebuildContent()

	case ModeNavigate:
		m.mode = ModeScroll

		// Re-enable viewport j/k/h/l.
		m.viewport.KeyMap.Up.SetEnabled(true)
		m.viewport.KeyMap.Down.SetEnabled(true)
		m.viewport.KeyMap.Left.SetEnabled(true)
		m.viewport.KeyMap.Right.SetEnabled(true)

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

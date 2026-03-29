package comments

import (
	"clx/bubble/list/message"
	"clx/comment"
	"clx/meta"
	"clx/settings"
	"clx/style"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model is the Bubble Tea model for the native comment view.
type Model struct {
	viewport viewport.Model
	keymap   KeyMap
	mode     Mode

	flat       []FlatComment
	visible    []int // indices into flat
	focusedIdx int   // index into visible (-1 = no focus, scroll mode)
	rc         renderContext
	title      string // story title for the fixed header

	// Per-comment render cache — invalidated on window resize.
	renderCache []cachedComment

	// Rendering artifacts — recomputed on every rebuildContent call.
	baseContent  string        // rendered content without focus styling
	lineMetrics  []LineMetrics // indexed by flat index
	contentLines int           // actual content lines (excluding bottom padding)
}

const (
	headerHeight  = 2 // title + overline separator
	footerHeight  = 2 // underline separator + mode indicator
	scrollPadding = 2 // breathing room above/below when scrolling to a comment
)

// New creates a new comment view model.
func New(thread *comment.Thread, lastVisited int64, config *settings.Config, width, height int) *Model {
	km := defaultKeyMap()

	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height-headerHeight-footerHeight),
	)

	// Viewport handles j/k in scroll mode (toggled off in navigate mode).
	// h/l are always handled by us (collapse/expand), so disable them on viewport.
	vp.KeyMap = viewport.DefaultKeyMap()
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	flat := flatten(thread)

	newComments := comment.NewCommentsCount(thread, lastVisited)
	header := meta.CommentSectionMetaBlock(thread, config, newComments) + "\n"

	m := Model{
		viewport:    vp,
		keymap:      km,
		mode:        ModeScroll,
		flat:        flat,
		focusedIdx:  -1, // no focus in scroll mode
		title:       thread.Title,
		renderCache: make([]cachedComment, len(flat)),
		rc: renderContext{
			header:         header,
			originalPoster: thread.Author,
			firstCommentID: comment.FirstCommentID(thread.Comments),
			config:         config,
			screenWidth:    width,
			viewportHeight: height - headerHeight - footerHeight,
			lastVisited:    lastVisited,
		},
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
		anchorIdx := m.anchorComment()
		screenPos := m.screenPosition(anchorIdx)

		m.rc.screenWidth = msg.Width
		m.rc.viewportHeight = msg.Height - headerHeight - footerHeight
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - headerHeight - footerHeight)
		m.renderCache = make([]cachedComment, len(m.flat))
		m.rebuildContent()
		m.restoreScreenPosition(anchorIdx, screenPos)

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

// View renders the comment view. Focus styling is baked into the viewport
// content by applyFocus, so View just assembles the layout.
func (m *Model) View() string {
	return m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerSeparator() + "\n" + m.modeIndicator()
}

func (m *Model) headerView() string {
	c := lipgloss.NewStyle().Foreground(style.HeaderC())
	l := lipgloss.NewStyle().Foreground(style.HeaderL())
	x := lipgloss.NewStyle().Foreground(style.HeaderX())

	logo := c.Render("  {") + l.Render("…") + x.Render("}  ")
	title := logo + m.title
	separator := strings.Repeat("‾", m.rc.screenWidth)

	return title + "\n" + separator
}

var focusStyle = lipgloss.NewStyle().Reverse(true)

// applyFocus applies focus styling to the cached base content and updates
// the viewport. Content rendering and focus are independent concerns:
// rebuildContent produces unstyled base content, applyFocus layers on
// the visual focus indicator using lipgloss.
func (m *Model) applyFocus() {
	if m.mode != ModeNavigate || m.focusedIdx < 0 || m.focusedIdx >= len(m.visible) {
		m.viewport.SetContent(m.baseContent)

		return
	}

	flatIdx := m.visible[m.focusedIdx]
	lm := m.lineMetrics[flatIdx]

	lines := strings.Split(m.baseContent, "\n")
	if lm.StartLine >= len(lines) {
		m.viewport.SetContent(m.baseContent)

		return
	}

	lines[lm.StartLine] = focusStyle.Render(lines[lm.StartLine])

	m.viewport.SetContent(strings.Join(lines, "\n"))
}

func (m *Model) footerSeparator() string {
	underscore := lipgloss.NewStyle().Underline(true).Render(" ")

	return strings.Repeat(underscore, m.rc.screenWidth)
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

		m.applyFocus()

	case ModeNavigate:
		m.mode = ModeScroll

		// Re-enable viewport j/k.
		m.viewport.KeyMap.Up.SetEnabled(true)
		m.viewport.KeyMap.Down.SetEnabled(true)

		m.focusedIdx = -1

		m.applyFocus()
	}
}

// findCommentAtScroll returns the visible index of the comment whose header
// line is topmost within the current viewport.
func (m *Model) findCommentAtScroll() int {
	yOffset := m.viewport.YOffset()
	bottom := yOffset + m.viewport.VisibleLineCount()

	for vi, flatIdx := range m.visible {
		if m.lineMetrics[flatIdx].StartLine >= yOffset && m.lineMetrics[flatIdx].StartLine < bottom {
			return vi
		}
	}

	return 0
}

func (m *Model) rebuildContent() {
	m.visible = computeVisible(m.flat)
	content, contentLines, metrics := renderFromFlat(m.rc, m.flat, m.visible, m.renderCache)
	m.contentLines = contentLines
	m.lineMetrics = metrics
	m.baseContent = content
	m.applyFocus()
}

func (m *Model) collapse() {
	if len(m.visible) == 0 || m.focusedIdx < 0 {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	fc := &m.flat[flatIdx]

	if fc.DescendantCount == 0 || fc.Collapsed {
		return
	}

	screenPos := m.screenPosition(flatIdx)

	fc.Collapsed = true

	m.rebuildContent()

	if m.focusedIdx >= len(m.visible) {
		m.focusedIdx = len(m.visible) - 1
		m.applyFocus()
	}

	m.restoreScreenPosition(flatIdx, screenPos)
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

	screenPos := m.screenPosition(flatIdx)

	fc.Collapsed = false

	m.rebuildContent()
	m.restoreScreenPosition(flatIdx, screenPos)
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
	m.applyFocus()
	m.scrollToFocused()
}

func (m *Model) jumpToTopLevel(direction int) {
	yOffset := m.viewport.YOffset()

	if direction > 0 {
		for _, flatIdx := range m.visible {
			if m.flat[flatIdx].Depth == 0 && m.lineMetrics[flatIdx].StartLine > yOffset+1 {
				m.viewport.SetYOffset(m.lineMetrics[flatIdx].StartLine)

				return
			}
		}
	} else {
		for i := len(m.visible) - 1; i >= 0; i-- {
			flatIdx := m.visible[i]
			if m.flat[flatIdx].Depth == 0 && m.lineMetrics[flatIdx].StartLine < yOffset {
				m.viewport.SetYOffset(m.lineMetrics[flatIdx].StartLine)

				return
			}
		}
	}
}

func (m *Model) collapseAll() {
	anchorIdx := m.anchorComment()
	screenPos := m.screenPosition(anchorIdx)

	for i := range m.flat {
		if m.flat[i].DescendantCount > 0 {
			m.flat[i].Collapsed = true
		}
	}

	m.rebuildContent()
	m.restoreScreenPosition(anchorIdx, screenPos)
}

func (m *Model) expandAll() {
	anchorIdx := m.anchorComment()
	screenPos := m.screenPosition(anchorIdx)

	for i := range m.flat {
		m.flat[i].Collapsed = false
	}

	m.rebuildContent()
	m.restoreScreenPosition(anchorIdx, screenPos)
}

// anchorComment returns the flat index of the comment nearest to the top of
// the viewport, used to keep the view stable across content rebuilds.
func (m *Model) anchorComment() int {
	yOffset := m.viewport.YOffset()

	best := -1

	for _, flatIdx := range m.visible {
		if m.lineMetrics[flatIdx].StartLine > yOffset+1 {
			break
		}

		best = flatIdx
	}

	return best
}

func (m *Model) screenPosition(flatIdx int) int {
	if flatIdx < 0 {
		return 0
	}

	return m.lineMetrics[flatIdx].StartLine - m.viewport.YOffset()
}

func (m *Model) restoreScreenPosition(flatIdx, screenPos int) {
	if flatIdx < 0 {
		return
	}

	m.viewport.SetYOffset(max(0, m.lineMetrics[flatIdx].StartLine-screenPos))
}

func (m *Model) gotoTop() {
	if m.mode == ModeNavigate && len(m.visible) > 0 {
		m.focusedIdx = 0
		m.applyFocus()
	}

	m.viewport.GotoTop()
}

func (m *Model) gotoBottom() {
	if m.mode == ModeNavigate && len(m.visible) > 0 {
		m.focusedIdx = len(m.visible) - 1
		m.applyFocus()
	}

	// Scroll so the last line of real content is at the bottom of the viewport,
	// ignoring the bottom padding.
	m.viewport.SetYOffset(max(0, m.contentLines-m.rc.viewportHeight))
}

func (m *Model) scrollToFocused() {
	if len(m.visible) == 0 || m.focusedIdx < 0 || m.focusedIdx >= len(m.visible) {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	lm := m.lineMetrics[flatIdx]

	top := m.viewport.YOffset()
	bottom := top + m.viewport.VisibleLineCount()

	// Only scroll if the focused comment is outside the visible area.
	if lm.StartLine < top {
		// Scrolling up — put comment a few lines below the top.
		m.viewport.SetYOffset(max(0, lm.StartLine-scrollPadding))
	} else if lm.StartLine+lm.LineCount > bottom {
		if lm.LineCount >= m.viewport.VisibleLineCount() {
			// Comment is taller than viewport — show its start.
			m.viewport.SetYOffset(max(0, lm.StartLine-scrollPadding))
		} else {
			// Comment fits — scroll just enough to show it fully.
			m.viewport.SetYOffset(lm.StartLine - m.viewport.VisibleLineCount() + lm.LineCount + scrollPadding)
		}
	}
}

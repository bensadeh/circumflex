package comments

import (
	"clx/comment"
	"clx/header"
	"clx/help"
	"clx/layout"
	"clx/meta"
	"clx/settings"
	"clx/style"
	"clx/syntax"
	"clx/view/message"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	termtext "github.com/MichaelMure/go-term-text"
	"github.com/muesli/reflow/truncate"
)

// Model is the Bubble Tea model for the native comment view.
type Model struct {
	viewport viewport.Model
	keymap   keyMap
	mode     mode

	flat          []flatComment
	visible       []int // indices into flat
	focusedIdx    int   // index into visible (-1 = no focus, scroll mode)
	expandedDepth int   // in scroll mode: h/l expand/collapse one level at a time
	maxDepth      int   // deepest comment depth in the tree
	rc            renderContext
	title         string // story title for the fixed header
	showHelp      bool

	// Pre-rendered comment blocks — rebuilt on window resize.
	prerendered []renderedComment

	// Rendering artifacts — recomputed on every rebuildContent or updateViewport call.
	lineMetrics  []lineMetrics // indexed by flat index
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
	vp.KeyMap.HalfPageDown.SetEnabled(false)
	vp.KeyMap.HalfPageUp.SetEnabled(false)
	vp.KeyMap.PageDown.SetEnabled(false)
	vp.KeyMap.PageUp.SetEnabled(false)
	vp.MouseWheelEnabled = false

	flat := flatten(thread)

	newComments := comment.NewCommentsCount(thread, lastVisited)
	commentWidth := min(width-layout.CommentSectionLeftMargin, config.CommentWidth)
	header := buildCommentHeader(thread, config, newComments, commentWidth) + "\n"

	rc := renderContext{
		header:         header,
		originalPoster: thread.Author,
		firstCommentID: comment.FirstCommentID(thread.Comments),
		config:         config,
		screenWidth:    width,
		viewportHeight: height - headerHeight - footerHeight,
		lastVisited:    lastVisited,
		thread:         thread,
		newComments:    newComments,
	}

	md := 0
	for _, fc := range flat {
		if fc.Depth > md {
			md = fc.Depth
		}
	}

	m := Model{
		viewport:      vp,
		keymap:        km,
		mode:          modeScroll,
		flat:          flat,
		focusedIdx:    -1, // no focus in scroll mode
		expandedDepth: 0,  // initial: only top-level visible
		maxDepth:      md,
		title:         thread.Title,
		prerendered:   prerenderComments(rc, flat),
		rc:            rc,
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
		return m.handleKeyPress(msg)
	case tea.MouseWheelMsg:
		if m.showHelp {
			return nil
		}

		delta := m.viewport.MouseWheelDelta
		maxOffset := max(0, m.contentLines-m.rc.viewportHeight)

		switch msg.Button {
		case tea.MouseWheelDown:
			m.viewport.SetYOffset(min(m.viewport.YOffset()+delta, maxOffset))
		case tea.MouseWheelUp:
			m.viewport.SetYOffset(max(0, m.viewport.YOffset()-delta))
		}

		return nil

	case tea.WindowSizeMsg:
		anchorIdx := m.anchorComment()
		screenPos := m.screenPosition(anchorIdx)

		m.rc.screenWidth = msg.Width
		m.rc.viewportHeight = msg.Height - headerHeight - footerHeight
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - headerHeight - footerHeight)

		commentWidth := min(msg.Width-layout.CommentSectionLeftMargin, m.rc.config.CommentWidth)
		m.rc.header = buildCommentHeader(m.rc.thread, m.rc.config, m.rc.newComments, commentWidth) + "\n"

		m.prerendered = prerenderComments(m.rc, m.flat)
		m.rebuildContent()
		m.restoreScreenPosition(anchorIdx, screenPos)

		return nil
	}

	return nil
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if m.showHelp {
		if key.Matches(msg, m.keymap.Quit, m.keymap.Help) {
			m.showHelp = false
		}

		return nil
	}

	if key.Matches(msg, m.keymap.Quit) {
		return func() tea.Msg { return message.CommentViewQuitMsg{} }
	}

	if key.Matches(msg, m.keymap.Help) {
		m.showHelp = true

		return nil
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

	if key.Matches(msg, m.keymap.NextTopLevel) {
		m.jumpToTopLevel(1)

		return nil
	}

	if key.Matches(msg, m.keymap.PrevTopLevel) {
		m.jumpToTopLevel(-1)

		return nil
	}

	if key.Matches(msg, m.keymap.Collapse) {
		if m.mode == modeScroll {
			m.collapseLevel()
		} else {
			m.setCollapsed(true)
		}

		return nil
	}

	if key.Matches(msg, m.keymap.Expand) {
		if m.mode == modeScroll {
			m.expandLevel()
		} else {
			m.setCollapsed(false)
		}

		return nil
	}

	if key.Matches(msg, m.keymap.ToggleCollapse) {
		if m.mode == modeScroll {
			m.toggleCollapseAll()
		} else {
			m.toggleCollapse()
		}

		return nil
	}

	if m.mode == modeScroll && key.Matches(msg, m.keymap.HalfPageDown) {
		m.halfPageDown()

		return nil
	}

	if m.mode == modeScroll && key.Matches(msg, m.keymap.HalfPageUp) {
		m.halfPageUp()

		return nil
	}

	if m.mode == modeScroll && key.Matches(msg, m.keymap.PageDown) {
		m.pageDown()

		return nil
	}

	if m.mode == modeScroll && key.Matches(msg, m.keymap.PageUp) {
		m.pageUp()

		return nil
	}

	if m.mode == modeNavigate {
		return m.handleNavigateKeys(msg)
	}

	// In scroll mode, delegate unhandled keys to the viewport.
	before := m.viewport.YOffset()

	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	m.clampScroll(before)

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
		before := m.viewport.YOffset()

		var cmd tea.Cmd

		m.viewport, cmd = m.viewport.Update(msg)
		m.clampScroll(before)

		return cmd
	}

	return nil
}

// View renders the comment view.
func (m *Model) View() string {
	if m.showHelp {
		content := help.FitToHeight(
			help.CommentHelpScreen(m.rc.screenWidth, m.rc.config.EnableNerdFonts),
			m.rc.viewportHeight,
		)

		return header.HelpHeader("Comment Section", m.rc.screenWidth) + "\n" +
			content + "\n" +
			m.footerSeparator() + "\n" +
			help.Footer(m.rc.screenWidth)
	}

	return m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerSeparator() + "\n" + m.modeIndicator()
}

func (m *Model) headerView() string {
	leftMargin := strings.Repeat(" ", layout.CommentSectionLeftMargin)
	maxTitleWidth := uint(max(0, m.rc.screenWidth-layout.CommentSectionLeftMargin))
	title := truncate.StringWithTail(m.title, maxTitleWidth, "…")

	if !m.rc.config.DisableHeadlineHighlighting {
		nf := m.rc.config.EnableNerdFonts
		title = syntax.HighlightYCStartupsInHeadlines(title, syntax.Unselected, nf)
		title = syntax.HighlightYear(title, syntax.Unselected)
		title = syntax.HighlightHackerNewsHeadlines(title, syntax.Unselected)
		title = syntax.HighlightSpecialContent(title, syntax.Unselected, nf)
	}

	title = leftMargin + title
	separator := strings.Repeat("‾", m.rc.screenWidth)

	return title + "\n" + separator
}

// updateViewport re-renders the viewport content with the current focus state.
// This is cheap: it concatenates pre-rendered strings, picking the focused
// header variant for the focused comment.
func (m *Model) updateViewport() {
	focusedFlatIdx := -1
	if m.mode == modeNavigate && m.focusedIdx >= 0 && m.focusedIdx < len(m.visible) {
		focusedFlatIdx = m.visible[m.focusedIdx]
	}

	content, contentLines, metrics := renderFromFlat(m.rc, m.flat, m.visible, m.prerendered, focusedFlatIdx)
	m.contentLines = contentLines
	m.lineMetrics = metrics
	m.viewport.SetContent(content)
}

func (m *Model) footerSeparator() string {
	return lipgloss.NewStyle().Underline(true).
		Width(m.rc.screenWidth).
		Render(strings.Repeat(" ", m.rc.screenWidth))
}

func (m *Model) modeIndicator() string {
	var left string

	switch m.mode {
	case modeScroll:
		left = style.ModeIndicator([]style.Binding{
			{Key: "⇥", Desc: "navigate mode"},
		})
	case modeNavigate:
		left = style.ModeIndicator([]style.Binding{
			{Key: "⇥", Desc: "read mode"},
		})
	}

	help := style.RenderBinding(style.Binding{Key: "i", Desc: "help"})

	// In scroll mode, reserve a fixed-width slot for the depth indicator
	// so that "i: help" stays in place as levels expand/collapse.
	diSlot := 0
	if m.mode == modeScroll && m.maxDepth > 0 {
		diSlot = 1 + 1 + len(fmt.Sprintf("%d", m.maxDepth)) + 1 // " ⋮" + digits + " "
	}

	commentWidth := min(m.rc.screenWidth-layout.CommentSectionLeftMargin, m.rc.config.CommentWidth)
	totalWidth := layout.CommentSectionLeftMargin + commentWidth
	padding := max(1, totalWidth-lipgloss.Width(left)-lipgloss.Width(help)-diSlot)

	result := left + strings.Repeat(" ", padding) + help

	if diSlot > 0 {
		di := m.depthIndicator()
		if di != "" {
			used := 1 + lipgloss.Width(di)
			result += " " + di + strings.Repeat(" ", max(0, diSlot-used))
		} else {
			result += strings.Repeat(" ", diSlot)
		}
	}

	return result
}

func (m *Model) depthIndicator() string {
	level := m.expandedDepth
	numStr := fmt.Sprintf("%d", level)

	cycle := style.IndentCycle()

	if level == 0 {
		return ""
	}

	if len(cycle) == 0 {
		return "\u22ee" + style.Faint(numStr) + " "
	}

	colorFn := cycle[(level-1)%len(cycle)]

	return "\u22ee" + colorFn(numStr) + " "
}

func (m *Model) toggleMode() {
	switch m.mode {
	case modeScroll:
		m.mode = modeNavigate

		// Disable viewport j/k so our navigate bindings take over.
		m.viewport.KeyMap.Up.SetEnabled(false)
		m.viewport.KeyMap.Down.SetEnabled(false)

		// Set focus to the comment nearest to the current scroll position.
		if m.focusedIdx < 0 && len(m.visible) > 0 {
			m.focusedIdx = m.findCommentAtScroll()
		}

		m.updateViewport()

	case modeNavigate:
		m.mode = modeScroll

		// Re-enable viewport j/k.
		m.viewport.KeyMap.Up.SetEnabled(true)
		m.viewport.KeyMap.Down.SetEnabled(true)

		m.focusedIdx = -1

		// Sync expandedDepth to the actual collapse state so the depth
		// indicator matches what's on screen without changing the view.
		m.syncExpandedDepth()

		m.updateViewport()
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
	m.updateViewport()
}

func (m *Model) setCollapsed(collapsed bool) {
	if len(m.visible) == 0 || m.focusedIdx < 0 {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	fc := &m.flat[flatIdx]

	if fc.DescendantCount == 0 || fc.Collapsed == collapsed {
		return
	}

	screenPos := m.screenPosition(flatIdx)

	fc.Collapsed = collapsed

	m.rebuildContent()
	m.syncExpandedDepth()

	if m.focusedIdx >= len(m.visible) {
		m.focusedIdx = len(m.visible) - 1
		m.updateViewport()
	}

	m.restoreScreenPosition(flatIdx, screenPos)
}

func (m *Model) toggleCollapse() {
	if len(m.visible) == 0 || m.focusedIdx < 0 {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	fc := &m.flat[flatIdx]

	if fc.DescendantCount == 0 {
		return
	}

	m.setCollapsed(!fc.Collapsed)
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
	m.updateViewport()
	m.scrollToFocused()
}

func (m *Model) jumpToTopLevel(direction int) {
	yOffset := m.viewport.YOffset()

	if direction > 0 {
		for vi, flatIdx := range m.visible {
			if m.flat[flatIdx].Depth == 0 && m.lineMetrics[flatIdx].StartLine > yOffset {
				m.viewport.SetYOffset(m.lineMetrics[flatIdx].StartLine)
				m.setFocusIfNavigating(vi)

				return
			}
		}
	} else {
		for i := len(m.visible) - 1; i >= 0; i-- {
			flatIdx := m.visible[i]
			if m.flat[flatIdx].Depth == 0 && m.lineMetrics[flatIdx].StartLine < yOffset {
				m.viewport.SetYOffset(m.lineMetrics[flatIdx].StartLine)
				m.setFocusIfNavigating(i)

				return
			}
		}

		if yOffset > 0 {
			m.viewport.SetYOffset(0)
		}
	}
}

func (m *Model) setFocusIfNavigating(visibleIdx int) {
	if m.mode != modeNavigate {
		return
	}

	m.focusedIdx = visibleIdx
	m.updateViewport()
}

func (m *Model) toggleCollapseAll() {
	allCollapsed := true

	for i := range m.flat {
		if m.flat[i].DescendantCount > 0 && !m.flat[i].Collapsed {
			allCollapsed = false

			break
		}
	}

	if allCollapsed {
		m.expandAll()
	} else {
		m.collapseAll()
	}
}

func (m *Model) collapseAll() {
	m.expandedDepth = 0
	m.setCollapseToDepth()
}

func (m *Model) expandAll() {
	m.expandedDepth = m.maxDepth
	m.setCollapseToDepth()
}

// setCollapseToDepth sets collapse state based on expandedDepth:
// comments at depth < expandedDepth are uncollapsed; the rest are collapsed.
func (m *Model) setCollapseToDepth() {
	anchorIdx := m.anchorComment()
	screenPos := m.screenPosition(anchorIdx)

	for i := range m.flat {
		if m.flat[i].DescendantCount == 0 {
			continue
		}

		m.flat[i].Collapsed = m.flat[i].Depth >= m.expandedDepth
	}

	m.rebuildContent()

	// If the anchor is still visible after the collapse, restore its exact
	// screen position so the viewport stays stable.
	if anchorIdx >= 0 && m.lineMetrics[anchorIdx].LineCount > 0 {
		m.restoreScreenPosition(anchorIdx, screenPos)

		return
	}

	// The anchor was collapsed away. Find the nearest visible ancestor,
	// then position the viewport at the next visible comment at the same
	// depth or shallower (the next sibling or uncle). This works at any
	// nesting level so collapsing never jumps out further than necessary.
	ancestorIdx := -1

	for i := anchorIdx - 1; i >= 0; i-- {
		if m.lineMetrics[i].LineCount > 0 {
			ancestorIdx = i

			break
		}
	}

	if ancestorIdx < 0 {
		return
	}

	ancestorDepth := m.flat[ancestorIdx].Depth

	for _, flatIdx := range m.visible {
		if flatIdx > ancestorIdx && m.flat[flatIdx].Depth <= ancestorDepth {
			m.viewport.SetYOffset(m.lineMetrics[flatIdx].SepStart)

			return
		}
	}

	// No next sibling — position at the end of the ancestor.
	lm := m.lineMetrics[ancestorIdx]
	m.viewport.SetYOffset(lm.StartLine + lm.LineCount)
}

// syncExpandedDepth derives expandedDepth from the actual collapse state,
// so the depth indicator matches what's on screen after navigate mode
// may have individually collapsed/expanded comments.
func (m *Model) syncExpandedDepth() {
	maxUncollapsed := -1

	for i := range m.flat {
		if m.flat[i].DescendantCount > 0 && !m.flat[i].Collapsed && m.flat[i].Depth > maxUncollapsed {
			maxUncollapsed = m.flat[i].Depth
		}
	}

	m.expandedDepth = maxUncollapsed + 1
}

func (m *Model) expandLevel() {
	if m.expandedDepth >= m.maxDepth {
		return
	}

	m.expandedDepth++
	m.setCollapseToDepth()
}

func (m *Model) collapseLevel() {
	if m.expandedDepth <= 0 {
		return
	}

	m.expandedDepth--
	m.setCollapseToDepth()
}

// anchorComment returns the flat index of the comment nearest to the top of
// the viewport, used to keep the view stable across content rebuilds.
// Uses SepStart so a comment whose separator is visible at the viewport
// top is chosen as the anchor rather than the previous comment.
func (m *Model) anchorComment() int {
	yOffset := m.viewport.YOffset()

	best := -1

	for _, flatIdx := range m.visible {
		if m.lineMetrics[flatIdx].SepStart > yOffset {
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

// clampScroll prevents scrolling down past the last content line while still
// allowing upward scrolling from a position beyond the clamp point (e.g. after
// an n/N top-level jump).
func (m *Model) clampScroll(before int) {
	maxOffset := max(0, m.contentLines-m.rc.viewportHeight)
	after := m.viewport.YOffset()

	if after > before && after > maxOffset {
		m.viewport.SetYOffset(max(before, maxOffset))
	}
}

func (m *Model) halfPageDown() {
	halfPage := m.rc.viewportHeight / 2
	maxOffset := max(0, m.contentLines-m.rc.viewportHeight)
	m.viewport.SetYOffset(min(m.viewport.YOffset()+halfPage, maxOffset))
}

func (m *Model) halfPageUp() {
	halfPage := m.rc.viewportHeight / 2
	m.viewport.SetYOffset(max(0, m.viewport.YOffset()-halfPage))
}

func (m *Model) pageDown() {
	maxOffset := max(0, m.contentLines-m.rc.viewportHeight)
	m.viewport.SetYOffset(min(m.viewport.YOffset()+m.rc.viewportHeight, maxOffset))
}

func (m *Model) pageUp() {
	m.viewport.SetYOffset(max(0, m.viewport.YOffset()-m.rc.viewportHeight))
}

func (m *Model) gotoTop() {
	if m.mode == modeNavigate && len(m.visible) > 0 {
		m.focusedIdx = 0
		m.updateViewport()
	}

	m.viewport.GotoTop()
}

func (m *Model) gotoBottom() {
	if m.mode == modeNavigate && len(m.visible) > 0 {
		m.focusedIdx = len(m.visible) - 1
		m.updateViewport()
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

func buildCommentHeader(t *comment.Thread, config *settings.Config, newComments int, width int) string {
	rootComment := renderRootComment(t.Content, config, width-boxOverhead)

	return meta.CommentSectionMetaBlock(t.URL, t.Domain, t.Author, t.TimeAgo, t.ID, t.CommentsCount, t.Points, newComments, config.EnableNerdFonts, rootComment, width)
}

const boxOverhead = 4 // meta block border (2) + padding (2)

func renderRootComment(c string, config *settings.Config, contentWidth int) string {
	if c == "" {
		return ""
	}

	rendered := comment.Print(c, config, contentWidth, contentWidth)
	wrapped, _ := termtext.Wrap(rendered, contentWidth)

	return "\n\n" + wrapped
}

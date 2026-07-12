package comments

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/view/message"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackKeysReturnCommentViewQuit(t *testing.T) {
	keys := []tea.KeyPressMsg{
		{Code: 'q', Text: "q"},
		{Code: tea.KeyEsc},
		{Code: tea.KeyBackspace},
	}

	for _, key := range keys {
		m := New(testThread(), 0, 80, 1, false, 120, 30)

		cmd := m.handleKeyPress(key)
		require.NotNil(t, cmd)

		_, ok := cmd().(message.CommentViewQuit)
		assert.True(t, ok, "back key should produce CommentViewQuit")
	}
}

// testThread builds a small but representative tree:
//
//	A (depth 0, has children)
//	  B (depth 1, has children)
//	    C (depth 2, leaf)
//	  D (depth 1, leaf)
//	E (depth 0, leaf)
func testThread() *comment.Thread {
	return newThread(
		newComment(1, "alice", "A",
			newComment(2, "bob", "B",
				newComment(3, "charlie", "C"),
			),
			newComment(4, "dave", "D"),
		),
		newComment(5, "eve", "E"),
	)
}

// newTestModel creates a Model from a thread with a generous viewport so
// scroll-clamping doesn't interfere with navigation tests.
func newTestModel(t *testing.T, thread *comment.Thread) *Model {
	t.Helper()

	return New(thread, 0, 80, 1, false, 120, 200)
}

func TestModeIndicator_NerdFontIcons(t *testing.T) {
	m := New(testThread(), 0, 80, 1, true, 120, 200)

	assert.Contains(t, m.modeIndicator(), nerdfonts.Document+"  ", "read mode shows the reader-mode glyph, with extra room after the wide glyph")

	// The thread starts fully collapsed, so the first focused comment
	// offers expanding.
	m.toggleMode()
	assert.Contains(t, m.modeIndicator(), nerdfonts.CommentPlusOutline, "collapsed focused comment offers expanding")

	m.setCollapsed(false)
	assert.Contains(t, m.modeIndicator(), nerdfonts.CommentMinusOutline, "expanded focused comment offers collapsing")

	m.gotoBottom() // E, a leaf
	assert.Contains(t, m.modeIndicator(), nerdfonts.CommentDraft, "a leaf has nothing to toggle")
}

func TestModeIndicator_UnicodeFallback(t *testing.T) {
	m := newTestModel(t, testThread())

	assert.Contains(t, m.modeIndicator(), "☰ ")
	assert.NotContains(t, m.modeIndicator(), "☰  ", "unicode glyphs are single-cell and need no extra room")

	m.toggleMode()
	assert.Contains(t, m.modeIndicator(), "+ ", "collapsed focused comment offers expanding")

	m.setCollapsed(false)
	assert.Contains(t, m.modeIndicator(), "− ", "expanded focused comment offers collapsing")

	m.gotoBottom() // E, a leaf
	assert.Contains(t, m.modeIndicator(), "… ", "a leaf has nothing to toggle")
}

func TestOpenInBrowser_ReturnsCmd(t *testing.T) {
	m := newTestModel(t, testThread())

	assert.NotNil(t, m.openStoryInBrowser())
	assert.NotNil(t, m.openCommentsInBrowser())
}

func TestNew_StartsInReadMode(t *testing.T) {
	m := newTestModel(t, testThread())

	assert.Equal(t, modeRead, m.mode)
	assert.Equal(t, -1, m.focusedIdx, "no focus in read mode")
}

func TestNew_StartsFullyCollapsed(t *testing.T) {
	m := newTestModel(t, testThread())

	assert.Equal(t, 0, m.expandedDepth)

	for _, vi := range m.visible {
		assert.Equal(t, 0, m.flat[vi].Depth)
	}
}

func TestExpandLevel_RevealsChildren(t *testing.T) {
	m := newTestModel(t, testThread())

	initialCount := len(m.visible)

	m.expandLevel()
	assert.Equal(t, 1, m.expandedDepth)
	assert.Greater(t, len(m.visible), initialCount, "expanding should reveal more comments")

	hasDepth1 := false

	for _, vi := range m.visible {
		if m.flat[vi].Depth == 1 {
			hasDepth1 = true

			break
		}
	}

	assert.True(t, hasDepth1, "depth-1 comments should be visible after expand")
}

func TestExpandLevel_FullExpand(t *testing.T) {
	m := newTestModel(t, testThread())

	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	assert.Len(t, m.visible, len(m.flat), "fully expanded should show all comments")
}

func TestCollapseLevel_HidesChildren(t *testing.T) {
	m := newTestModel(t, testThread())

	m.expandLevel()
	expanded := len(m.visible)

	m.collapseLevel()
	assert.Less(t, len(m.visible), expanded)
	assert.Equal(t, 0, m.expandedDepth)
}

func TestCollapseLevel_ClampsAtZero(t *testing.T) {
	m := newTestModel(t, testThread())

	m.collapseLevel()
	assert.Equal(t, 0, m.expandedDepth)
}

func TestExpandLevel_ClampsAtMax(t *testing.T) {
	m := newTestModel(t, testThread())

	for range m.maxDepth + 5 {
		m.expandLevel()
	}

	assert.Equal(t, m.maxDepth, m.expandedDepth)
}

func TestToggleCollapseAll_ExpandsThenCollapses(t *testing.T) {
	m := newTestModel(t, testThread())

	m.toggleCollapseAll()
	assert.Len(t, m.visible, len(m.flat))

	m.toggleCollapseAll()

	for _, vi := range m.visible {
		assert.Equal(t, 0, m.flat[vi].Depth)
	}
}

func TestToggleMode_SwitchesToNavigate(t *testing.T) {
	m := newTestModel(t, testThread())

	m.toggleMode()
	assert.Equal(t, modeNavigate, m.mode)
	assert.GreaterOrEqual(t, m.focusedIdx, 0, "should set focus on mode switch")
}

func TestToggleMode_SwitchesBackToRead(t *testing.T) {
	m := newTestModel(t, testThread())

	m.toggleMode()
	m.toggleMode()

	assert.Equal(t, modeRead, m.mode)
	assert.Equal(t, -1, m.focusedIdx, "focus cleared in read mode")
}

func TestNavigateComment_MovesForward(t *testing.T) {
	m := newTestModel(t, testThread())
	m.toggleMode()
	require.Equal(t, modeNavigate, m.mode)

	initial := m.focusedIdx
	m.navigateComment(1)
	assert.Equal(t, initial+1, m.focusedIdx)
}

func TestNavigateComment_ClampsAtBounds(t *testing.T) {
	m := newTestModel(t, testThread())
	m.toggleMode()

	m.focusedIdx = 0
	m.navigateComment(-1)
	assert.Equal(t, 0, m.focusedIdx, "should not go below 0")

	m.focusedIdx = len(m.visible) - 1
	m.navigateComment(1)
	assert.Equal(t, len(m.visible)-1, m.focusedIdx, "should not exceed visible length")
}

func TestSetCollapsed_CollapsesAndExpands(t *testing.T) {
	m := newTestModel(t, testThread())

	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()

	m.focusedIdx = 0
	flatIdx := m.visible[m.focusedIdx]
	require.Positive(t, m.flat[flatIdx].DescendantCount)

	visibleBefore := len(m.visible)

	m.setCollapsed(true)
	assert.True(t, m.flat[flatIdx].Collapsed)
	assert.Less(t, len(m.visible), visibleBefore)

	m.setCollapsed(false)
	assert.False(t, m.flat[flatIdx].Collapsed)
	assert.Len(t, m.visible, visibleBefore)
}

func TestSetCollapsed_NoOpOnLeaf(t *testing.T) {
	m := newTestModel(t, testThread())

	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()

	for vi, fi := range m.visible {
		if m.flat[fi].DescendantCount == 0 {
			m.focusedIdx = vi

			break
		}
	}

	visibleBefore := len(m.visible)

	m.setCollapsed(true)
	assert.Len(t, m.visible, visibleBefore, "collapsing a leaf should be a no-op")
}

func TestToggleCollapse_Toggles(t *testing.T) {
	m := newTestModel(t, testThread())

	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()

	m.focusedIdx = 0
	flatIdx := m.visible[m.focusedIdx]
	require.Positive(t, m.flat[flatIdx].DescendantCount)

	collapsed := m.flat[flatIdx].Collapsed
	m.toggleCollapse()
	assert.NotEqual(t, collapsed, m.flat[flatIdx].Collapsed)

	m.toggleCollapse()
	assert.Equal(t, collapsed, m.flat[flatIdx].Collapsed)
}

func TestGotoTop_Navigate(t *testing.T) {
	m := newTestModel(t, testThread())
	m.toggleMode()

	m.focusedIdx = len(m.visible) - 1
	m.gotoTop()
	assert.Equal(t, 0, m.focusedIdx)
}

func TestGotoBottom_Navigate(t *testing.T) {
	m := newTestModel(t, testThread())
	m.toggleMode()

	m.gotoBottom()
	assert.Equal(t, len(m.visible)-1, m.focusedIdx)
}

// deepThread builds a tree large enough that content exceeds the viewport,
// making scroll anchoring observable. Structure:
//
//	T1 (depth 0)
//	  T1-R1 (depth 1)
//	    T1-R1-R1 (depth 2)
//	  T1-R2 (depth 1)
//	T2 (depth 0)
//	  T2-R1 (depth 1)
//	    T2-R1-R1 (depth 2)
//	  T2-R2 (depth 1)
//	T3 (depth 0)
//	  T3-R1 (depth 1)
//	T4 (depth 0)
//	  T4-R1 (depth 1)
//	    T4-R1-R1 (depth 2)
func deepThread() *comment.Thread {
	return newThread(
		newComment(10, "a", "Top-level 1 with enough text to occupy lines",
			newComment(11, "b", "Reply to T1 with some content here",
				newComment(12, "c", "Nested reply deep in T1"),
			),
			newComment(13, "d", "Second reply to T1"),
		),
		newComment(20, "e", "Top-level 2 with another block of text",
			newComment(21, "f", "Reply to T2 with content",
				newComment(22, "g", "Nested reply in T2"),
			),
			newComment(23, "h", "Second reply to T2"),
		),
		newComment(30, "i", "Top-level 3",
			newComment(31, "j", "Reply to T3"),
		),
		newComment(40, "k", "Top-level 4",
			newComment(41, "l", "Reply to T4",
				newComment(42, "m", "Nested in T4"),
			),
		),
	)
}

// newScrollableModel creates a model with a small viewport (height 30)
// so that expanded content overflows and scroll anchoring is exercised.
func newScrollableModel(t *testing.T) *Model {
	t.Helper()

	return New(deepThread(), 0, 80, 1, false, 120, 30)
}

func TestNavigateMode_PageKeysScrollAndSnapFocus(t *testing.T) {
	m := newScrollableModel(t)

	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()
	require.Equal(t, modeNavigate, m.mode)

	top := m.Viewport.YOffset()
	m.handleKeyPress(tea.KeyPressMsg{Code: 'f', Text: "f"})

	assert.Greater(t, m.Viewport.YOffset(), top, "page down should scroll in navigate mode")
	assertFocusOnScreen(t, m)

	scrolled := m.Viewport.YOffset()
	m.handleKeyPress(tea.KeyPressMsg{Code: 'b', Text: "b"})

	assert.Less(t, m.Viewport.YOffset(), scrolled, "page up should scroll in navigate mode")
	assertFocusOnScreen(t, m)
}

func assertFocusOnScreen(t *testing.T, m *Model) {
	t.Helper()

	require.GreaterOrEqual(t, m.focusedIdx, 0)

	start := m.lineMetrics[m.visible[m.focusedIdx]].StartLine
	assert.GreaterOrEqual(t, start, m.Viewport.YOffset())
	assert.Less(t, start, m.Viewport.YOffset()+m.Viewport.VisibleLineCount())
}

func TestViewportStable_ExpandLevel(t *testing.T) {
	m := newScrollableModel(t)

	m.expandLevel()
	m.Viewport.SetYOffset(m.ContentLines / 3)

	anchor := m.anchorComment()
	posBefore := m.screenPosition(anchor)

	m.expandLevel()

	posAfter := m.screenPosition(anchor)
	assert.Equal(t, posBefore, posAfter,
		"anchor comment should not move on screen after expanding a level")
}

func TestViewportStable_CollapseLevel(t *testing.T) {
	m := newScrollableModel(t)

	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.Viewport.SetYOffset(m.ContentLines / 3)

	anchor := m.anchorComment()
	posBefore := m.screenPosition(anchor)

	m.collapseLevel()

	// If the anchor is still visible, its position should be preserved.
	if m.lineMetrics[anchor].LineCount > 0 {
		posAfter := m.screenPosition(anchor)
		assert.Equal(t, posBefore, posAfter,
			"anchor comment should not move on screen after collapsing a level")
	}
}

func TestViewportStable_IndividualCollapse(t *testing.T) {
	m := newScrollableModel(t)

	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()
	m.Viewport.SetYOffset(m.ContentLines / 3)

	found := false

	for vi, fi := range m.visible {
		if m.flat[fi].DescendantCount > 0 && m.lineMetrics[fi].StartLine >= m.Viewport.YOffset() {
			m.focusedIdx = vi
			found = true

			break
		}
	}

	require.True(t, found, "need a collapsible comment in the viewport")

	flatIdx := m.visible[m.focusedIdx]
	posBefore := m.screenPosition(flatIdx)

	m.setCollapsed(true)

	posAfter := m.screenPosition(flatIdx)
	assert.Equal(t, posBefore, posAfter,
		"individually collapsed comment should stay in place on screen")
}

func TestViewportStable_Resize(t *testing.T) {
	m := newScrollableModel(t)

	m.expandLevel()
	m.Viewport.SetYOffset(m.ContentLines / 3)

	anchor := m.anchorComment()
	posBefore := m.screenPosition(anchor)

	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	posAfter := m.screenPosition(anchor)
	assert.Equal(t, posBefore, posAfter,
		"anchor comment should not move on screen after resize")
}

func linearChain(depth int) []flatComment {
	var children []*comment.Comment
	for i := depth; i >= 1; i-- {
		c := newComment(i, "u", "body", children...)
		children = []*comment.Comment{c}
	}

	return flatten(newThread(children...))
}

func leadingIndentCols(lines []string) int {
	first := lines[0]

	return len(first) - len(strings.TrimLeft(first, " ")) - layout.CommentSectionLeftMargin
}

func TestPrerenderComments_IndentPlateausUnderDeepNesting(t *testing.T) {
	t.Parallel()

	flat := linearChain(12)
	require.Len(t, flat, 12)

	const (
		commentWidth = 70
		indentSize   = 5
	)

	rc := renderContext{
		commentWidth: commentWidth,
		indent:       indentSize,
		paneWidth:    120,
	}

	rendered := prerenderComments(rc, flat)

	// Floor for depth >= 1 is MinCommentWidth(40) + symbolCol(1) = 41.
	// Headroom = commentWidth(70) - floor(41) = 29.
	// Desired = (depth - 1) * 5, capped at 29. Plateau begins at depth 7.
	wantIndent := []int{0, 0, 5, 10, 15, 20, 25, 29, 29, 29, 29, 29}

	for i := range flat {
		got := leadingIndentCols(rendered[i].content)
		assert.Equalf(t, wantIndent[i], got, "flat[%d] depth=%d", i, flat[i].Depth)

		symbolCols := 0
		if flat[i].Depth > 0 {
			symbolCols = 1
		}

		adjusted := commentWidth - got - symbolCols
		assert.GreaterOrEqualf(t, adjusted, layout.MinCommentWidth, "flat[%d] depth=%d", i, flat[i].Depth)
	}
}

// indicatorLeadingCols returns the count of leading spaces on the line
// containing the ↩ marker, or -1 if not found. Leading spaces are always
// plain ASCII; any ANSI escapes sit to the right of them.
func indicatorLeadingCols(lines []string) int {
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if strings.Contains(trimmed, "↩") {
			return len(line) - len(trimmed)
		}
	}

	return -1
}

func TestPrerenderComments_RepliesIndicatorAlignsWithChildAuthor(t *testing.T) {
	t.Parallel()

	flat := linearChain(10)

	const (
		commentWidth = 70
		indentSize   = 5
	)

	rc := renderContext{
		commentWidth: commentWidth,
		indent:       indentSize,
		paneWidth:    120,
	}

	rendered := prerenderComments(rc, flat)

	// Every comment in the chain except the last has exactly one child at the
	// next flatten index. The indicator's ↩ column should equal the child's
	// author column (content-indent + 1 col for the ▎ position).
	for i := range len(flat) - 1 {
		require.Positivef(t, flat[i].DescendantCount, "flat[%d] should have descendants", i)

		indicatorCol := indicatorLeadingCols(rendered[i].repliesCollapsed)
		require.GreaterOrEqualf(t, indicatorCol, 0, "flat[%d] missing ↩ in indicator", i)

		childIndentCols := leadingIndentCols(rendered[i+1].content)
		expectedAuthorCol := layout.CommentSectionLeftMargin + childIndentCols + 1

		assert.Equalf(t, expectedAuthorCol, indicatorCol, "parent depth=%d child depth=%d", flat[i].Depth, flat[i+1].Depth)
	}
}

func TestPrerenderComments_IndentCollapsesOnNarrowTerminal(t *testing.T) {
	t.Parallel()

	flat := linearChain(6)

	// contentWidth = 30 - 2 = 28, commentWidth = min(28, 70) = 28 < MinCommentWidth.
	// Headroom becomes 0, so indent collapses to zero for all depths.
	rc := renderContext{
		commentWidth: 70,
		indent:       5,
		paneWidth:    30,
	}

	rendered := prerenderComments(rc, flat)

	for i := range flat {
		got := leadingIndentCols(rendered[i].content)
		assert.Equalf(t, 0, got, "flat[%d] depth=%d", i, flat[i].Depth)
	}
}

func TestSyncExpandedDepth_MatchesCollapseState(t *testing.T) {
	m := newTestModel(t, testThread())

	m.expandLevel()
	m.expandLevel()

	expected := m.expandedDepth
	m.syncExpandedDepth()
	assert.Equal(t, expected, m.expandedDepth, "sync should match actual state after uniform expand")

	m.toggleMode()
	m.focusedIdx = 0
	m.setCollapsed(true)
	m.toggleMode() // switches back to read, which calls syncExpandedDepth

	// expandedDepth should reflect the deepest uncollapsed-with-children depth + 1.
	for i := range m.flat {
		if m.flat[i].DescendantCount > 0 && !m.flat[i].Collapsed {
			assert.LessOrEqual(t, m.flat[i].Depth, m.expandedDepth-1)
		}
	}
}

// The view hands its meta block the same left margin and column width it
// gives the comments, so the block's frame opens exactly where top-level
// authors start and the block ends one cell inside the separator rule's
// right edge (the block's rightInset). The block's own edge arithmetic is
// meta's TestBlockGeometryContract; this pins the plumbing — one margin, one
// width, shared by the block and the comments under it. The thread's URL is
// longer than any column, so its truncated row reaches the block's edge
// alongside the rules.
func TestMetaBlockAlignsWithCommentColumn(t *testing.T) {
	thread := testThread()
	thread.URL = "https://example.com/" + strings.Repeat("long-path/", 30)
	thread.Domain = "example.com"

	m := newTestModel(t, thread)

	openCol, linkCol, ruleCol, authorCol := -1, -1, -1, -1
	blockEdge, sepWidth := 0, 0

	for line := range strings.SplitSeq(m.Viewport.View(), "\n") {
		// The viewport pads rows to the pane width; the trailing spaces are
		// not part of the rendered content being measured.
		s := strings.TrimRight(xansi.Strip(line), " ")
		trimmed := strings.TrimLeft(s, " ")

		switch {
		case openCol == -1 && strings.HasPrefix(trimmed, "╭"):
			openCol = len(s) - len(trimmed)

			assert.Contains(t, trimmed, "by ", "the opening rule must carry the byline")
		case linkCol == -1 && strings.HasPrefix(trimmed, "│ example.com"):
			linkCol = len(s) - len(trimmed)
			blockEdge = max(blockEdge, xansi.StringWidth(s))
		case strings.HasPrefix(trimmed, "╰"):
			ruleCol = len(s) - len(trimmed)
			blockEdge = max(blockEdge, xansi.StringWidth(s))
		case strings.HasPrefix(trimmed, "▁"):
			sepWidth = xansi.StringWidth(s)
		case authorCol == -1 && strings.HasPrefix(trimmed, "alice"):
			authorCol = len(s) - len(trimmed)
		}
	}

	require.NotEqual(t, -1, openCol, "no meta block opening rule in the view")
	require.NotEqual(t, -1, linkCol, "no meta block URL row in the view")
	require.NotEqual(t, -1, ruleCol, "no closing rule in the view")
	require.NotEqual(t, -1, authorCol, "no top-level comment header in the view")
	require.NotZero(t, sepWidth, "no separator rule in the view")

	assert.Equal(t, layout.CommentSectionLeftMargin, openCol, "the opening rule must open at the comment margin")
	assert.Equal(t, layout.CommentSectionLeftMargin, linkCol, "the URL row must open at the comment margin")
	assert.Equal(t, layout.CommentSectionLeftMargin, ruleCol, "the closing rule must open at the comment margin")
	assert.Equal(t, layout.CommentSectionLeftMargin, authorCol, "top-level authors must start at the comment margin")
	assert.Equal(t, sepWidth-1, blockEdge, "the block must end one cell inside the separator's right edge")
}

func TestFocusDecoration_AppliedAtViewTime(t *testing.T) {
	m := newTestModel(t, testThread())
	m.toggleMode()

	raw := m.Viewport.View()
	focused := m.prerendered[m.visible[m.focusedIdx]].headerFocused[0]

	assert.NotContains(t, raw, focused, "the focused variant is not baked into the content")
	assert.Contains(t, m.DecorateView(raw), focused, "the focused variant appears at display time")
}

func TestResize_HeightOnlySkipsPrerender(t *testing.T) {
	m := newTestModel(t, testThread())

	before := &m.prerendered[0]

	m.Update(tea.WindowSizeMsg{Width: 120, Height: 60})
	assert.Same(t, before, &m.prerendered[0], "height-only resize must not re-prerender")
	assert.Equal(t, m.ContentLines+m.Viewport.Height(), m.Viewport.TotalLineCount(),
		"the bottom padding tracks the new height")

	m.Update(tea.WindowSizeMsg{Width: 100, Height: 60})
	assert.NotSame(t, before, &m.prerendered[0], "a width change re-prerenders")
}

func TestResize_HeightOnlyKeepsSearchMatches(t *testing.T) {
	m := searchModel(t)
	commitCommentSearch(m, "needle")

	before := &m.searchMatches[0]

	m.Update(tea.WindowSizeMsg{Width: 120, Height: 60})

	assert.Same(t, before, &m.searchMatches[0], "height-only resize must not recompute matches")
	assert.Contains(t, m.DecorateView(m.Viewport.View()), ansi.Reverse)
}

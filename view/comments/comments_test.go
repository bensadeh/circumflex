package comments

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/layout"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestNew_StartsInReadMode(t *testing.T) {
	m := newTestModel(t, testThread())

	assert.Equal(t, modeRead, m.mode)
	assert.Equal(t, -1, m.focusedIdx, "no focus in read mode")
}

func TestNew_StartsFullyCollapsed(t *testing.T) {
	m := newTestModel(t, testThread())

	assert.Equal(t, 0, m.expandedDepth)

	// Only top-level comments should be visible.
	for _, vi := range m.visible {
		assert.Equal(t, 0, m.flat[vi].Depth)
	}
}

func TestExpandLevel_RevealsChildren(t *testing.T) {
	m := newTestModel(t, testThread())

	// Initially only depth-0 visible.
	initialCount := len(m.visible)

	m.expandLevel()
	assert.Equal(t, 1, m.expandedDepth)
	assert.Greater(t, len(m.visible), initialCount, "expanding should reveal more comments")

	// Depth-1 comments should now be visible.
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

	// Expand all the way.
	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	assert.Len(t, m.visible, len(m.flat), "fully expanded should show all comments")
}

func TestCollapseLevel_HidesChildren(t *testing.T) {
	m := newTestModel(t, testThread())

	// Expand then collapse back.
	m.expandLevel()
	expanded := len(m.visible)

	m.collapseLevel()
	assert.Less(t, len(m.visible), expanded)
	assert.Equal(t, 0, m.expandedDepth)
}

func TestCollapseLevel_ClampsAtZero(t *testing.T) {
	m := newTestModel(t, testThread())

	m.collapseLevel() // already at 0
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

	// First toggle expands all.
	m.toggleCollapseAll()
	assert.Len(t, m.visible, len(m.flat))

	// Second toggle collapses all.
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

	m.toggleMode() // read -> navigate
	m.toggleMode() // navigate -> read

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

	// Move backward from 0.
	m.focusedIdx = 0
	m.navigateComment(-1)
	assert.Equal(t, 0, m.focusedIdx, "should not go below 0")

	// Move forward past end.
	m.focusedIdx = len(m.visible) - 1
	m.navigateComment(1)
	assert.Equal(t, len(m.visible)-1, m.focusedIdx, "should not exceed visible length")
}

func TestSetCollapsed_CollapsesAndExpands(t *testing.T) {
	m := newTestModel(t, testThread())

	// Fully expand so we can test individual collapse.
	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()

	// Focus on the first comment (alice, has descendants).
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

	// Fully expand.
	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()

	// Find a leaf comment (no descendants).
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

	// Fully expand.
	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()

	// Focus on first comment with descendants.
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

func TestViewportStable_ExpandLevel(t *testing.T) {
	m := newScrollableModel(t)

	// Expand once so we have content, then scroll partway down.
	m.expandLevel()
	m.viewport.SetYOffset(m.contentLines / 3)

	anchor := m.anchorComment()
	posBefore := m.screenPosition(anchor)

	m.expandLevel()

	posAfter := m.screenPosition(anchor)
	assert.Equal(t, posBefore, posAfter,
		"anchor comment should not move on screen after expanding a level")
}

func TestViewportStable_CollapseLevel(t *testing.T) {
	m := newScrollableModel(t)

	// Fully expand, scroll down, then collapse one level.
	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.viewport.SetYOffset(m.contentLines / 3)

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

	// Fully expand, enter navigate mode, scroll down, focus a comment with children.
	for range m.maxDepth + 1 {
		m.expandLevel()
	}

	m.toggleMode()
	m.viewport.SetYOffset(m.contentLines / 3)

	// Find a visible comment with descendants below the current scroll.
	found := false

	for vi, fi := range m.visible {
		if m.flat[fi].DescendantCount > 0 && m.lineMetrics[fi].StartLine >= m.viewport.YOffset() {
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
	m.viewport.SetYOffset(m.contentLines / 3)

	anchor := m.anchorComment()
	posBefore := m.screenPosition(anchor)

	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	posAfter := m.screenPosition(anchor)
	assert.Equal(t, posBefore, posAfter,
		"anchor comment should not move on screen after resize")
}

// linearChain builds a thread of a single depth-N chain where comment i has
// comment i+1 as its only child.
func linearChain(depth int) []flatComment {
	var children []*comment.Comment
	for i := depth; i >= 1; i-- {
		c := newComment(i, "u", "body", children...)
		children = []*comment.Comment{c}
	}

	return flatten(newThread(children...))
}

func leadingIndentCols(s string) int {
	return len(s) - len(strings.TrimLeft(s, " ")) - layout.CommentSectionLeftMargin
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
		screenWidth:  120,
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
func indicatorLeadingCols(s string) int {
	for line := range strings.SplitSeq(s, "\n") {
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
		screenWidth:  120,
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
		screenWidth:  30,
	}

	rendered := prerenderComments(rc, flat)

	for i := range flat {
		got := leadingIndentCols(rendered[i].content)
		assert.Equalf(t, 0, got, "flat[%d] depth=%d", i, flat[i].Depth)
	}
}

func TestSyncExpandedDepth_MatchesCollapseState(t *testing.T) {
	m := newTestModel(t, testThread())

	// Expand to level 2, then individually collapse something in navigate mode.
	m.expandLevel()
	m.expandLevel()

	expected := m.expandedDepth
	m.syncExpandedDepth()
	assert.Equal(t, expected, m.expandedDepth, "sync should match actual state after uniform expand")

	// Now individually collapse a comment to create a non-uniform state.
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

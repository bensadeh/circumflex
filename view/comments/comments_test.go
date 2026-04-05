package comments

import (
	"testing"

	"github.com/bensadeh/circumflex/comment"

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

	return New(thread, 0, 80, false, 120, 200)
}

// --- Initial state ---

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

// --- Expand / collapse levels (read mode) ---

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

// --- Toggle collapse all (read mode) ---

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

// --- Mode switching ---

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

// --- Navigate mode: comment-by-comment ---

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

// --- Navigate mode: individual collapse/expand ---

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

// --- Goto top/bottom ---

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

// --- syncExpandedDepth ---

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

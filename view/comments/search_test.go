package comments

import (
	"testing"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// searchModel builds a thread where both "needle" matches start out hidden:
//
//	A "top level alpha"            (flat 0, collapsed)
//	  B "first needle reply"       (flat 1, hidden, collapsed)
//	    C "second needle deeper"   (flat 2, hidden)
//	D "unrelated closing comment"  (flat 3)
func searchModel(t *testing.T) *Model {
	t.Helper()

	thread := newThread(
		newComment(1, "alice", "top level alpha",
			newComment(2, "bob", "first needle reply",
				newComment(3, "carol", "second needle deeper"),
			),
		),
		newComment(4, "dave", "unrelated closing comment"),
	)

	return newTestModel(t, thread)
}

func commitCommentSearch(m *Model, query string) {
	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})

	for _, r := range query {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
}

func TestCommentSearch_ExpandsAllOnPromptOpen(t *testing.T) {
	m := searchModel(t)

	require.Equal(t, 0, m.lineMetrics[1].LineCount, "needle comments start hidden")

	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})

	for i := range m.flat {
		assert.Positive(t, m.lineMetrics[i].LineCount, "comment %d is visible", i)
	}

	assert.Equal(t, m.maxDepth, m.expandedDepth, "the depth indicator reflects the full expansion")

	for _, r := range "needle" {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.Len(t, m.searchMatches, 2)
	assert.Equal(t, 0, m.searchCurrent)

	line, visible := m.matchLine(m.searchMatches[0])
	require.True(t, visible)
	assert.Equal(t, max(0, line-scrollPadding), m.Viewport.YOffset())
}

func TestCommentSearch_CanceledPromptKeepsExpansion(t *testing.T) {
	m := searchModel(t)

	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	assert.False(t, m.SearchPrompting())
	assert.Positive(t, m.lineMetrics[1].LineCount, "the expansion outlives the canceled prompt")
}

func TestCommentSearch_JumpRevealsRecollapsedBranch(t *testing.T) {
	m := searchModel(t)
	commitCommentSearch(m, "needle")

	m.collapseAll()
	require.Equal(t, 0, m.lineMetrics[1].LineCount)

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

	assert.Equal(t, 1, m.searchCurrent)
	assert.Positive(t, m.lineMetrics[2].LineCount, "n reveals a match the user collapsed away")
}

func TestCommentSearch_HighlightsAndCounter(t *testing.T) {
	m := searchModel(t)
	commitCommentSearch(m, "needle")

	assert.Contains(t, m.DecorateView(m.Viewport.View()), ansi.Reverse)
	assert.Contains(t, ansi.Strip(m.modeIndicator()), "1/2")

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	assert.Contains(t, ansi.Strip(m.modeIndicator()), "2/2")

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	assert.Contains(t, ansi.Strip(m.modeIndicator()), "1/2", "n wraps around")

	m.Update(tea.KeyPressMsg{Code: 'N', Text: "N"})
	assert.Contains(t, ansi.Strip(m.modeIndicator()), "2/2", "N wraps back")
}

func TestCommentSearch_EscClearsThenQuits(t *testing.T) {
	m := searchModel(t)
	commitCommentSearch(m, "needle")

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	assert.Nil(t, cmd)
	assert.False(t, m.SearchActive(), "the first esc only clears the search")
	assert.Empty(t, m.searchMatches)
	assert.NotContains(t, m.DecorateView(m.Viewport.View()), ansi.Reverse)
	assert.Positive(t, m.lineMetrics[1].LineCount, "revealed comments stay revealed")

	cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	require.NotNil(t, cmd)
	assert.IsType(t, message.DetailQuit{}, cmd(), "the second esc quits")
}

func TestCommentSearch_NKeepsTopLevelJumpWhenInactive(t *testing.T) {
	m := searchModel(t)

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

	assert.Equal(t, m.lineMetrics[0].StartLine, m.Viewport.YOffset(),
		"n without a search keeps its top-level-jump meaning")
}

func TestCommentSearch_SurvivesResize(t *testing.T) {
	m := searchModel(t)
	commitCommentSearch(m, "needle")

	m.Update(tea.WindowSizeMsg{Width: 90, Height: 40})

	assert.True(t, m.SearchActive())
	assert.Len(t, m.searchMatches, 2, "matches are recomputed against the rewrapped comments")
	assert.Contains(t, m.DecorateView(m.Viewport.View()), ansi.Reverse)
}

func TestCommentSearch_FindsRootSelfText(t *testing.T) {
	thread := newThread(newComment(1, "alice", "body"))
	thread.Content = "story self text with a needle inside"
	m := newTestModel(t, thread)

	commitCommentSearch(m, "self text")

	require.Len(t, m.searchMatches, 1)
	assert.Equal(t, -1, m.searchMatches[0].flatIdx, "the match sits in the thread header")
	assert.Contains(t, m.DecorateView(m.Viewport.View()), ansi.Reverse)
}

func TestCommentSearch_NavModeFocusFollowsMatch(t *testing.T) {
	m := searchModel(t)
	m.toggleMode()

	commitCommentSearch(m, "needle")

	require.GreaterOrEqual(t, m.focusedIdx, 0)
	assert.Equal(t, 1, m.visible[m.focusedIdx], "focus lands on the matched comment")
}

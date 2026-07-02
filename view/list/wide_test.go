package list

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wideTestWidth  = 250
	wideTestHeight = 30
)

// newWideTestModel returns a browsing model sized past the wide-layout
// threshold, with items loaded.
func newWideTestModel(t *testing.T) *Model {
	t.Helper()
	m := newTestModel(t)

	m, _ = m.Update(tea.WindowSizeMsg{Width: wideTestWidth, Height: wideTestHeight})

	m.pager.items[categories.Top] = testItems()
	m.status.StopSpinner()
	m.state = stateBrowsing
	m.updatePagination()

	return m
}

// openTestComments drives the model through the full open-story flow: Enter,
// then the fetched comment tree arriving.
func openTestComments(t *testing.T, m *Model) {
	t.Helper()

	m, _ = m.Update(keyMsg("enter"))
	require.Equal(t, stateFetching, m.state)

	thread := comment.ToThread(&hn.CommentTree{ID: 1, Title: "First item", CommentsCount: 5})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})
	require.Equal(t, stateCommentView, m.state)
}

func TestWideView_Threshold(t *testing.T) {
	m := newWideTestModel(t)

	m.setSize(239, wideTestHeight)
	assert.False(t, m.isWide())
	assert.Equal(t, 239, m.listWidth())
	assert.Equal(t, 239, m.detailWidth())

	m.setSize(240, wideTestHeight)
	assert.True(t, m.isWide())
	assert.Equal(t, (240-dividerWidth)/2, m.listWidth())
	assert.Equal(t, 240-m.listWidth()-dividerWidth, m.detailWidth())
}

func TestWideView_ConfiguredThreshold(t *testing.T) {
	m := newWideTestModel(t)

	m.config.WideViewMinWidth, _ = settings.ParseWideView("never")
	assert.False(t, m.isWide(), "never should keep the full-screen layout at any width")

	m.config.WideViewMinWidth, _ = settings.ParseWideView("always")
	assert.True(t, m.isWide(), "always should split on any reasonably sized terminal")

	m.setSize(wideViewFloor-1, wideTestHeight)
	assert.False(t, m.isWide(), "always should still not split below the sanity floor")

	m.setSize(120, wideTestHeight)
	m.config.WideViewMinWidth, _ = settings.ParseWideView("120")
	assert.True(t, m.isWide())

	m.config.WideViewMinWidth, _ = settings.ParseWideView("121")
	assert.False(t, m.isWide())
}

func TestWideView_ShowsPlaceholderWhileBrowsing(t *testing.T) {
	m := newWideTestModel(t)

	view := m.View()

	assert.Contains(t, view, "Select a story")
	assert.Contains(t, view, "│")
	assert.Contains(t, view, "First item")

	lines := strings.Split(view, "\n")
	assert.Len(t, lines, wideTestHeight)

	for i, line := range lines {
		assert.Equal(t, wideTestWidth, xansi.StringWidth(line), "line %d should span the full terminal width", i)
	}

	// The placeholder pane draws the same header rule the detail views do, so
	// both panes carry the rule even before a story opens.
	assert.Equal(t, wideTestWidth-dividerWidth, strings.Count(xansi.Strip(lines[1]), "‾"),
		"header rule should span both panes")
}

func TestWideView_LeftPaneDimsOnceWhileStoryIsOpen(t *testing.T) {
	m := newWideTestModel(t)

	browsing := m.listView()

	m, _ = m.Update(keyMsg("enter"))
	require.Equal(t, stateFetching, m.state)

	loading := m.listView()
	assert.NotEqual(t, browsing, loading, "left pane should dim when the story starts loading")

	thread := comment.ToThread(&hn.CommentTree{ID: 1, Title: "First item", CommentsCount: 5})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})
	require.Equal(t, stateCommentView, m.state)
	assert.Equal(t, loading, m.listView(), "left pane should not change again when the story arrives")

	m, _ = m.Update(message.CommentViewQuit{})
	require.Equal(t, stateBrowsing, m.state)
	assert.Equal(t, browsing, m.listView(), "left pane should restore when the story closes")
}

func TestWideView_OpenStoryShowsReadingMarker(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)

	assert.True(t, m.dimList())

	// The open story renders faint and bold — dimmed with the list but
	// marking the J/K reading position.
	var open strings.Builder

	m.renderItem(&open, m.Index(), m.SelectedItem())
	assert.NotContains(t, open.String(), "\x1b[7", "open story should not use the browsing highlight")
	assert.Contains(t, open.String(), "\x1b[1;2m", "open story should render bold and faint")

	var other strings.Builder

	m.renderItem(&other, m.Index()+1, m.VisibleItems()[m.Index()+1])
	assert.Contains(t, other.String(), "\x1b[3;2m", "other stories should dim")
}

// openAdjacent presses J or K in the open detail view and delivers the
// resulting OpenAdjacentStory message back to the list.
func openAdjacent(t *testing.T, m *Model, key string) tea.Cmd {
	t.Helper()

	m, cmd := m.Update(keyMsg(key))
	require.NotNil(t, cmd, "%s should emit a command", key)

	_, cmd = m.Update(cmd())

	return cmd
}

func TestWideView_AdjacentStoryNavigationFlipsPages(t *testing.T) {
	m := newTestModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: wideTestWidth, Height: 9})
	m.pager.items[categories.Top] = testItems()
	m.status.StopSpinner()
	m.state = stateBrowsing
	m.updatePagination()
	require.Equal(t, 2, m.pager.Paginator.PerPage)

	openTestComments(t, m)
	require.Equal(t, 0, m.Index())

	// J: second story on the same page.
	openAdjacent(t, m, "J")
	require.Equal(t, stateFetching, m.state)
	assert.Equal(t, 1, m.Index())
	assert.Equal(t, 0, m.pager.Paginator.Page)

	thread := comment.ToThread(&hn.CommentTree{ID: 2, Title: "Second item", CommentsCount: 3})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})
	require.Equal(t, stateCommentView, m.state)

	// J again: third story, which lives on the next page.
	openAdjacent(t, m, "J")
	assert.Equal(t, 2, m.Index())
	assert.Equal(t, 1, m.pager.Paginator.Page)
	assert.Equal(t, 0, m.pager.cursor)

	thread = comment.ToThread(&hn.CommentTree{ID: 3, Title: "Third item", CommentsCount: 1})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})

	// K: back to the second story, flipping back to the first page.
	openAdjacent(t, m, "K")
	assert.Equal(t, 1, m.Index())
	assert.Equal(t, 0, m.pager.Paginator.Page)
	assert.Equal(t, 1, m.pager.cursor)
}

func TestWideView_AdjacentStoryNavigationStopsAtEdges(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)
	require.Equal(t, 0, m.Index())

	// K at the first story: nothing happens, the view stays open.
	cmd := openAdjacent(t, m, "K")
	assert.Nil(t, cmd)
	assert.Equal(t, 0, m.Index())
	assert.Equal(t, stateCommentView, m.state)
}

func TestWideView_CommentSectionFillsDetailPane(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)

	view := m.View()

	assert.NotContains(t, view, "Select a story")
	assert.Contains(t, xansi.Strip(view), "5 comments")

	lines := strings.Split(view, "\n")
	assert.Len(t, lines, wideTestHeight)

	for i, line := range lines {
		assert.Equal(t, wideTestWidth, xansi.StringWidth(line), "line %d should span the full terminal width", i)
	}
}

func TestWideView_QuitRestoresPlaceholder(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)

	// q inside the comment view emits CommentViewQuit as a command; deliver it.
	m, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	m, _ = m.Update(cmd())

	assert.Equal(t, stateBrowsing, m.state)
	assert.Nil(t, m.commentView)
	assert.Contains(t, m.View(), "Select a story")
}

func TestWideView_NarrowBehaviorUnchanged(t *testing.T) {
	m := newTestModelReady(t)
	openTestComments(t, m)

	// Full-screen comment section: no divider column, no list content.
	view := m.View()
	assert.NotContains(t, view, "Select a story")
	assert.NotContains(t, view, "Second item")
}

func TestWideView_ResizeAcrossThreshold(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)

	// Shrink below the threshold: the comment section takes the full screen.
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: wideTestHeight})
	assert.False(t, m.isWide())
	assert.NotContains(t, m.View(), "Second item")

	// Widen again: the list returns next to the comment section.
	m, _ = m.Update(tea.WindowSizeMsg{Width: wideTestWidth, Height: wideTestHeight})
	assert.True(t, m.isWide())
	assert.Contains(t, m.View(), "Second item")
}

func TestPaneLines_NormalizesWidthAndHeight(t *testing.T) {
	lines := paneLines("short\n"+strings.Repeat("x", 20), 10, 4)

	require.Len(t, lines, 4)
	assert.Equal(t, "short     ", lines[0])
	assert.Equal(t, strings.Repeat("x", 10), lines[1])
	assert.Equal(t, strings.Repeat(" ", 10), lines[2])
	assert.Equal(t, strings.Repeat(" ", 10), lines[3])
}

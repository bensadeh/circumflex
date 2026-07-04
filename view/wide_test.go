package view

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

func newWideTestModel(t *testing.T) *model {
	t.Helper()
	m := newTestModel(t)

	m, _ = m.Update(tea.WindowSizeMsg{Width: wideTestWidth, Height: wideTestHeight})

	m.list.SetItems(categories.Top, testItems())
	m.status.StopSpinner()
	m.fetching = false
	m.updatePagination()

	return m
}

func openTestComments(t *testing.T, m *model) {
	t.Helper()

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetching)

	thread := comment.ToThread(&hn.CommentTree{ID: 1, Title: "First item", CommentsCount: 5})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})
	require.Equal(t, screenComments, m.screen)
}

func TestWideView_Threshold(t *testing.T) {
	m := newWideTestModel(t)

	m.setSize(179, wideTestHeight)
	assert.False(t, m.isWide())
	assert.Equal(t, 179, m.listWidth())
	assert.Equal(t, 179, m.detailWidth())

	m.setSize(180, wideTestHeight)
	assert.True(t, m.isWide())
	assert.Equal(t, (180-dividerWidth)/2, m.listWidth())
	assert.Equal(t, 180-m.listWidth()-dividerWidth, m.detailWidth())
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

	browsing := m.browsingView()

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetching)

	loading := m.browsingView()
	assert.NotEqual(t, browsing, loading, "left pane should dim when the story starts loading")

	thread := comment.ToThread(&hn.CommentTree{ID: 1, Title: "First item", CommentsCount: 5})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})
	require.Equal(t, screenComments, m.screen)
	assert.Equal(t, loading, m.browsingView(), "left pane should not change again when the story arrives")

	m, _ = m.Update(message.CommentViewQuit{})
	require.Equal(t, screenList, m.screen)
	assert.Equal(t, browsing, m.browsingView(), "left pane should restore when the story closes")
}

// openAdjacent presses J or K in the open detail view and delivers the
// resulting OpenAdjacentStory message back to the coordinator.
func openAdjacent(t *testing.T, m *model, key string) tea.Cmd {
	t.Helper()

	m, cmd := m.Update(keyMsg(key))
	require.NotNil(t, cmd, "%s should emit a command", key)

	_, cmd = m.Update(cmd())

	return cmd
}

func TestWideView_AdjacentStoryNavigationFlipsPages(t *testing.T) {
	m := newTestModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: wideTestWidth, Height: 9})
	m.list.SetItems(categories.Top, testItems())
	m.status.StopSpinner()
	m.fetching = false
	m.updatePagination()
	require.Equal(t, 2, m.list.PerPage())

	openTestComments(t, m)
	require.Equal(t, 0, m.list.Index())

	openAdjacent(t, m, "J")
	require.True(t, m.fetching)
	assert.Equal(t, 1, m.list.Index())
	assert.Equal(t, 0, m.list.Page())

	thread := comment.ToThread(&hn.CommentTree{ID: 2, Title: "Second item", CommentsCount: 3})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})
	require.Equal(t, screenComments, m.screen)

	openAdjacent(t, m, "J")
	assert.Equal(t, 2, m.list.Index())
	assert.Equal(t, 1, m.list.Page())
	assert.Equal(t, 0, m.list.Cursor())

	thread = comment.ToThread(&hn.CommentTree{ID: 3, Title: "Third item", CommentsCount: 1})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetchID})

	openAdjacent(t, m, "K")
	assert.Equal(t, 1, m.list.Index())
	assert.Equal(t, 0, m.list.Page())
	assert.Equal(t, 1, m.list.Cursor())
}

func TestWideView_AdjacentStoryNavigationStopsAtEdges(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)
	require.Equal(t, 0, m.list.Index())

	cmd := openAdjacent(t, m, "K")
	assert.Nil(t, cmd)
	assert.Equal(t, 0, m.list.Index())
	assert.Equal(t, screenComments, m.screen)
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

func TestWideView_HelpShowsInDetailPane(t *testing.T) {
	m := newWideTestModel(t)

	browsing := m.browsingView()

	m, _ = m.Update(keyMsg("i"))
	require.Equal(t, screenHelp, m.screen)

	view := m.View()
	assert.Contains(t, xansi.Strip(view), "Keyboard Shortcuts")
	assert.Contains(t, view, "First item", "the story list should stay visible next to help")
	assert.NotEqual(t, browsing, m.browsingView(), "left pane should dim while help is open")

	lines := strings.Split(view, "\n")
	assert.Len(t, lines, wideTestHeight)

	for i, line := range lines {
		assert.Equal(t, wideTestWidth, xansi.StringWidth(line), "line %d should span the full terminal width", i)
	}

	m, _ = m.Update(keyMsg("q"))
	assert.Equal(t, screenList, m.screen)
	assert.Contains(t, m.View(), "Select a story")
	assert.Equal(t, browsing, m.browsingView(), "left pane should restore when help closes")
}

func TestWideView_QuitRestoresPlaceholder(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)

	// q inside the comment view emits CommentViewQuit as a command; deliver it.
	m, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	m, _ = m.Update(cmd())

	assert.Equal(t, screenList, m.screen)
	assert.Nil(t, m.commentView)
	assert.Contains(t, m.View(), "Select a story")
}

func TestWideView_NarrowBehaviorUnchanged(t *testing.T) {
	m := newTestModelReady(t)
	openTestComments(t, m)

	view := m.View()
	assert.NotContains(t, view, "Select a story")
	assert.NotContains(t, view, "Second item")
}

func TestWideView_ResizeAcrossThreshold(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: wideTestHeight})
	assert.False(t, m.isWide())
	assert.NotContains(t, m.View(), "Second item")

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

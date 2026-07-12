package view

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/layout"
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
	m, _ = m.Update(message.StoriesReady{Stories: testItems(), Category: categories.Top, FetchID: m.fetch.id})

	return m
}

func openTestComments(t *testing.T, m *model) {
	t.Helper()

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())

	thread := comment.ToThread(&hn.CommentTree{ID: 1, Title: "First item", CommentsCount: 5})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.id})
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
	assert.Equal(t, (180-layout.PaneDividerWidth)/2, m.listWidth())
	assert.Equal(t, 180-m.listWidth()-layout.PaneDividerWidth, m.detailWidth())
}

func TestWideView_ConfiguredThreshold(t *testing.T) {
	m := newWideTestModel(t)

	m.config.WideViewMinWidth, _ = settings.ParseWideView("never")
	assert.False(t, m.isWide(), "never should keep the full-screen layout at any width")

	m.config.WideViewMinWidth, _ = settings.ParseWideView("always")
	assert.True(t, m.isWide(), "always should split on any reasonably sized terminal")

	m.setSize(layout.WideViewFloor-1, wideTestHeight)
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
	assert.Equal(t, wideTestWidth-layout.PaneDividerWidth, strings.Count(xansi.Strip(lines[1]), "‾"),
		"header rule should span both panes")
}

func TestWideView_LeftPaneDimsOnceWhileStoryIsOpen(t *testing.T) {
	m := newWideTestModel(t)

	browsing := m.browsingView()

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())

	loading := m.browsingView()
	assert.NotEqual(t, browsing, loading, "left pane should dim when the story starts loading")

	thread := comment.ToThread(&hn.CommentTree{ID: 1, Title: "First item", CommentsCount: 5})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.id})
	require.Equal(t, screenComments, m.screen)
	assert.Equal(t, loading, m.browsingView(), "left pane should not change again when the story arrives")

	m, _ = m.Update(message.DetailQuit{})
	require.Equal(t, screenList, m.screen)
	assert.Equal(t, browsing, m.browsingView(), "left pane should restore when the story closes")
}

// A story fetch heads the detail pane with the incoming story's title right
// away — unbolded until the content arrives and the full header takes over.
func TestWideView_LoadingShowsUnboldedTitle(t *testing.T) {
	m := newWideTestModel(t)

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())

	loading := m.detailPaneView()
	assert.Contains(t, xansi.Strip(loading), "First item")
	assert.NotContains(t, loading, "\x1b[1m", "loading title must not be bold")

	thread := comment.ToThread(&hn.CommentTree{ID: 1, Title: "First item", CommentsCount: 5})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.id})
	assert.Contains(t, m.detailPaneView(), "\x1b[1m", "the opened story's title regains its bold")
}

// While a story loads, the pane reserves the meta block's spot: the block's
// empty dimmed frame, spanning the same rows as the loaded block, so the
// block neither moves nor resizes when the content arrives.
func TestWideView_LoadingShowsMetaBlockPlaceholder(t *testing.T) {
	m := newWideTestModel(t)

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())

	loading := m.detailPaneView()
	loadingBox := metaBoxLines(t, loading)
	assert.Contains(t, loadingBox[len(loadingBox)-1], "\x1b[2m", "the placeholder's closing rule must render dimmed")

	frameRunes := strings.NewReplacer("╭", "", "╮", "", "╰", "", "╯", "", "│", "", "─", "", " ", "")
	for i, line := range loadingBox[:len(loadingBox)-1] {
		assert.Empty(t, frameRunes.Replace(strings.TrimSpace(xansi.Strip(line))), "placeholder row %d must hold no text", i)
	}

	thread := comment.ToThread(&hn.CommentTree{
		ID: 1, Title: "First item", CommentsCount: 5,
		URL: "https://example.com/story", Domain: "example.com",
	})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.id})

	loadedBox := metaBoxLines(t, m.detailPaneView())
	assert.Len(t, loadingBox, len(loadedBox), "placeholder must span the same rows as the loaded meta block")
}

// The placeholder survives a failed load: the error view keeps it in place
// instead of flashing it away with the loading pane.
func TestWideView_ErrorViewKeepsMetaBlockPlaceholder(t *testing.T) {
	m := newWideTestModel(t)

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())

	loadingBox := metaBoxLines(t, m.detailPaneView())

	m, _ = m.Update(message.CommentTreeDataReady{Err: errors.New("server returned status 403"), FetchID: m.fetch.id})
	require.Equal(t, screenComments, m.screen)

	assert.Equal(t, loadingBox, metaBoxLines(t, m.detailPaneView()),
		"the placeholder must not move or change when the load fails")
}

// metaBoxLines returns the view's run of meta block rows: everything between
// the pane's header rule and the block's closing rule. The block sits
// directly under the header in the loading, loaded, and error panes alike,
// so the slice is the block's spot whether it holds the skeleton or the
// loaded content.
func metaBoxLines(t *testing.T, view string) []string {
	t.Helper()

	lines := strings.Split(view, "\n")

	top := slices.IndexFunc(lines, func(l string) bool {
		return strings.Contains(xansi.Strip(l), "‾")
	})
	require.GreaterOrEqual(t, top, 0, "no pane header rule in view")

	rulePrefix := strings.Repeat(" ", layout.CommentSectionLeftMargin) + "╰"
	bottom := slices.IndexFunc(lines, func(l string) bool {
		return strings.HasPrefix(xansi.Strip(l), rulePrefix)
	})
	require.Greater(t, bottom, top, "no meta block closing rule in view")

	return lines[top+1 : bottom+1]
}

// A story load that fails swaps the detail pane to an error view — not a
// status-bar message. It counts as an open view, exactly like the narrow
// layout keeping its story open: the reading marker sits on the story that
// failed, scroll keys have nothing to act on, and J/K page on to the
// neighboring stories.
func TestWideView_StoryLoadErrorBecomesView(t *testing.T) {
	var progress strings.Builder

	progressOut = &progress

	t.Cleanup(func() { progressOut = io.Discard })

	m := newWideTestModel(t)
	openTestComments(t, m)

	openAdjacent(t, m, "J")
	require.True(t, m.fetch.inFlight())
	require.Equal(t, 1, m.list.Index())

	m, _ = m.Update(message.CommentTreeDataReady{Err: errors.New("boom"), FetchID: m.fetch.id})

	require.IsType(t, &errorView{}, m.detail, "the error should take the pane as a view of its own")
	assert.Equal(t, screenComments, m.screen)
	assert.Equal(t, 1, m.list.Index(), "the reading marker stays on the story that failed")

	errorPane := m.detailPaneView()
	assert.Contains(t, xansi.Strip(errorPane), "Boom")
	assert.Contains(t, xansi.Strip(errorPane), "Second item", "the failed story's title heads the pane")
	assert.NotContains(t, errorPane, "\x1b[1m", "the unopened story's title stays unbolded")
	assert.Empty(t, m.status.message, "wide layout errors bypass the status bar")
	assert.Contains(t, progress.String(), "\x1b]9;4;2;100\a", "the terminal progress indicator should show the error")

	// Scroll keys have nothing to scroll in an error view; nothing changes
	// and, crucially, nothing hidden receives the key.
	m, _ = m.Update(keyMsg("j"))
	assert.Contains(t, xansi.Strip(m.detailPaneView()), "Boom")
	assert.Equal(t, 1, m.list.Index())

	// J pages on to the next story, like from any open view.
	openAdjacent(t, m, "J")
	require.True(t, m.fetch.inFlight())
	assert.Equal(t, 2, m.list.Index())

	thread := comment.ToThread(&hn.CommentTree{ID: 3, Title: "Third item", CommentsCount: 3})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.id})
	require.Equal(t, screenComments, m.screen)
	assert.Contains(t, xansi.Strip(m.detailPaneView()), "3 comments")
}

// The terminal progress indicator settles on its own a few seconds after the
// failure, like a status message expiring; the error view itself stays up. A
// stale timeout from an older fetch must not touch a newer fetch's indicator.
func TestWideView_ErrorProgressTimeout(t *testing.T) {
	var progress strings.Builder

	progressOut = &progress

	t.Cleanup(func() { progressOut = io.Discard })

	m := newWideTestModel(t)
	openTestComments(t, m)

	openAdjacent(t, m, "J")
	m, _ = m.Update(message.CommentTreeDataReady{Err: errors.New("boom"), FetchID: m.fetch.id})

	require.Contains(t, progress.String(), "\x1b]9;4;2;100\a")

	m, _ = m.Update(message.ErrorProgressTimeout{FetchID: m.fetch.id - 1})

	assert.False(t, strings.HasSuffix(progress.String(), "\x1b]9;4;0\a"),
		"a stale timeout must not clear the indicator")

	m, _ = m.Update(message.ErrorProgressTimeout{FetchID: m.fetch.id})

	assert.True(t, strings.HasSuffix(progress.String(), "\x1b]9;4;0\a"),
		"the timeout should clear the indicator")
	require.IsType(t, &errorView{}, m.detail, "the error view itself stays up")
}

// Quitting the error view returns to browsing and settles the terminal
// progress indicator that held its error state while the view was up.
func TestWideView_ErrorViewQuit(t *testing.T) {
	var progress strings.Builder

	progressOut = &progress

	t.Cleanup(func() { progressOut = io.Discard })

	m := newWideTestModel(t)
	openTestComments(t, m)

	openAdjacent(t, m, "J")
	m, _ = m.Update(message.CommentTreeDataReady{Err: errors.New("boom"), FetchID: m.fetch.id})
	require.IsType(t, &errorView{}, m.detail)

	m, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd, "q in the error view should emit its quit message")
	m, _ = m.Update(cmd())

	assert.Nil(t, m.detail)
	assert.Equal(t, screenList, m.screen)
	assert.Contains(t, m.View(), "Select a story")
	assert.True(t, strings.HasSuffix(progress.String(), "\x1b]9;4;0\a"),
		"closing the error view should clear the terminal progress indicator")
}

// The same failure in the narrow layout keeps the outgoing story on screen —
// the error only takes a status row there — so the selection moves back to
// the story the screen still shows.
func TestNarrowStoryLoadError_KeepsOpenStoryAndSelection(t *testing.T) {
	m := newTestModelReady(t)
	openTestComments(t, m)

	openAdjacent(t, m, "J")
	require.True(t, m.fetch.inFlight())
	require.Equal(t, 1, m.list.Index())

	m, _ = m.Update(message.CommentTreeDataReady{Err: errors.New("boom"), FetchID: m.fetch.id})

	assert.NotNil(t, m.detail, "narrow layout keeps the outgoing story open")
	assert.Equal(t, 0, m.list.Index(), "the selection moves back to the story still on screen")
	assert.Contains(t, m.status.message, "Boom")
}

// Cancelling a J/K story fetch keeps the open story, so the selection must
// move back to it as well.
func TestWideView_CancelledStoryLoadRestoresSelection(t *testing.T) {
	m := newWideTestModel(t)
	openTestComments(t, m)

	openAdjacent(t, m, "J")
	require.True(t, m.fetch.inFlight())
	require.Equal(t, 1, m.list.Index())

	m, _ = m.Update(keyMsg("esc"))
	assert.False(t, m.fetch.inFlight())
	assert.Equal(t, 0, m.list.Index(), "cancel should move the selection back to the open story")
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
	m, _ = m.Update(message.StoriesReady{Stories: testItems(), Category: categories.Top, FetchID: m.fetch.id})
	require.Equal(t, 2, m.list.PerPage())

	openTestComments(t, m)
	require.Equal(t, 0, m.list.Index())

	openAdjacent(t, m, "J")
	require.True(t, m.fetch.inFlight())
	assert.Equal(t, 1, m.list.Index())
	assert.Equal(t, 0, m.list.Page())

	thread := comment.ToThread(&hn.CommentTree{ID: 2, Title: "Second item", CommentsCount: 3})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.id})
	require.Equal(t, screenComments, m.screen)

	openAdjacent(t, m, "J")
	assert.Equal(t, 2, m.list.Index())
	assert.Equal(t, 1, m.list.Page())
	assert.Equal(t, 0, m.list.Cursor())

	thread = comment.ToThread(&hn.CommentTree{ID: 3, Title: "Third item", CommentsCount: 1})
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.id})

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

	// q inside the comment view emits DetailQuit as a command; deliver it.
	m, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	m, _ = m.Update(cmd())

	assert.Equal(t, screenList, m.screen)
	assert.Nil(t, m.detail)
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

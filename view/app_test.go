package view

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/settings"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"
	"github.com/bensadeh/circumflex/view/reader"
	xansi "github.com/charmbracelet/x/ansi"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// instantMockService implements hn.Service without the 1-second sleep.
type instantMockService struct{}

func (instantMockService) FetchItems(_ context.Context, _ int, _ string) ([]*hn.Story, error) {
	return testItems(), nil
}

func (instantMockService) FetchComments(_ context.Context, _ int, _ func(int, int)) (*hn.CommentTree, error) {
	return &hn.CommentTree{Story: hn.Story{ID: 1, Title: "test", CommentsCount: 5}}, nil
}

func (instantMockService) FetchItem(_ context.Context, _ int) (*hn.Story, error) {
	return &hn.Story{}, nil
}

func (instantMockService) SearchItems(_ context.Context, req hn.SearchRequest) ([]*hn.Story, error) {
	var hits []*hn.Story

	for _, s := range testItems() {
		if strings.Contains(strings.ToLower(s.Title), strings.ToLower(req.Query)) {
			hits = append(hits, s)
		}
	}

	return hits, nil
}

func testItems() []*hn.Story {
	return []*hn.Story{
		{ID: 1, Title: "First item", Points: 100, Author: "alice", Time: time.Now().Unix(), Domain: "example.com", CommentsCount: 10, URL: "https://example.com/1"},
		{ID: 2, Title: "Second item", Points: 200, Author: "bob", Time: time.Now().Unix(), Domain: "test.com", CommentsCount: 20, URL: "https://test.com/2"},
		{ID: 3, Title: "Third item", Points: 300, Author: "charlie", Time: time.Now().Unix(), Domain: "demo.com", CommentsCount: 30, URL: "https://demo.com/3"},
		{ID: 4, Title: "Fourth item", Points: 400, Author: "dave", Time: time.Now().Unix(), Domain: "site.com", CommentsCount: 40, URL: "https://site.com/4"},
		{ID: 5, Title: "Fifth item", Points: 500, Author: "eve", Time: time.Now().Unix(), Domain: "web.com", CommentsCount: 50, URL: "https://web.com/5"},
	}
}

func newTestModel(t *testing.T) *model {
	t.Helper()

	config := settings.Default()
	cat, _ := categories.New("top,best,ask,show")
	fav, err := favorites.New(filepath.Join(t.TempDir(), "favorites.toml"), filepath.Join(t.TempDir(), "favorites.json"))
	require.NoError(t, err)

	service := &instantMockService{}
	hist := history.NewMockHistory()

	return newModel(config, cat, fav, 80, 24, service, hist)
}

func newTestModelReady(t *testing.T) *model {
	t.Helper()
	m := newTestModel(t)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(message.StoriesReady{Stories: testItems(), Category: categories.Top, FetchID: m.fetch.currentID()})

	return m
}

// startTestFetch puts a detail fetch in flight the way the real handlers do,
// so a hand-delivered result passes the finish guard.
func startTestFetch(m *model, target screen) {
	_, _ = m.startDetailFetch(0, target, m.detailRollback(m.list.Index()))
}

func keyMsg(s string) tea.KeyPressMsg {
	switch s {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	default:
		r := []rune(s)

		return tea.KeyPressMsg{Code: r[0], Text: s}
	}
}

func TestStartup_WaitsForWindowSizeMsg(t *testing.T) {
	m := newTestModel(t)
	assert.False(t, m.started)

	m, cmd := m.Update(message.StatusMessageTimeout{})
	assert.False(t, m.started)
	assert.Nil(t, cmd)
}

func TestStartup_InitializesOnWindowSizeMsg(t *testing.T) {
	m := newTestModel(t)
	assert.False(t, m.started)

	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.True(t, m.fetch.inFlight())
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, cmd, "should return batch cmd with spinner + fetch")
}

func TestDetailQuit_FromReaderRestoresState(t *testing.T) {
	m := newTestModelReady(t)
	m.screen = screenReader

	m, _ = m.Update(message.DetailQuit{})
	assert.Equal(t, screenList, m.screen)
}

// A quit that races a fetch the detail view minted a cycle earlier must
// abort it: its result would otherwise reopen a story the user just left.
func TestDetailQuit_AbortsRacedFetch(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())

	staleID := m.fetch.currentID()

	m, _ = m.Update(message.DetailQuit{})
	assert.False(t, m.fetch.inFlight(), "the quit aborts the fetch")

	// The aborted fetch's late result is stale and dropped.
	m, _ = m.Update(message.CommentTreeDataReady{
		FetchID: staleID,
		Thread:  comment.ToThread(&hn.CommentTree{}),
	})
	assert.Nil(t, m.detail, "the story must not reopen over the front page")
	assert.Equal(t, screenList, m.screen)
}

func TestStatusMessageTimeout_ClearsExpiredMessage(t *testing.T) {
	m := newTestModelReady(t)

	timeout := m.status.NewStatusMessageWithDuration("transient", time.Millisecond)()
	m, _ = m.Update(timeout)

	assert.Empty(t, m.status.text.Message())
}

// The favorites-prompt scenario: a transient message's timer is still
// pending when the prompt's permanent text replaces it. The stale timer
// must not blank the prompt.
func TestStatusMessageTimeout_StaleTimerKeepsPermanentMessage(t *testing.T) {
	m := newTestModelReady(t)

	timeout := m.status.NewStatusMessageWithDuration("Item added", time.Millisecond)()
	m.status.SetPermanentStatusMessage("add to favorites?")

	m, _ = m.Update(timeout)

	assert.Equal(t, "add to favorites?", m.status.text.Message())
}

// A message expiring mid-fetch must not clear the fetch's terminal progress
// indicator; an indeterminate fetch never rewrites it.
func TestStatusMessageTimeout_MidFetchKeepsProgressIndicator(t *testing.T) {
	m := newTestModelReady(t)

	timeout := m.status.NewStatusMessageWithDuration("Item added", time.Millisecond)()

	m, _ = m.Update(keyMsg("tab"))
	require.True(t, m.fetch.inFlight())

	var progress strings.Builder

	prev := pane.ProgressOut
	pane.ProgressOut = &progress

	t.Cleanup(func() { pane.ProgressOut = prev })

	m, _ = m.Update(timeout)

	assert.Empty(t, m.status.text.Message(), "the message itself still expires")
	assert.NotContains(t, progress.String(), "\x1b]9;4;0", "the fetch's indicator survives")
}

func TestAddToFavorites_AddsItem(t *testing.T) {
	m := newTestModelReady(t)
	initialFavCount := len(m.favorites.Items())

	testItem := &hn.Story{ID: 99, Title: "Favorite item"}
	m, _ = m.Update(message.AddToFavorites{Item: testItem})

	assert.Len(t, m.favorites.Items(), initialFavCount+1)
	assert.Equal(t, 99, m.favorites.Items()[initialFavCount].ID)
}

func TestWindowResize_UpdatesDimensions(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestStoriesReady_UpdatesState(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("tab"))
	require.True(t, m.fetch.inFlight())
	require.True(t, m.list.InTransition())

	m, _ = m.Update(message.StoriesReady{Category: categories.Best, Index: 1, Cursor: 0, FetchID: m.fetch.currentID()})

	assert.False(t, m.fetch.inFlight())
	assert.False(t, m.status.showSpinner)
	assert.False(t, m.list.InTransition())
}

// A successful fetch that returns no stories keeps the cursor pinned at 0
// rather than clamping it negative.
func TestStoriesReady_EmptyResultKeepsCursorAtZero(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("tab"))
	require.True(t, m.fetch.inFlight())

	m, _ = m.Update(message.StoriesReady{Category: categories.Best, Index: 1, FetchID: m.fetch.currentID()})

	assert.False(t, m.fetch.inFlight())
	assert.Equal(t, 0, m.list.Cursor())
}

// favorites.Item is the one story shape that cannot share hn.Story — its
// TOML tags are the on-disk favorites contract — so its copy pair is pinned
// by a round trip instead: every hn.Story field, filled with a distinct
// non-zero value, must survive Story → Item → Story.
func TestFavorites_ItemRoundTripIsLossless(t *testing.T) {
	t.Parallel()

	story := &hn.Story{}

	v := reflect.ValueOf(story).Elem()
	for i := range v.NumField() {
		switch f := v.Field(i); {
		case f.CanInt():
			f.SetInt(int64(i) + 1)
		case f.Kind() == reflect.String:
			f.SetString(fmt.Sprintf("field-%d", i))
		default:
			t.Fatalf("hn.Story field %s has kind %s — extend this test to fill it",
				v.Type().Field(i).Name, f.Kind())
		}
	}

	got := favItemsToStories([]*favorites.Item{favorites.ItemFromStory(story)})
	require.Equal(t, story, got[0],
		"a field added to hn.Story must be copied in both ItemFromStory and favItemsToStories")
}

// A second J/K minted before the first press began its fetch arrives
// mid-flight; acting on it would move the selection again and record the
// half-open story as the rollback point.
func TestOpenAdjacentStory_IgnoredWhileFetchInFlight(t *testing.T) {
	m := newTestModelReady(t)
	m.screen = screenComments

	m, _ = m.Update(message.OpenAdjacentStory{Direction: 1})
	require.True(t, m.fetch.inFlight())
	require.Equal(t, 1, m.list.Index())
	firstFetch := m.fetch.currentID()

	m, _ = m.Update(message.OpenAdjacentStory{Direction: 1})
	assert.Equal(t, 1, m.list.Index(), "selection must not move while a fetch is in flight")
	assert.Equal(t, firstFetch, m.fetch.currentID(), "the in-flight fetch must not be superseded")
}

func TestQuit(t *testing.T) {
	for _, k := range []string{"q", "esc", "ctrl+c"} {
		m := newTestModelReady(t)
		_, cmd := m.Update(keyMsg(k))
		assert.NotNil(t, cmd, "key %q should return quit cmd", k)
	}
}

func TestBackspaceDoesNotQuitList(t *testing.T) {
	m := newTestModelReady(t)

	_, cmd := m.Update(keyMsg("backspace"))
	assert.Nil(t, cmd, "backspace in the list must not quit the app")
}

func TestBackspaceClosesHelpScreen(t *testing.T) {
	m := newTestModelReady(t)
	m.screen = screenHelp

	m, _ = m.Update(keyMsg("backspace"))
	assert.Equal(t, screenList, m.screen, "backspace should leave the help screen")
}

func TestBackspaceInterruptsFetch(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("tab"))
	require.True(t, m.fetch.inFlight())

	m, _ = m.Update(keyMsg("backspace"))
	assert.False(t, m.fetch.inFlight(), "backspace should cancel an in-flight fetch")
}

func TestNavigationUpDown(t *testing.T) {
	m := newTestModelReady(t)

	assert.Equal(t, 0, m.list.Cursor())
	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, 1, m.list.Cursor())

	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, 2, m.list.Cursor())

	m, _ = m.Update(keyMsg("k"))
	assert.Equal(t, 1, m.list.Cursor())

	m, _ = m.Update(keyMsg("down"))
	assert.Equal(t, 2, m.list.Cursor())

	m, _ = m.Update(keyMsg("up"))
	assert.Equal(t, 1, m.list.Cursor())
}

func TestNavigationUpDown_Clamped(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("k"))
	assert.Equal(t, 0, m.list.Cursor())

	for range 100 {
		m, _ = m.Update(keyMsg("j"))
	}

	assert.LessOrEqual(t, m.list.Cursor(), len(m.list.VisibleItems())-1)
}

func TestGoToTopBottom(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("G"))
	assert.Equal(t, len(m.list.VisibleItems())-1, m.list.Cursor())

	m, _ = m.Update(keyMsg("g"))
	assert.Equal(t, 0, m.list.Cursor())
}

func TestTabToCachedCategory(t *testing.T) {
	m := newTestModelReady(t)

	m.list.SetItems(categories.Best, testItems())

	initialIndex := m.cat.CurrentIndex()
	m, cmd := m.Update(keyMsg("tab"))

	assert.NotEqual(t, initialIndex, m.cat.CurrentIndex())
	assert.Nil(t, cmd, "cached category should not trigger a fetch")
}

func TestTabToUncachedCategory(t *testing.T) {
	m := newTestModelReady(t)

	assert.Empty(t, m.list.Items(categories.Best))

	m, cmd := m.Update(keyMsg("tab"))

	assert.True(t, m.fetch.inFlight())
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, cmd, "should return batch cmd with spinner + fetch")
	assert.True(t, m.list.InTransition(), "should have a transition with old items")
}

func TestTabFetchError_RollsBackCategory(t *testing.T) {
	m := newTestModelReady(t)
	require.Equal(t, 0, m.cat.CurrentIndex())

	m, _ = m.Update(keyMsg("tab"))
	require.True(t, m.fetch.inFlight())
	require.Equal(t, 1, m.cat.CurrentIndex())

	m, _ = m.Update(message.StoriesReady{Err: errors.New("boom"), FetchID: m.fetch.currentID()})

	assert.False(t, m.fetch.inFlight())
	assert.Equal(t, 0, m.cat.CurrentIndex(), "failed fetch should restore the category we left")
	assert.False(t, m.list.InTransition())
}

func TestTabFetchCancel_RollsBackCategory(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("tab"))
	require.True(t, m.fetch.inFlight())
	require.Equal(t, 1, m.cat.CurrentIndex())

	m, _ = m.Update(keyMsg("esc"))

	assert.False(t, m.fetch.inFlight())
	assert.Equal(t, 0, m.cat.CurrentIndex(), "cancelled fetch should restore the category we left")
	assert.False(t, m.list.InTransition())
}

func TestEnterCommentSection(t *testing.T) {
	m := newTestModelReady(t)

	m, cmd := m.Update(keyMsg("enter"))
	assert.True(t, m.fetch.inFlight())
	assert.True(t, m.status.showSpinner)
	assert.True(t, m.detailLoading())
	assert.NotNil(t, cmd)
}

func TestFetchComments_ReturnsCmd(t *testing.T) {
	m := newTestModelReady(t)

	cmd := m.fetchComments(fetchToken{ctx: context.Background()}, &hn.Story{ID: 1, CommentsCount: 10})
	assert.NotNil(t, cmd, "should return cmd for async comment fetching")

	msg := cmd()
	result, ok := msg.(message.CommentTreeDataReady)
	assert.True(t, ok, "cmd should produce CommentTreeDataReady message")
	assert.NotNil(t, result.Thread)
}

// A fetch command carries the fetch id that was current when it was built, so
// a result delivered after the user cancels must be dropped instead of opening
// the story they backed out of.
func TestFetchComments_ResultAfterCancelIsDropped(t *testing.T) {
	m := newTestModelReady(t)

	tok, _ := m.startDetailFetch(0, screenComments, m.detailRollback(m.list.Index()))
	cmd := m.fetchComments(tok, m.list.SelectedItem())

	_ = m.handleCancelFetch()
	require.False(t, m.fetch.inFlight())

	m, _ = m.Update(cmd())
	assert.Nil(t, m.detail, "a cancelled fetch's late result must not open its story")
	assert.Equal(t, screenList, m.screen)
}

func TestCommentTreeDataReady_OpensCommentView(t *testing.T) {
	m := newTestModelReady(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	startTestFetch(m, screenComments)

	thread := &comment.Thread{Story: hn.Story{ID: 1, Title: "test", CommentsCount: 5}}
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.currentID()})
	assert.Equal(t, screenComments, m.screen)
	assert.NotNil(t, m.detail)
}

func TestDetailQuit_FromCommentsRestoresState(t *testing.T) {
	m := newTestModelReady(t)
	m.screen = screenComments

	m, _ = m.Update(message.DetailQuit{})
	assert.Equal(t, screenList, m.screen)
}

func TestTimeRefreshTick_ReschedulesInEveryState(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenComments)

	thread := &comment.Thread{Story: hn.Story{ID: 1, Title: "test", CommentsCount: 5}}
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.currentID()})
	require.Equal(t, screenComments, m.screen)

	_, cmd := m.Update(message.TimeRefreshTick{})
	assert.NotNil(t, cmd, "tick in comment view must reschedule the next refresh")

	m.screen = screenHelp
	_, cmd = m.Update(message.TimeRefreshTick{})
	assert.NotNil(t, cmd, "tick in help screen must reschedule the next refresh")
}

// While a story loads in the narrow layout the front page stays up, dimmed;
// the selected story carries the same muted reading marker the wide layout
// uses instead of dimming into the rest, and the header carries the spinner.
func TestNarrowLoading_SelectedStoryShowsReadingMarker(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())

	view := m.View()
	assert.Contains(t, view, "\x1b[100m", "loading story should render on the bright-black bar")
	assert.NotContains(t, view, "\x1b[7m", "loading story should not keep the browsing highlight")
	assert.Contains(t, xansi.Strip(view), "Second item", "the front page must stay up during the fetch")
}

// In the narrow layout, J/K story navigation must keep the open story on
// screen while the next one loads instead of flashing the front page.
func TestNarrowAdjacentStory_StaysOnOpenStory(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenComments)

	thread := &comment.Thread{Story: hn.Story{ID: 1, Title: "An unmistakable thread title", CommentsCount: 5}}
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.currentID()})
	require.Equal(t, screenComments, m.screen)

	m, cmd := m.Update(keyMsg("J"))
	require.NotNil(t, cmd)
	m, _ = m.Update(cmd())

	require.True(t, m.fetch.inFlight())
	require.Equal(t, screenComments, m.screen)

	view := m.View()
	assert.Contains(t, view, "An unmistakable thread title")
	assert.NotContains(t, view, "Second item", "the front page must not flash through during the fetch")

	lines := strings.Split(view, "\n")
	assert.NotEmpty(t, strings.TrimSpace(xansi.Strip(lines[len(lines)-1])),
		"the last row should carry the loading spinner")
}

// Status messages raised while a detail view owns the full screen (e.g. J
// toward a story with no article) surface on the view's bottom row.
func TestNarrowDetail_StatusMessageShowsOnLastRow(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenComments)

	thread := &comment.Thread{Story: hn.Story{ID: 1, Title: "test", CommentsCount: 5}}
	m, _ = m.Update(message.CommentTreeDataReady{Thread: thread, FetchID: m.fetch.currentID()})
	require.Equal(t, screenComments, m.screen)

	m.status.NewStatusMessageWithDuration("Story has no article to read", statusMessageShort)

	lines := strings.Split(m.View(), "\n")
	assert.Contains(t, xansi.Strip(lines[len(lines)-1]), "Story has no article to read")
}

func TestFetchArticle_ReturnsCmd(t *testing.T) {
	m := newTestModelReady(t)

	cmd := m.fetchArticle(fetchToken{ctx: context.Background()}, &hn.Story{
		URL:    "https://example.com",
		Title:  "Test Article",
		Domain: "example.com",
		ID:     1,
	})
	assert.NotNil(t, cmd, "should return cmd for async article fetching")
}

func TestFetchArticle_InvalidDomain(t *testing.T) {
	m := newTestModelReady(t)

	cmd := m.fetchArticle(fetchToken{ctx: context.Background()}, &hn.Story{
		URL:    "https://youtube.com/watch?v=123",
		Title:  "Test Video",
		Domain: "youtube.com",
		ID:     1,
	})
	assert.NotNil(t, cmd)

	msg := cmd()
	result, ok := msg.(message.ArticleReady)
	assert.True(t, ok, "cmd should produce ArticleReady message")
	assert.Error(t, result.Err)
}

func TestArticleReady_WithError(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenReader)

	_, cmd := m.Update(message.ArticleReady{Err: errors.New("Reader Mode not supported"), FetchID: m.fetch.currentID()})
	assert.NotNil(t, cmd, "should return batch cmd with status message and editor finished")
}

func testParsedArticle() *article.Parsed {
	return article.NewParsedFromHTML("<h1>Test</h1><p>Article content for the reader view.</p>")
}

func TestArticleReady_WithParsedArticle(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenReader)

	m, _ = m.Update(message.ArticleReady{Parsed: testParsedArticle(), Title: "Test", FetchID: m.fetch.currentID()})
	assert.Equal(t, screenReader, m.screen)
	assert.NotNil(t, m.detail, "should create reader view")
}

func TestAddFavoritesPrompt(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("f"))
	assert.Equal(t, promptAddFavorite, m.prompt)
	assert.NotEmpty(t, m.status.text.Message())
}

func TestAddFavoritesConfirm(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("f"))
	assert.Equal(t, promptAddFavorite, m.prompt)

	m, cmd := m.Update(keyMsg("y"))
	assert.Equal(t, promptNone, m.prompt)
	assert.NotNil(t, cmd, "should return AddToFavorites cmd")
}

func TestAddFavoritesCancel(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("f"))
	assert.Equal(t, promptAddFavorite, m.prompt)

	m, _ = m.Update(keyMsg("n"))
	assert.Equal(t, promptNone, m.prompt)
}

func newFavoritesTestModel(t *testing.T) *model {
	t.Helper()

	config := settings.Default()
	cat, err := categories.New("top,favorites")
	require.NoError(t, err)

	fav, err := favorites.New(filepath.Join(t.TempDir(), "favorites.toml"), filepath.Join(t.TempDir(), "favorites.json"))
	require.NoError(t, err)

	m := newModel(config, cat, fav, 80, 24, &instantMockService{}, history.NewMockHistory())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(message.StoriesReady{Stories: testItems(), Category: categories.Top, FetchID: m.fetch.currentID()})

	return m
}

// Startup on the favorites tab is served through a fabricated fetch result;
// with no favorites saved it must land cleanly on the empty tab.
func TestStartup_OnEmptyFavorites(t *testing.T) {
	config := settings.Default()
	cat, err := categories.New("favorites,top")
	require.NoError(t, err)

	fav, err := favorites.New(filepath.Join(t.TempDir(), "favorites.toml"), filepath.Join(t.TempDir(), "favorites.json"))
	require.NoError(t, err)

	m := newModel(config, cat, fav, 80, 24, &instantMockService{}, history.NewMockHistory())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	require.True(t, m.fetch.inFlight())

	msg := m.fetchStoriesForFirstCategory(fetchToken{id: m.fetch.currentID()})()
	stories, ok := msg.(message.StoriesReady)
	require.True(t, ok, "the favorites pseudo-fetch should produce StoriesReady")
	require.NoError(t, stories.Err)

	m, _ = m.Update(msg)

	assert.False(t, m.fetch.inFlight())
	assert.Equal(t, categories.Favorites, m.cat.CurrentCategory())
	assert.Empty(t, m.list.VisibleItems())
	assert.Equal(t, 0, m.list.Cursor(), "the empty favorites tab must not drive the cursor negative")
}

func TestFavorites_HeaderAlwaysShown(t *testing.T) {
	m := newFavoritesTestModel(t)

	require.Empty(t, m.favorites.Items())
	assert.Contains(t, m.View(), "favorites")
}

func TestFavorites_EmptyShowsHint(t *testing.T) {
	m := newFavoritesTestModel(t)

	m.cat.SetIndex(1)
	require.Equal(t, categories.Favorites, m.cat.CurrentCategory())

	assert.Contains(t, m.list.View(m.listFrame()), "No favorites yet")
	assert.Contains(t, m.View(), "No favorites yet")
}

func TestFavorites_TabToEmptyDoesNotFetch(t *testing.T) {
	m := newFavoritesTestModel(t)

	m, cmd := m.Update(keyMsg("tab"))

	assert.Equal(t, categories.Favorites, m.cat.CurrentCategory())
	assert.False(t, m.fetch.inFlight(), "favorites is local; tabbing to it must not fetch")
	assert.Nil(t, cmd)
}

func TestDisabledInput_IgnoresKeys(t *testing.T) {
	m := newTestModelReady(t)
	m.fetch.begin(0, fetchList, screenList, rollbackPoint{})

	cursorBefore := m.list.Cursor()
	m, cmd := m.Update(keyMsg("j"))
	assert.Equal(t, cursorBefore, m.list.Cursor(), "cursor should not move when input is disabled")
	assert.Nil(t, cmd)
}

func TestHelpScreen_Toggle(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("i"))
	assert.Equal(t, screenHelp, m.screen)

	m, _ = m.Update(keyMsg("q"))
	assert.Equal(t, screenList, m.screen)
}

func TestRefresh(t *testing.T) {
	m := newTestModelReady(t)

	m, cmd := m.Update(keyMsg("r"))
	assert.True(t, m.fetch.inFlight())
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, cmd)
	assert.True(t, m.list.InTransition())
}

func TestOpenLink_ReturnsCmd(t *testing.T) {
	m := newTestModelReady(t)

	// Don't execute the cmd — it would open a real browser.
	cmd := m.handleOpenLink()
	assert.NotNil(t, cmd)
}

func TestSpinnerTick_WhenActive(t *testing.T) {
	m := newTestModelReady(t)
	m.status.showSpinner = true

	_, cmd := m.Update(m.status.spinner.Tick())
	assert.NotNil(t, cmd, "active spinner should return a follow-up tick cmd")
}

func TestSpinnerAnimation_FrameAdvances(t *testing.T) {
	m := newTestModelReady(t)

	startCmd := m.status.StartSpinner()
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, startCmd)

	initialView := m.status.spinner.View()
	tickMsg := startCmd()

	m, cmd := m.Update(tickMsg)
	afterFirstTick := m.status.spinner.View()

	assert.NotEqual(t, initialView, afterFirstTick, "spinner frame should change after tick")
	assert.NotNil(t, cmd, "should return next tick cmd")
}

func TestSpinnerTick_WhenInactive(t *testing.T) {
	m := newTestModelReady(t)
	m.status.showSpinner = false

	_, cmd := m.Update(m.status.spinner.Tick())
	// cmd may be non-nil from handleBrowsing, but the spinner-specific cmd won't be appended
	_ = cmd
}

func TestViewReaderView_HasContent(t *testing.T) {
	m := newTestModelReady(t)
	m.detail = reader.NewWithArticle(testParsedArticle(), "Test Title", 72, 80, 24, reader.Options{},
		meta.ReaderMode(meta.Data{URL: "https://example.com"}).Render)
	m.screen = screenReader

	got := m.View()
	assert.NotEmpty(t, got)
}

func TestViewBrowsing_HasContent(t *testing.T) {
	m := newTestModelReady(t)

	got := m.View()
	assert.NotEmpty(t, got)
}

// The browsing view must fill the terminal exactly: one row over and the
// renderer clips the status bar off the bottom of the screen.
func TestViewBrowsing_FillsScreenExactly(t *testing.T) {
	m := newTestModelReady(t)

	// Two pages of stories, so the paginator dots render on the last row —
	// a single page hides them.
	var stories []*hn.Story
	for i := range 20 {
		stories = append(stories, &hn.Story{ID: i + 1, Title: fmt.Sprintf("story %d", i+1), Author: "a", Time: time.Now().Unix()})
	}

	m.list.SetItems(categories.Top, stories)
	m.updatePagination()

	lines := strings.Split(m.View(), "\n")
	require.Len(t, lines, 24)
	assert.Contains(t, lines[23], "•", "last row should carry the paginator dots")
}

// A single page renders no paginator — a lone dot indicates nothing.
func TestPaginator_HiddenForSinglePage(t *testing.T) {
	m := newTestModelReady(t)

	assert.Empty(t, m.list.PaginatorView())
	assert.Empty(t, m.list.DimmedPaginatorView())
}

// "No results" must wait for the fetch: while a search is in flight the
// empty pane means "loading", not "nothing matched".
func TestSearch_NoResultsMessageWaitsForFetch(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("/"))
	for _, r := range "zzz" {
		m, _ = m.Update(keyMsg(string(r)))
	}

	m, _ = m.Update(keyMsg("enter"))
	require.True(t, m.fetch.inFlight())
	assert.NotContains(t, m.View(), "No results", "the fetch is still running")

	m, _ = m.Update(message.StoriesReady{Category: categories.Search, Index: -1, FetchID: m.fetch.currentID()})
	assert.Contains(t, m.View(), "No results for “zzz”")
}

func TestViewHelpScreen(t *testing.T) {
	m := newTestModelReady(t)
	m.screen = screenHelp

	got := m.View()
	assert.NotEmpty(t, got)
}

func TestSpinnerView_WhenActive(t *testing.T) {
	m := newTestModelReady(t)
	m.status.showSpinner = true

	got := m.statusAndPaginationView()
	assert.NotEmpty(t, got)
}

type failingHistory struct {
	history.Mock

	writeErr error
}

func (f failingHistory) MarkRead(_ int, _ int) error {
	return f.writeErr
}

func (f failingHistory) MarkArticleRead(_ int) error {
	return f.writeErr
}

func (f failingHistory) MarkUnread(_ int) error {
	return f.writeErr
}

func TestCommentTreeDataReady_HistoryWarning(t *testing.T) {
	m := newTestModelReady(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	startTestFetch(m, screenComments)

	thread := &comment.Thread{Story: hn.Story{ID: 1, Title: "test", CommentsCount: 5}}
	histErr := errors.New("disk full")

	m, cmd := m.Update(message.CommentTreeDataReady{
		Thread:         thread,
		FetchID:        m.fetch.currentID(),
		HistoryWarning: histErr,
	})

	assert.Equal(t, screenComments, m.screen)
	assert.NotNil(t, cmd, "should return batched cmd with init + warning")
}

func TestArticleReady_HistoryWarning(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenReader)

	histErr := errors.New("disk full")

	m, cmd := m.Update(message.ArticleReady{
		Parsed:         testParsedArticle(),
		Title:          "Test",
		FetchID:        m.fetch.currentID(),
		HistoryWarning: histErr,
	})

	assert.Equal(t, screenReader, m.screen)
	assert.NotNil(t, cmd, "should return batched cmd with init + warning")
}

func TestArticleReady_NoHistoryWarning(t *testing.T) {
	m := newTestModelReady(t)
	startTestFetch(m, screenReader)

	m, cmd := m.Update(message.ArticleReady{
		Parsed:  testParsedArticle(),
		Title:   "Test",
		FetchID: m.fetch.currentID(),
	})

	assert.Equal(t, screenReader, m.screen)
	assert.Nil(t, cmd, "Init returns nil, no warning — cmd should be nil")
}

func TestBrowserOpenFailed_RecordsErrorForExit(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(message.BrowserOpenFailed{Err: errors.New("xdg-open not found")})

	require.Error(t, m.browserErr)
	assert.Contains(t, m.browserErr.Error(), "xdg-open not found")
}

func TestFetchComments_HistoryWriteFailure(t *testing.T) {
	config := settings.Default()
	cat, _ := categories.New("top,best,ask,show")
	fav, err := favorites.New(filepath.Join(t.TempDir(), "favorites.toml"), filepath.Join(t.TempDir(), "favorites.json"))
	require.NoError(t, err)

	hist := failingHistory{writeErr: errors.New("permission denied")}
	m := newModel(config, cat, fav, 80, 24, &instantMockService{}, hist)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.list.SetItems(categories.Top, testItems())

	cmd := m.fetchComments(fetchToken{ctx: context.Background()}, &hn.Story{ID: 1, CommentsCount: 10})
	assert.NotNil(t, cmd)

	msg := cmd()
	result, ok := msg.(message.CommentTreeDataReady)
	assert.True(t, ok)
	require.NoError(t, result.Err, "fetch itself should succeed")
	require.Error(t, result.HistoryWarning, "history write should fail")
	assert.Contains(t, result.HistoryWarning.Error(), "permission denied")
}

func TestFetchArticle_ValidationFailure_SkipsHistoryWrite(t *testing.T) {
	config := settings.Default()
	cat, _ := categories.New("top,best,ask,show")
	fav, err := favorites.New(filepath.Join(t.TempDir(), "favorites.toml"), filepath.Join(t.TempDir(), "favorites.json"))
	require.NoError(t, err)

	hist := failingHistory{writeErr: errors.New("should not be reached")}
	m := newModel(config, cat, fav, 80, 24, &instantMockService{}, hist)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.list.SetItems(categories.Top, testItems())

	cmd := m.fetchArticle(fetchToken{ctx: context.Background()}, &hn.Story{
		URL:    "https://youtube.com/watch?v=123",
		Title:  "Test Video",
		Domain: "youtube.com",
		ID:     1,
	})
	require.NotNil(t, cmd)

	msg := cmd()
	result, ok := msg.(message.ArticleReady)
	require.True(t, ok)
	require.Error(t, result.Err, "validation should fail for youtube.com")
	assert.NoError(t, result.HistoryWarning, "history write should be skipped on validation failure")
}

// openTestReader puts a reader in the detail pane the way handleArticleReady
// does.
func openTestReader(t *testing.T, m *model) *model {
	t.Helper()

	startTestFetch(m, screenReader)
	m, _ = m.Update(message.ArticleReady{Parsed: testParsedArticle(), Title: "Root", FetchID: m.fetch.currentID()})
	require.Equal(t, screenReader, m.screen)
	require.NotNil(t, m.detail)

	return m
}

func TestLinkArticleReady_ReplacesReader(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))
	root := m.detail

	_, _ = m.startLinkFetch(0)
	m, _ = m.Update(message.LinkArticleReady{Parsed: testParsedArticle(), Title: "Linked", URL: "https://example.com/linked", FetchID: m.fetch.currentID()})

	assert.Equal(t, screenReader, m.screen)
	assert.NotSame(t, root, m.detail, "the linked page takes the article's place")

	m, _ = m.Update(message.DetailQuit{})
	assert.Equal(t, screenList, m.screen, "quit from the linked page returns to the front page")
}

func TestLinkArticleReady_ErrorStaysOnArticle(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))
	root := m.detail

	_, _ = m.startLinkFetch(0)
	m, cmd := m.Update(message.LinkArticleReady{Err: errors.New("server returned status 404"), FetchID: m.fetch.currentID()})

	assert.Same(t, root, m.detail, "the open article never transitions on failure")
	assert.Equal(t, screenReader, m.screen)
	assert.NotNil(t, cmd, "the failure surfaces as a status message")
}

func TestOpenReaderLink_InvalidDomainStaysPut(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))
	root := m.detail

	m, cmd := m.Update(message.OpenReaderLink{URL: "https://youtube.com/watch?v=123"})

	assert.False(t, m.fetch.inFlight(), "a blocked domain never starts a fetch")
	assert.Same(t, root, m.detail)
	assert.NotNil(t, cmd, "the rejection surfaces as a status message")
}

func TestOpenReaderLink_StartsLinkFetch(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))

	m, cmd := m.Update(message.OpenReaderLink{URL: "https://example.com/page"})

	assert.True(t, m.fetch.linkLoading())
	assert.NotNil(t, cmd)
}

func TestLinkArticleReady_TrailFeedsDepthBadge(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))

	_, _ = m.startLinkFetch(0)
	m, _ = m.Update(message.LinkArticleReady{
		Parsed: testParsedArticle(),
		Title:  "Two Deep",
		URL:    "https://example.com/b",
		Trail: []message.TrailEntry{
			{URL: "https://example.com/story", Story: true},
			{URL: "https://example.com/a"},
		},
		FetchID: m.fetch.currentID(),
	})

	assert.Contains(t, xansi.Strip(m.detail.View()), "⧉  2", "two links followed, two steps back")
}

func TestRestoreReaderPage_LinkEntryKeepsChain(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))

	m, _ = m.Update(message.RestoreReaderPage{
		Entry: message.TrailEntry{URL: "https://example.com/a", Title: "Page A", Parsed: testParsedArticle()},
		Trail: []message.TrailEntry{{URL: "https://example.com/story", Story: true}},
	})

	require.NotNil(t, m.detail)
	assert.False(t, m.fetch.inFlight(), "a restore never touches the network")
	assert.Equal(t, screenReader, m.screen)
	assert.Contains(t, xansi.Strip(m.detail.View()), "⧉  1", "one step back remains")
}

func TestRestoreReaderPage_StoryEntryGetsStoryMeta(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))

	m, _ = m.Update(message.RestoreReaderPage{
		Entry: message.TrailEntry{URL: "https://example.com/1", Title: "First item", Parsed: testParsedArticle(), Story: true},
	})

	require.NotNil(t, m.detail)
	assert.False(t, m.fetch.inFlight())

	view := xansi.Strip(m.detail.View())
	assert.Contains(t, view, "by alice", "the story article gets its byline back")
	assert.NotContains(t, view, "⧉", "the story article carries no badge")

	// Quit from the restored story article goes to the front page.
	m, cmd := m.Update(message.DetailQuit{})
	assert.Equal(t, screenList, m.screen)
	assert.Nil(t, m.detail)
	assert.Nil(t, cmd)
}

// Quit on a linked page steps back to the story article by re-fetching it —
// direction 0 through the adjacent-story flow, so no saved view exists to go
// stale.
func TestOpenAdjacentStory_ZeroReopensSelectedStory(t *testing.T) {
	m := openTestReader(t, newTestModelReady(t))

	_, _ = m.startLinkFetch(0)
	m, _ = m.Update(message.LinkArticleReady{Parsed: testParsedArticle(), Title: "Linked", URL: "https://example.com/linked", FetchID: m.fetch.currentID()})
	linked := m.detail

	m, cmd := m.Update(message.OpenAdjacentStory{Direction: 0})

	assert.True(t, m.fetch.detailLoading(), "the story article re-fetches")
	assert.Equal(t, 0, m.list.Index(), "the selection stays on the story")
	assert.Same(t, linked, m.detail, "the linked page stays visible while the article loads")
	require.NotNil(t, cmd)

	m, _ = m.Update(message.ArticleReady{Parsed: testParsedArticle(), Title: "Root again", FetchID: m.fetch.currentID()})
	assert.Equal(t, screenReader, m.screen)
	assert.NotSame(t, linked, m.detail, "the article replaces the linked page")
}

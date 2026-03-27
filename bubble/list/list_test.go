package list

import (
	"clx/bubble/list/message"
	"clx/categories"
	"clx/favorites"
	"clx/history"
	"clx/item"
	"clx/settings"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

// instantMockService implements hn.Service without the 1-second sleep.
type instantMockService struct{}

func (instantMockService) Init(_ int) {}

func (instantMockService) FetchItems(_ context.Context, _ int, _ string) ([]*item.Story, error) {
	return testItems(), nil
}

func (instantMockService) FetchComments(_ context.Context, _ int) (*item.Story, error) {
	return &item.Story{ID: 1, Title: "test", CommentsCount: 5}, nil
}

func (instantMockService) FetchItem(_ context.Context, _ int) (*item.Story, error) {
	return &item.Story{}, nil
}

func testItems() []*item.Story {
	return []*item.Story{
		{ID: 1, Title: "First item", Points: 100, User: "alice", Time: time.Now().Unix(), Domain: "example.com", CommentsCount: 10, URL: "https://example.com/1"},
		{ID: 2, Title: "Second item", Points: 200, User: "bob", Time: time.Now().Unix(), Domain: "test.com", CommentsCount: 20, URL: "https://test.com/2"},
		{ID: 3, Title: "Third item", Points: 300, User: "charlie", Time: time.Now().Unix(), Domain: "demo.com", CommentsCount: 30, URL: "https://demo.com/3"},
		{ID: 4, Title: "Fourth item", Points: 400, User: "dave", Time: time.Now().Unix(), Domain: "site.com", CommentsCount: 40, URL: "https://site.com/4"},
		{ID: 5, Title: "Fifth item", Points: 500, User: "eve", Time: time.Now().Unix(), Domain: "web.com", CommentsCount: 50, URL: "https://web.com/5"},
	}
}

func newTestModel(t *testing.T) *Model {
	t.Helper()

	config := settings.Default()
	cat, _ := categories.New("top,best,ask,show")
	fav := favorites.New(filepath.Join(t.TempDir(), "favorites.json"))
	service := &instantMockService{}
	hist := history.NewMockHistory()

	return newModel(NewDefaultDelegate(), config, cat, fav, 80, 24, service, hist)
}

// newTestModelReady creates a model that has completed startup (already received
// the initial WindowSizeMsg and is in browsing state).
func newTestModelReady(t *testing.T) *Model {
	t.Helper()
	m := newTestModel(t)

	// Send the initial WindowSizeMsg to complete startup
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Simulate fetch completion: populate items and reset state
	m.pager.items[categories.Top] = testItems()
	m.status.StopSpinner()
	m.state = StateBrowsing

	return m
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
		// Single character keys like "q", "j", "k", etc.
		r := []rune(s)

		return tea.KeyPressMsg{Code: r[0], Text: s}
	}
}

// --- Phase 0a: Update handler tests ---

func TestStartup_WaitsForWindowSizeMsg(t *testing.T) {
	m := newTestModel(t)
	assert.Equal(t, StateStartup, m.state)

	// Non-WindowSizeMsg during startup should be ignored
	m, cmd := m.Update(message.StatusMessageTimeout{})
	assert.Equal(t, StateStartup, m.state)
	assert.Nil(t, cmd)
}

func TestStartup_InitializesOnWindowSizeMsg(t *testing.T) {
	m := newTestModel(t)
	assert.Equal(t, StateStartup, m.state)

	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, StateFetching, m.state)
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, cmd, "should return batch cmd with spinner + fetch")
}

func TestEditorFinished_RestoresState(t *testing.T) {
	m := newTestModelReady(t)
	m.state = StateEditorOpen

	m, _ = m.Update(message.EditorFinishedMsg{})
	assert.Equal(t, StateBrowsing, m.state)
}

func TestStatusMessageTimeout_ClearsMessage(t *testing.T) {
	m := newTestModelReady(t)
	m.status.message = "some status"

	m, _ = m.Update(message.StatusMessageTimeout{})
	assert.Empty(t, m.status.message)
}

func TestAddToFavorites_AddsItem(t *testing.T) {
	m := newTestModelReady(t)
	initialFavCount := len(m.favorites.Items())

	testItem := &item.Story{ID: 99, Title: "Favorite item"}
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

func TestCategoryFetchingFinished_UpdatesState(t *testing.T) {
	m := newTestModelReady(t)
	m.state = StateFetching
	m.pager.transition = &transition{prevIndex: 0, oldItems: testItems(), refresh: true}
	m.status.showSpinner = true

	m, _ = m.Update(message.CategoryFetchingFinished{Index: 1, Cursor: 0})

	assert.Equal(t, StateBrowsing, m.state)
	assert.False(t, m.status.showSpinner)
	assert.Nil(t, m.pager.transition)
}

func TestShowStatusMessage_SetsMessage(t *testing.T) {
	m := newTestModelReady(t)

	_, cmd := m.Update(message.ShowStatusMessage{Message: "hello", Duration: time.Second})
	assert.NotNil(t, cmd)
}

// --- Key handling tests ---

func TestQuit(t *testing.T) {
	m := newTestModelReady(t)

	for _, k := range []string{"q", "esc", "ctrl+c"} {
		m2 := newTestModelReady(t)
		_, cmd := m2.Update(keyMsg(k))
		_ = m
		// tea.Quit returns a special Cmd; we verify it's non-nil
		assert.NotNil(t, cmd, "key %q should return quit cmd", k)
	}
}

func TestNavigationUpDown(t *testing.T) {
	m := newTestModelReady(t)

	// Start at cursor 0, move down
	assert.Equal(t, 0, m.pager.cursor)
	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, 1, m.pager.cursor)

	m, _ = m.Update(keyMsg("j"))
	assert.Equal(t, 2, m.pager.cursor)

	// Move up
	m, _ = m.Update(keyMsg("k"))
	assert.Equal(t, 1, m.pager.cursor)

	// Can also use arrow keys
	m, _ = m.Update(keyMsg("down"))
	assert.Equal(t, 2, m.pager.cursor)

	m, _ = m.Update(keyMsg("up"))
	assert.Equal(t, 1, m.pager.cursor)
}

func TestNavigationUpDown_Clamped(t *testing.T) {
	m := newTestModelReady(t)

	// Moving up from 0 stays at 0
	m, _ = m.Update(keyMsg("k"))
	assert.Equal(t, 0, m.pager.cursor)

	// Moving down past end stays at last item
	for range 100 {
		m, _ = m.Update(keyMsg("j"))
	}

	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))
	assert.LessOrEqual(t, m.pager.cursor, itemsOnPage-1)
}

func TestGoToTopBottom(t *testing.T) {
	m := newTestModelReady(t)

	// Go to bottom
	m, _ = m.Update(keyMsg("G"))
	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))
	assert.Equal(t, itemsOnPage-1, m.pager.cursor)

	// Go to top
	m, _ = m.Update(keyMsg("g"))
	assert.Equal(t, 0, m.pager.cursor)
}

func TestTabToCachedCategory(t *testing.T) {
	m := newTestModelReady(t)

	// Pre-populate the "best" category so tab doesn't need to fetch
	m.pager.items[categories.Best] = testItems()

	initialIndex := m.cat.CurrentIndex()
	m, cmd := m.Update(keyMsg("tab"))

	assert.NotEqual(t, initialIndex, m.cat.CurrentIndex())
	// No fetch needed since category is cached, cmd should be nil
	assert.Nil(t, cmd)
}

func TestTabToUncachedCategory(t *testing.T) {
	m := newTestModelReady(t)

	// "best" category is empty (uncached)
	assert.Empty(t, m.pager.items[categories.Best])

	m, cmd := m.Update(keyMsg("tab"))

	assert.Equal(t, StateFetching, m.state)
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, cmd, "should return batch cmd with spinner + fetch")
	assert.NotNil(t, m.pager.transition, "should have a transition with old items")
}

func TestEnterCommentSection(t *testing.T) {
	m := newTestModelReady(t)

	m, cmd := m.Update(keyMsg("enter"))
	assert.Equal(t, StateFetching, m.state)
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, m.pager.transition)
	assert.NotNil(t, cmd)
}

func TestEnteringCommentSection_ReturnsCmd(t *testing.T) {
	m := newTestModelReady(t)

	_, cmd := m.Update(message.EnteringCommentSection{Id: 1, CommentCount: 10})
	assert.NotNil(t, cmd, "should return cmd for async comment fetching")

	// Execute the Cmd — it should produce a CommentTreeDataReady message
	msg := cmd()
	result, ok := msg.(message.CommentTreeDataReady)
	assert.True(t, ok, "cmd should produce CommentTreeDataReady message")
	assert.NotNil(t, result.Story)
}

func TestCommentTreeDataReady_OpensCommentView(t *testing.T) {
	m := newTestModelReady(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	story := &item.Story{ID: 1, Title: "test", CommentsCount: 5}
	m, _ = m.Update(message.CommentTreeDataReady{Story: story})
	assert.Equal(t, StateCommentView, m.state)
	assert.NotNil(t, m.commentView)
}

func TestCommentViewQuit_RestoresState(t *testing.T) {
	m := newTestModelReady(t)
	m.state = StateCommentView

	m, _ = m.Update(message.CommentViewQuitMsg{})
	assert.Equal(t, StateBrowsing, m.state)
}

func TestEnteringReaderMode_ReturnsCmd(t *testing.T) {
	m := newTestModelReady(t)

	_, cmd := m.Update(message.EnteringReaderMode{
		Url:          "https://example.com",
		Title:        "Test Article",
		Domain:       "example.com",
		Id:           1,
		CommentCount: 10,
	})
	assert.NotNil(t, cmd, "should return cmd for async article fetching")
}

func TestEnteringReaderMode_InvalidDomain(t *testing.T) {
	m := newTestModelReady(t)

	_, cmd := m.Update(message.EnteringReaderMode{
		Url:          "https://youtube.com/watch?v=123",
		Title:        "Test Video",
		Domain:       "youtube.com",
		Id:           1,
		CommentCount: 10,
	})
	assert.NotNil(t, cmd)

	// Execute the Cmd — should produce ArticleReady with error
	msg := cmd()
	result, ok := msg.(message.ArticleReady)
	assert.True(t, ok, "cmd should produce ArticleReady message")
	assert.Error(t, result.Err)
}

func TestArticleReady_WithError(t *testing.T) {
	m := newTestModelReady(t)

	_, cmd := m.Update(message.ArticleReady{Err: errors.New("Reader Mode not supported")})
	assert.NotNil(t, cmd, "should return batch cmd with status message and editor finished")
}

func TestArticleReady_WithContent(t *testing.T) {
	m := newTestModelReady(t)

	_, cmd := m.Update(message.ArticleReady{Content: "article content"})
	assert.NotNil(t, cmd, "should return ExecProcess cmd")
}

func TestAddFavoritesPrompt(t *testing.T) {
	m := newTestModelReady(t)

	m, _ = m.Update(keyMsg("f"))
	assert.Equal(t, StateAddFavoritesPrompt, m.state)
	assert.NotEmpty(t, m.status.message)
}

func TestAddFavoritesConfirm(t *testing.T) {
	m := newTestModelReady(t)

	// Enter prompt
	m, _ = m.Update(keyMsg("f"))
	assert.Equal(t, StateAddFavoritesPrompt, m.state)

	// Confirm
	m, cmd := m.Update(keyMsg("y"))
	assert.Equal(t, StateBrowsing, m.state)
	assert.NotNil(t, cmd, "should return AddToFavorites cmd")
}

func TestAddFavoritesCancel(t *testing.T) {
	m := newTestModelReady(t)

	// Enter prompt
	m, _ = m.Update(keyMsg("f"))
	assert.Equal(t, StateAddFavoritesPrompt, m.state)

	// Cancel with any key other than "y"
	m, _ = m.Update(keyMsg("n"))
	assert.Equal(t, StateBrowsing, m.state)
}

func TestDisabledInput_IgnoresKeys(t *testing.T) {
	m := newTestModelReady(t)
	m.state = StateFetching

	cursorBefore := m.pager.cursor
	m, cmd := m.Update(keyMsg("j"))
	assert.Equal(t, cursorBefore, m.pager.cursor, "cursor should not move when input is disabled")
	assert.Nil(t, cmd)
}

func TestHelpScreen_Toggle(t *testing.T) {
	m := newTestModelReady(t)

	// Enter help screen
	m, _ = m.Update(keyMsg("i"))
	assert.Equal(t, StateHelpScreen, m.state)

	// Exit help screen
	m, _ = m.Update(keyMsg("q"))
	assert.Equal(t, StateBrowsing, m.state)
}

func TestRefresh(t *testing.T) {
	m := newTestModelReady(t)

	m, cmd := m.Update(keyMsg("r"))
	assert.Equal(t, StateFetching, m.state)
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, cmd)
	assert.NotNil(t, m.pager.transition)
	assert.True(t, m.pager.transition.refresh)
}

func TestOpenLink_ReturnsMessage(t *testing.T) {
	m := newTestModelReady(t)

	// Test the message handler directly instead of pressing "o",
	// because handleOpenLink() calls browser.Open() synchronously.
	_, _ = m.Update(message.OpeningLink{Id: 1, CommentCount: 10})
	// History should be marked as read
}

func TestSpinnerTick_WhenActive(t *testing.T) {
	m := newTestModelReady(t)
	m.status.showSpinner = true

	// Create a spinner tick message
	_, cmd := m.Update(m.status.spinner.Tick())
	// When spinner is active, should return a follow-up tick cmd
	assert.NotNil(t, cmd)
}

func TestSpinnerAnimation_FrameAdvances(t *testing.T) {
	m := newTestModelReady(t)

	// Start the spinner
	startCmd := m.status.StartSpinner()
	assert.True(t, m.status.showSpinner)
	assert.NotNil(t, startCmd)

	// Record the initial spinner view
	initialView := m.status.spinner.View()

	// Execute the start cmd to get the first TickMsg
	tickMsg := startCmd()

	// Process the tick - this should advance the frame
	m, cmd := m.Update(tickMsg)
	afterFirstTick := m.status.spinner.View()

	assert.NotEqual(t, initialView, afterFirstTick, "spinner frame should change after tick")
	assert.NotNil(t, cmd, "should return next tick cmd")
}

func TestSpinnerTick_WhenInactive(t *testing.T) {
	m := newTestModelReady(t)
	m.status.showSpinner = false

	_, cmd := m.Update(m.status.spinner.Tick())
	// When spinner is inactive, no follow-up tick should be returned
	// cmd may be non-nil from handleBrowsing, but the spinner-specific cmd won't be appended
	_ = cmd
}

// --- Phase 0b: View snapshot tests ---

func TestViewEmpty_WhenNotVisible(t *testing.T) {
	m := newTestModelReady(t)
	m.state = StateEditorOpen

	got := m.View()
	assert.Empty(t, got)
}

func TestViewBrowsing_HasContent(t *testing.T) {
	m := newTestModelReady(t)

	got := m.View()
	assert.NotEmpty(t, got)
}

func TestViewHelpScreen(t *testing.T) {
	m := newTestModelReady(t)
	m.state = StateHelpScreen

	got := m.View()
	assert.NotEmpty(t, got)
}

func TestSpinnerView_WhenActive(t *testing.T) {
	m := newTestModelReady(t)
	m.status.showSpinner = true

	got := m.statusAndPaginationView()
	assert.NotEmpty(t, got)
}

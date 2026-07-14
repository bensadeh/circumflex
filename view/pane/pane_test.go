package pane

import (
	"errors"
	"image/color"
	"testing"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenStoryInBrowser(t *testing.T) {
	assert.Nil(t, OpenStoryInBrowser("", 0), "no URL and no ID should not open anything")
	assert.NotNil(t, OpenStoryInBrowser("", 42), "self-post should fall back to the HN item URL")
	assert.NotNil(t, OpenStoryInBrowser("https://example.com", 42))
}

func TestOpenCommentsInBrowser(t *testing.T) {
	assert.Nil(t, OpenCommentsInBrowser(0), "no ID should not open a comments page")
	assert.NotNil(t, OpenCommentsInBrowser(42))
}

type fakeView struct {
	msgs []tea.Msg
}

func (f *fakeView) Init() tea.Cmd { return nil }

func (f *fakeView) Update(msg tea.Msg) tea.Cmd {
	f.msgs = append(f.msgs, msg)

	return nil
}

func (f *fakeView) View() string { return "fake" }

func TestStandalone_CreatesViewOnFirstWindowSize(t *testing.T) {
	fv := &fakeView{}

	var gotWidth, gotHeight int

	s := standalone{makeView: func(width, height int) View {
		gotWidth, gotHeight = width, height

		return fv
	}}

	next, _ := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	assert.Equal(t, 80, gotWidth)
	assert.Equal(t, 24, gotHeight)
	assert.Empty(t, fv.msgs, "the creating WindowSizeMsg should not also be forwarded")

	created, ok := next.(standalone)
	require.True(t, ok)

	created.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	assert.Len(t, fv.msgs, 1, "later WindowSizeMsgs should be forwarded")
}

func TestStandalone_ReplaysEarlyBackgroundColorToNewView(t *testing.T) {
	fv := &fakeView{}
	s := standalone{makeView: func(int, int) View { return fv }}

	bg := tea.BackgroundColorMsg{Color: color.White}

	next, _ := s.Update(bg)
	withBG, ok := next.(standalone)
	require.True(t, ok)

	next, _ = withBG.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	require.Len(t, fv.msgs, 1, "the stashed background color is replayed on view creation")
	assert.Equal(t, bg, fv.msgs[0])

	created, ok := next.(standalone)
	require.True(t, ok)

	created.Update(bg)
	assert.Len(t, fv.msgs, 2, "a report arriving after creation is forwarded directly")
}

func TestStandalone_QuitMessagesEndProgram(t *testing.T) {
	s := standalone{view: &fakeView{}}

	_, cmd := s.Update(message.DetailQuit{})
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd(), "a detail quit should end the program")
}

func TestStandalone_CtrlCQuits(t *testing.T) {
	s := standalone{view: &fakeView{}}

	_, cmd := s.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestStandalone_CapturesBrowserError(t *testing.T) {
	s := standalone{view: &fakeView{}}
	err := errors.New("no browser")

	next, _ := s.Update(message.BrowserOpenFailed{Err: err})

	updated, ok := next.(standalone)
	require.True(t, ok)
	assert.Equal(t, err, updated.browserErr)
}

// pageFactory records the pages it is asked to build.
type pageFactory struct {
	entries []message.TrailEntry
	trails  [][]message.TrailEntry
	view    *fakeView
}

func (f *pageFactory) make(entry message.TrailEntry, trail []message.TrailEntry, _, _ int) View {
	f.entries = append(f.entries, entry)
	f.trails = append(f.trails, trail)

	return f.view
}

func TestStandalone_FollowedLinkSwapsPageInPlace(t *testing.T) {
	factory := &pageFactory{view: &fakeView{}}
	s := standalone{view: &fakeView{}, makePageView: factory.make}

	next, cmd := s.Update(message.OpenReaderLink{URL: "https://example.com/page"})
	require.NotNil(t, cmd, "a valid link should start a fetch")

	fetching, ok := next.(standalone)
	require.True(t, ok)

	parsed := article.NewParsedFromHTML("<p>linked page</p>")
	trail := []message.TrailEntry{{URL: "https://example.com/", Story: true}}

	next, initCmd := fetching.Update(message.LinkArticleReady{
		Parsed: parsed, Title: "Linked", URL: "https://example.com/page", Trail: trail, FetchID: fetching.fetchID,
	})

	swapped, ok := next.(standalone)
	require.True(t, ok)
	assert.Same(t, factory.view, swapped.view, "the fetched page replaces the view")
	assert.Nil(t, initCmd, "the fake view's Init returns no command")

	require.Len(t, factory.entries, 1)
	assert.Equal(t, "Linked", factory.entries[0].Title)
	assert.Equal(t, trail, factory.trails[0])
}

func TestStandalone_StaleAndFailedLinkFetchesKeepThePage(t *testing.T) {
	original := &fakeView{}
	factory := &pageFactory{view: &fakeView{}}
	s := standalone{view: original, makePageView: factory.make, fetchID: 2}

	next, _ := s.Update(message.LinkArticleReady{FetchID: 1, Title: "stale"})
	updated, ok := next.(standalone)
	require.True(t, ok)
	assert.Same(t, original, updated.view, "a superseded fetch's result is dropped")

	err := errors.New("could not fetch URL")

	next, _ = updated.Update(message.LinkArticleReady{FetchID: 2, Err: err})
	updated, ok = next.(standalone)
	require.True(t, ok)
	assert.Same(t, original, updated.view, "a failed fetch leaves the open page")
	assert.Equal(t, "Could not fetch URL", updated.statusMsg, "the failure shows on the footer row")
	assert.Empty(t, factory.entries)

	next, _ = updated.Update(statusTimeoutMsg{generation: updated.statusGen})
	updated, ok = next.(standalone)
	require.True(t, ok)
	assert.Empty(t, updated.statusMsg, "the failure message expires")
}

func TestStandalone_InvalidLinkNeverLeavesThePage(t *testing.T) {
	factory := &pageFactory{view: &fakeView{}}
	s := standalone{view: &fakeView{}, makePageView: factory.make}

	next, cmd := s.Update(message.OpenReaderLink{URL: "ftp://example.com/file"})
	assert.NotNil(t, cmd, "the validation failure schedules its message expiry")

	updated, ok := next.(standalone)
	require.True(t, ok)
	assert.NotEmpty(t, updated.statusMsg)
	assert.Nil(t, updated.cancelFetch, "no fetch starts for an unreadable link")
}

func TestStandalone_LinkFallsBackToBrowserWithoutFactory(t *testing.T) {
	s := standalone{view: &fakeView{}}

	_, cmd := s.Update(message.OpenReaderLink{URL: "https://example.com/page"})
	assert.NotNil(t, cmd)
}

func TestStandalone_RestoreRebuildsPageFromItsParse(t *testing.T) {
	factory := &pageFactory{view: &fakeView{}}
	s := standalone{view: &fakeView{}, makePageView: factory.make}

	entry := message.TrailEntry{URL: "https://example.com/", Story: true, Parsed: article.NewParsedFromHTML("<p>root</p>")}

	next, _ := s.Update(message.RestoreReaderPage{Entry: entry})

	updated, ok := next.(standalone)
	require.True(t, ok)
	assert.Same(t, factory.view, updated.view)

	require.Len(t, factory.entries, 1)
	assert.True(t, factory.entries[0].Story)
}

func TestSetLinesCountsAndPads(t *testing.T) {
	s := Scroller{Viewport: NewViewport(80, 10)}

	s.SetLines([]string{"one", "two", "three"})

	assert.Equal(t, 3, s.ContentLines)
	assert.Equal(t, 13, s.Viewport.TotalLineCount(),
		"one viewport height of padding lets jump targets scroll to the top")
}

func TestPlainLinesStripsStyling(t *testing.T) {
	s := Scroller{Viewport: NewViewport(80, 10)}

	s.SetLines([]string{ansi.Red + "foo" + ansi.Reset, "bar"})
	assert.Equal(t, []string{"foo", "bar"}, s.PlainLines())

	s.SetLines([]string{ansi.Faint + "baz" + ansi.Reset})
	assert.Equal(t, []string{"baz"}, s.PlainLines(), "a content change invalidates the cache")
}

// Match cells are computed from plain-line byte offsets, so stripping must
// remove every escape form the renderers emit — hyperlink URLs especially,
// which would otherwise both false-match queries and shift offsets.
func TestPlainLinesStripsHyperlinksAndTruecolor(t *testing.T) {
	s := Scroller{Viewport: NewViewport(80, 10)}

	s.SetLines([]string{
		"see " + ansi.Hyperlink("https://example.com/secret", "the link") + " here",
		"\x1b[38;2;255;128;0mhalf\x1b[48;2;0;0;0mblock\x1b[0m",
	})

	assert.Equal(t, []string{"see the link here", "halfblock"}, s.PlainLines())
}

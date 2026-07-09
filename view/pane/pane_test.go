package pane

import (
	"errors"
	"image/color"
	"testing"

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
	for _, quit := range []tea.Msg{message.CommentViewQuit{}, message.ReaderViewQuit{}} {
		s := standalone{view: &fakeView{}}

		_, cmd := s.Update(quit)
		require.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd(), "%T should end the program", quit)
	}
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

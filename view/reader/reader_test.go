package reader

import (
	"image/color"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_PreRenderedContent(t *testing.T) {
	content := "line1\nline2\nline3"
	m := newFromContent(content, "Test Title", 100, 24)

	assert.Equal(t, "Test Title", m.title)
	assert.Equal(t, 100, m.paneWidth)
	assert.Nil(t, m.parsed, "pre-rendered model should not have parsed")
}

func TestNewWithArticle_StoresForRerender(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Options{URL: "https://example.com"},
		func(int) string { return "injected header" })

	assert.NotNil(t, m.parsed, "should retain parsed for re-rendering")
	assert.Equal(t, 72, m.maxWidth)
	assert.Equal(t, "Article", m.title)
	assert.Contains(t, m.Viewport.View(), "injected header", "the injected header opens the article")
}

func TestResize_RerenderChangesContent(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Options{URL: "https://example.com"}, nil)

	contentBefore := m.Viewport.View()

	m.Update(tea.WindowSizeMsg{Width: 60, Height: 30})

	assert.Equal(t, 60, m.paneWidth)
	assert.Equal(t, 30-layout.PaneChromeHeight, m.Viewport.Height())

	contentAfter := m.Viewport.View()
	assert.NotEqual(t, contentBefore, contentAfter, "content should change after resize")
}

func TestResize_PreRendered_NoRerender(t *testing.T) {
	m := newFromContent("line1\nline2\nline3", "Title", 80, 24)

	contentBefore := m.ContentLines

	m.Update(tea.WindowSizeMsg{Width: 60, Height: 30})

	assert.Equal(t, contentBefore, m.ContentLines)
	assert.Equal(t, 60, m.paneWidth)
}

func TestResize_PreservesScrollPosition(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Options{URL: "https://example.com"}, nil)

	m.Viewport.SetYOffset(5)

	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Scroll position should be preserved (clamped to max if content shrank).
	assert.GreaterOrEqual(t, m.Viewport.YOffset(), 0)
	maxOffset := max(0, m.ContentLines-m.Viewport.Height())
	assert.LessOrEqual(t, m.Viewport.YOffset(), maxOffset)
}

func TestQuit_ReturnsReaderViewQuitMsg(t *testing.T) {
	keys := []tea.KeyPressMsg{
		{Code: 'q', Text: "q"},
		{Code: tea.KeyEsc},
		{Code: tea.KeyBackspace},
	}

	for _, key := range keys {
		m := newFromContent("content", "Title", 80, 24)

		cmd := m.Update(key)
		require.NotNil(t, cmd)

		msg := cmd()
		_, ok := msg.(message.ReaderViewQuit)
		assert.True(t, ok, "back key should produce ReaderViewQuit")
	}
}

func TestHelpToggle(t *testing.T) {
	m := newFromContent("content", "Title", 80, 24)
	assert.False(t, m.showHelp)

	m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	assert.True(t, m.showHelp)

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.True(t, m.showHelp, "help should remain open on unrelated key")

	m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	assert.False(t, m.showHelp)
}

func TestHeaderLines_DetectsSectionMarkers(t *testing.T) {
	content := "intro\n■ Section One\nbody\n■ Section Two\nend"
	m := newFromContent(content, "Title", 80, 24)

	assert.Len(t, m.headerLines, 2)
}

func TestJumpToHeader(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "text"
	}

	lines[10] = "■ First"
	lines[30] = "■ Second"

	m := newFromContent(strings.Join(lines, "\n"), "Title", 80, 60)
	require.Len(t, m.headerLines, 2)

	m.jumpToHeader(1)
	assert.Equal(t, 10, m.Viewport.YOffset())

	m.jumpToHeader(1)
	assert.Equal(t, 30, m.Viewport.YOffset())

	m.jumpToHeader(-1)
	assert.Equal(t, 10, m.Viewport.YOffset())
}

func TestReader_HideShowImagesToggle(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Options{Images: true}, nil)

	require.True(t, m.showImages, "starts shown when the flag enabled it")

	m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	assert.False(t, m.showImages, "h hides images")

	m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	assert.True(t, m.showImages, "l shows images")
}

func TestImageIndicator_BlankWithoutImages(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Options{Images: true}, nil)

	assert.Empty(t, m.imageIndicator(), "articles without images get no status line")
}

func TestImageStatusLine(t *testing.T) {
	shown := imageStatusLine(true, false, 80)
	assert.Contains(t, shown, "▣ ")
	assert.NotContains(t, shown, "▣  ", "unicode glyphs are single-cell and need no extra room")
	assert.Contains(t, shown, "images shown")

	hidden := imageStatusLine(false, false, 80)
	assert.Contains(t, hidden, "▢ ")
	assert.NotContains(t, hidden, "▢  ")
	assert.Contains(t, hidden, "images hidden")

	assert.Contains(t, imageStatusLine(true, true, 80), nerdfonts.Image+"  ", "wide nerd font glyphs get extra room")
	assert.Contains(t, imageStatusLine(false, true, 80), nerdfonts.ImageOff+"  ")
}

func TestReader_BackgroundColorMsgRerendersWithTermBG(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Options{Images: true}, nil)

	require.Nil(t, m.termBG)

	m.Update(tea.BackgroundColorMsg{Color: color.White})

	assert.Equal(t, color.White, m.termBG, "a late terminal report reaches the renderer")
}

func TestRemapYOffset(t *testing.T) {
	// Three blocks after a 3-line header. In the new render the middle block
	// (an image, lines 10-18) collapsed to a single label line.
	oldStarts := []int{3, 10, 20}
	newStarts := []int{3, 10, 12}

	const newTotal = 22

	assert.Equal(t, 1, remapYOffset(1, oldStarts, newStarts, newTotal), "header lines keep their offset")
	assert.Equal(t, 5, remapYOffset(5, oldStarts, newStarts, newTotal), "a line in an unchanged block keeps its offset")
	assert.Equal(t, 9, remapYOffset(9, oldStarts, newStarts, newTotal), "the separator after an unchanged block keeps its offset")
	assert.Equal(t, 10, remapYOffset(10, oldStarts, newStarts, newTotal), "the start of the shrunk block maps to its start")
	assert.Equal(t, 10, remapYOffset(15, oldStarts, newStarts, newTotal), "deep inside the shrunk block clamps to its last line")
	assert.Equal(t, 17, remapYOffset(25, oldStarts, newStarts, newTotal), "a block after the shrunk one shifts up with it")

	// The reverse toggle: the label grows back into the tall image.
	assert.Equal(t, 10, remapYOffset(10, newStarts, oldStarts, 30), "the label maps back to the top of the image")
	assert.Equal(t, 25, remapYOffset(17, newStarts, oldStarts, 30), "text after the image shifts back down with it")
}

func parseTestArticle(t *testing.T) *article.Parsed {
	t.Helper()

	return article.NewParsedFromHTML("<h1>Hello</h1>" +
		"<p>This is a test paragraph with enough words to cause wrapping at narrow widths.</p>" +
		"<h2>Second Section</h2>" +
		"<p>Another paragraph here.</p>")
}

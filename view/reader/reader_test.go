package reader

import (
	"image/color"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
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

func TestHeaderLines_ComeFromArticleStructure(t *testing.T) {
	parsed := article.NewParsedFromHTML("<p>intro</p><h2>Section One</h2><p>body</p><h2>Section Two</h2><p>end</p>")
	m := NewWithArticle(parsed, "Title", 72, 80, 24, Options{}, nil)

	require.Len(t, m.headerLines, 2)

	plain := m.PlainLines()
	assert.Contains(t, plain[m.headerLines[0]], "Section One")
	assert.Contains(t, plain[m.headerLines[1]], "Section Two")
}

func TestHeaderLines_SurviveRerender(t *testing.T) {
	parsed := article.NewParsedFromHTML("<p>intro</p><h2>Section</h2><p>body</p>")
	m := NewWithArticle(parsed, "Title", 72, 80, 24, Options{}, nil)

	m.Update(tea.WindowSizeMsg{Width: 60, Height: 30})

	require.Len(t, m.headerLines, 1)
	assert.Contains(t, m.PlainLines()[m.headerLines[0]], "Section")
}

func TestJumpToHeader(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "text"
	}

	m := newFromContent(strings.Join(lines, "\n"), "Title", 80, 60)
	m.headerLines = []int{10, 30}

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

// The reader hands its meta block the same left margin it gives the article
// text, so the block's frame opens exactly where prose starts. The block's
// right-edge arithmetic is meta's TestBlockGeometryContract; this pins the
// plumbing — one margin shared by the block and the article.
func TestMetaBlockAlignsWithArticleColumn(t *testing.T) {
	block := meta.ReaderMode(meta.Data{
		URL: "https://example.com/story", Author: "alice", TimeAgo: "1 hour ago",
	})

	m := NewWithArticle(parseTestArticle(t), "Article", 72, 120, 40, Options{}, block.Render)

	openCol, ruleCol, proseCol := -1, -1, -1

	for line := range strings.SplitSeq(m.Viewport.View(), "\n") {
		// The viewport pads rows to the pane width; only the leading columns
		// matter here.
		s := strings.TrimRight(xansi.Strip(line), " ")
		trimmed := strings.TrimLeft(s, " ")

		switch {
		case openCol == -1 && strings.HasPrefix(trimmed, "╭"):
			openCol = len(s) - len(trimmed)

			assert.Contains(t, trimmed, "by alice", "the opening rule must carry the byline")
		case strings.HasPrefix(trimmed, "╰"):
			ruleCol = len(s) - len(trimmed)
		case strings.HasPrefix(trimmed, "This is a test paragraph"):
			proseCol = len(s) - len(trimmed)
		}
	}

	require.NotEqual(t, -1, openCol, "no meta block opening rule in the view")
	require.NotEqual(t, -1, ruleCol, "no closing rule in the view")
	require.NotEqual(t, -1, proseCol, "no article prose in the view")

	assert.Equal(t, layout.ReaderViewLeftMargin, openCol, "the opening rule must open at the reader margin")
	assert.Equal(t, layout.ReaderViewLeftMargin, ruleCol, "the closing rule must open at the reader margin")
	assert.Equal(t, layout.ReaderViewLeftMargin, proseCol, "article prose must start at the reader margin")
}

func searchableReader(t *testing.T) *Model {
	t.Helper()

	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "filler text"
	}

	lines[5] = "the needle is here"
	lines[25] = "another needle below"

	return newFromContent(strings.Join(lines, "\n"), "Title", 80, 20)
}

func commitSearch(m *Model, query string) {
	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})

	for _, r := range query {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
}

func TestReaderSearch_CommitJumpsToFirstMatch(t *testing.T) {
	m := searchableReader(t)

	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	assert.True(t, m.SearchPrompting())

	for _, r := range "needle" {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.True(t, m.SearchActive())
	require.Len(t, m.SearchMatches(), 2)
	assert.Equal(t, 3, m.Viewport.YOffset(), "first match line 5 sits two lines below the top")
}

func TestReaderSearch_NCyclesMatchesNotSections(t *testing.T) {
	m := searchableReader(t)
	m.headerLines = []int{15}

	commitSearch(m, "needle")
	assert.Equal(t, 3, m.Viewport.YOffset())

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	assert.Equal(t, 23, m.Viewport.YOffset(), "n goes to the next match, not the section at 15")

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	assert.Equal(t, 3, m.Viewport.YOffset(), "n wraps around")

	m.Update(tea.KeyPressMsg{Code: 'N', Text: "N"})
	assert.Equal(t, 23, m.Viewport.YOffset())
}

func TestReaderSearch_NFallsBackToSectionsWhenInactive(t *testing.T) {
	m := searchableReader(t)
	m.headerLines = []int{15}

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

	assert.Equal(t, 15, m.Viewport.YOffset())
}

func TestReaderSearch_EscClearsThenQuits(t *testing.T) {
	m := searchableReader(t)
	commitSearch(m, "needle")

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Nil(t, cmd)
	assert.False(t, m.SearchActive(), "the first esc only clears the search")

	cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, cmd)
	assert.IsType(t, message.ReaderViewQuit{}, cmd(), "the second esc quits")
}

func TestReaderSearch_SurvivesRerender(t *testing.T) {
	parsed := article.NewParsedFromHTML("<p>alpha needle</p><h2>Head</h2><p>beta needle</p>")
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, nil)

	commitSearch(m, "needle")
	require.Len(t, m.SearchMatches(), 2)

	m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})

	assert.True(t, m.SearchActive())
	assert.Len(t, m.SearchMatches(), 2, "matches are recomputed against the rewrapped text")
}

func TestResize_HeightOnlySkipsRerender(t *testing.T) {
	parsed := article.NewParsedFromHTML("<h2>One</h2><p>body text</p><h2>Two</h2>")
	m := NewWithArticle(parsed, "Title", 72, 80, 24, Options{}, nil)

	before := &m.blockStarts[0]

	m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	assert.Same(t, before, &m.blockStarts[0], "height-only resize must not re-render the article")
	assert.Equal(t, m.ContentLines+m.Viewport.Height(), m.Viewport.TotalLineCount(),
		"the bottom padding tracks the new height")

	m.Update(tea.WindowSizeMsg{Width: 60, Height: 40})
	assert.NotSame(t, before, &m.blockStarts[0], "a width change re-renders")
}

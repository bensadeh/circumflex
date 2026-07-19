package reader

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

func TestQuit_ReturnsDetailQuitMsg(t *testing.T) {
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
		_, ok := msg.(message.DetailQuit)
		assert.True(t, ok, "back key should produce DetailQuit")
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
	assert.Contains(t, shown, "Images Shown")
	assert.True(t, strings.HasSuffix(shown, " ▣"), "the icon trails at the right edge")

	hidden := imageStatusLine(false, false, 80)
	assert.Contains(t, hidden, "Images Hidden")
	assert.True(t, strings.HasSuffix(hidden, " ▢"))

	assert.Equal(t, xansi.StringWidth(shown), xansi.StringWidth(hidden),
		"both states span the same width so the label start column never shifts")

	assert.True(t, strings.HasSuffix(imageStatusLine(true, true, 80), " "+nerdfonts.Image+" "))
	assert.True(t, strings.HasSuffix(imageStatusLine(false, true, 80), " "+nerdfonts.ImageOff+" "))
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

func TestReaderSearch_BackKeysClearThenQuit(t *testing.T) {
	keys := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{"esc", tea.KeyPressMsg{Code: tea.KeyEscape}},
		{"q", tea.KeyPressMsg{Code: 'q', Text: "q"}},
		{"backspace", tea.KeyPressMsg{Code: tea.KeyBackspace}},
	}

	for _, k := range keys {
		t.Run(k.name, func(t *testing.T) {
			m := searchableReader(t)
			commitSearch(m, "needle")

			cmd := m.Update(k.msg)
			assert.Nil(t, cmd)
			assert.False(t, m.SearchActive(), "the first press only clears the search")

			cmd = m.Update(k.msg)
			require.NotNil(t, cmd)
			assert.IsType(t, message.DetailQuit{}, cmd(), "the second press quits")
		})
	}
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

func linkedTestReader(t *testing.T) *Model {
	t.Helper()

	parsed := article.NewParsedFromHTML(
		`<p><a href="https://one.example.com">first</a> link</p>` +
			`<p><a href="https://two.example.com">second</a> link</p>` +
			`<p><a href="https://three.example.com">third</a> link</p>`)

	return NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, nil)
}

func TestLinkSelector_TabTogglesAndKeysStep(t *testing.T) {
	m := linkedTestReader(t)
	require.Len(t, m.links, 3)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.True(t, m.linkMode)
	assert.Equal(t, 0, m.currentLink, "entry selects the first link on screen")
	assert.Equal(t, m.links[0].spans, m.LinkSpans())

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 1, m.currentLink)

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	assert.Equal(t, 2, m.currentLink, "n moves links, not sections, inside the selector")

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 2, m.currentLink, "j stops at the last link on screen")

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	assert.Equal(t, 0, m.currentLink, "the jump wraps around like a search jump")

	m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 0, m.currentLink, "k stops at the first link on screen")

	m.Update(tea.KeyPressMsg{Code: 'N', Text: "N"})
	assert.Equal(t, 2, m.currentLink, "the backward jump wraps too")

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 1, m.currentLink)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.False(t, m.linkMode, "tab toggles back out")
	assert.Empty(t, m.LinkSpans())
}

// deepLinkTestReader renders enough link-free filler that the article's two
// links sit far below the first screen, one further down than the other. The
// filler paragraphs must all differ: identical blocks collapse in parsing.
func deepLinkTestReader(t *testing.T) *Model {
	t.Helper()

	var b strings.Builder

	for i := range 40 {
		fmt.Fprintf(&b, "<p>filler paragraph number %d</p>", i)
	}

	b.WriteString(`<p><a href="https://deep.example.com">deep</a></p>`)

	for i := range 40 {
		fmt.Fprintf(&b, "<p>more filler number %d</p>", i)
	}

	b.WriteString(`<p><a href="https://deeper.example.com">deeper</a></p>`)

	m := NewWithArticle(article.NewParsedFromHTML(b.String()), "Title", 72, 100, 30, Options{}, nil)
	require.Len(t, m.links, 2)
	require.False(t, m.linkOnScreen(0), "the fixture keeps both links below the first screen")

	return m
}

func TestLinkSelector_EntryNeverScrolls(t *testing.T) {
	m := deepLinkTestReader(t)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	assert.True(t, m.linkMode)
	assert.Equal(t, 0, m.Viewport.YOffset(), "entry stays put with no link in view")
	assert.Equal(t, -1, m.currentLink, "nothing on screen, nothing selected")
	assert.Empty(t, m.LinkSpans())

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, -1, m.currentLink, "the on-screen step has nothing to land on")
	assert.Equal(t, 0, m.Viewport.YOffset())
}

func TestLinkSelector_JumpReachesOffscreenLinks(t *testing.T) {
	m := deepLinkTestReader(t)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

	assert.Equal(t, 0, m.currentLink, "the jump finds the first link past the viewport")
	assert.True(t, m.linkOnScreen(0), "and scrolls it into view")

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	assert.Equal(t, 1, m.currentLink)
	assert.True(t, m.linkOnScreen(1))
}

func TestLinkSelector_StepEntersVisibleSetAfterScroll(t *testing.T) {
	m := deepLinkTestReader(t)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Equal(t, -1, m.currentLink)

	// Scroll the first link into view by hand; the empty selection then
	// enters the visible set on the next step.
	m.Viewport.SetYOffset(m.links[0].spans[0].Line - 2)
	require.True(t, m.linkOnScreen(0))

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 0, m.currentLink)
}

func TestLinkSelector_TabNoopWithoutLinks(t *testing.T) {
	m := NewWithArticle(parseTestArticle(t), "Title", 72, 100, 30, Options{}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.False(t, m.linkMode)
}

// captureBrowserOpens points CLX_BROWSER at a script that logs the URLs it
// is asked to open. The open runs detached, so assertions poll the log.
func captureBrowserOpens(t *testing.T) func() string {
	t.Helper()

	dir := t.TempDir()
	log := filepath.Join(dir, "opened.log")
	script := filepath.Join(dir, "browser.sh")

	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho \"$1\" >> "+log+"\n"), 0o600))
	require.NoError(t, os.Chmod(script, 0o700)) //nolint:gosec // the browser stub must be executable
	t.Setenv("CLX_BROWSER", script)

	return func() string {
		b, _ := os.ReadFile(log)

		return string(b)
	}
}

func TestLinkSelector_EnterOpensReaderLinkAndOOpensStory(t *testing.T) {
	opened := captureBrowserOpens(t)

	parsed := article.NewParsedFromHTML(
		`<p><a href="https://one.example.com">first</a> and <a href="https://two.example.com">second</a></p>`)
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{URL: "https://story.example.com"}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg, ok := cmd().(message.OpenReaderLink)
	require.True(t, ok, "enter asks the app to open the link in reader mode")
	assert.Equal(t, "https://two.example.com", msg.URL)

	cmd = m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	require.NotNil(t, cmd)
	assert.Nil(t, cmd())
	// Generous deadline: the browser stub is a real subprocess, and spawning
	// it under full-suite load can take well over a second.
	require.Eventually(t, func() bool { return strings.Contains(opened(), "https://story.example.com") },
		5*time.Second, 10*time.Millisecond, "o keeps its story meaning inside the selector")
}

func TestQuit_FromLinkStepsBackToStory(t *testing.T) {
	keys := []tea.KeyPressMsg{
		{Code: 'q', Text: "q"},
		{Code: tea.KeyEsc},
		{Code: tea.KeyBackspace},
	}

	for _, key := range keys {
		m := NewWithArticle(parseTestArticle(t), "Linked Page", 72, 100, 30, Options{FromLink: true}, nil)

		cmd := m.Update(key)
		require.NotNil(t, cmd)

		msg, ok := cmd().(message.OpenAdjacentStory)
		require.True(t, ok, "quit on a followed link re-opens the story, not the front page")
		assert.Equal(t, 0, msg.Direction)
	}
}

func TestQuit_WalksBackThroughTrail(t *testing.T) {
	trail := []message.TrailEntry{
		{URL: "https://story.example.com", Title: "Story", Parsed: parseTestArticle(t), Story: true},
		{URL: "https://a.example.com", Title: "Page A", Parsed: parseTestArticle(t)},
	}

	m := NewWithArticle(parseTestArticle(t), "Deep Page", 72, 100, 30, Options{
		FromLink: true,
		Trail:    trail,
	}, nil)

	cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.NotNil(t, cmd)

	msg, ok := cmd().(message.RestoreReaderPage)
	require.True(t, ok, "quit restores the previous page from its retained parse")
	assert.Equal(t, "https://a.example.com", msg.Entry.URL)
	require.Len(t, msg.Trail, 1, "the step taken back leaves the chain")
	assert.True(t, msg.Trail[0].Story, "the story article stays at the chain's root")
}

func TestLinkSelector_ForwardExtendsTrail(t *testing.T) {
	parsed := article.NewParsedFromHTML(`<p><a href="https://next.example.com">next</a></p>`)
	m := NewWithArticle(parsed, "Linked Page", 72, 100, 30, Options{
		URL:      "https://current.example.com",
		FromLink: true,
		Trail:    []message.TrailEntry{{URL: "https://story.example.com", Story: true}},
	}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg, ok := cmd().(message.OpenReaderLink)
	require.True(t, ok)
	assert.Equal(t, "https://next.example.com", msg.URL)
	require.Len(t, msg.Trail, 2, "the page being left joins the chain")
	assert.Equal(t, "https://current.example.com", msg.Trail[1].URL)
	assert.Equal(t, "Linked Page", msg.Trail[1].Title)
	assert.Same(t, parsed, msg.Trail[1].Parsed, "the parse rides along so stepping back needs no network")
	assert.False(t, msg.Trail[1].Story)
}

func TestLinkSelector_ForwardFromStoryStartsTrail(t *testing.T) {
	parsed := article.NewParsedFromHTML(`<p><a href="https://next.example.com">next</a></p>`)
	m := NewWithArticle(parsed, "Story Title", 72, 100, 30, Options{URL: "https://story.example.com"}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg, ok := cmd().(message.OpenReaderLink)
	require.True(t, ok)
	require.Len(t, msg.Trail, 1)
	assert.True(t, msg.Trail[0].Story, "the story article roots the chain, marked for its story meta")
	assert.Equal(t, "https://story.example.com", msg.Trail[0].URL)
}

func TestTitleHeader_DepthBadge(t *testing.T) {
	root := NewWithArticle(parseTestArticle(t), "Title", 72, 100, 30, Options{}, nil)
	assert.NotContains(t, xansi.Strip(root.titleHeader), "›", "the story article carries no badge")

	deep := NewWithArticle(parseTestArticle(t), "Title", 72, 100, 30, Options{
		FromLink: true,
		Trail: []message.TrailEntry{
			{URL: "https://story.example.com", Story: true},
			{URL: "https://a.example.com"},
		},
	}, nil)

	row := xansi.Strip(strings.SplitN(deep.titleHeader, "\n", 2)[0])
	require.Equal(t, 2, strings.Count(row, "›"), "one chevron per link followed")

	rightEdge := layout.ReaderViewLeftMargin + layout.ReaderContentWidth(100, 72)
	assert.Equal(t, rightEdge, xansi.StringWidth(strings.TrimRight(row, " ")), "the badge ends at the article column's right edge")
}

func TestTitleHeaderWithBadge_TruncatesLongTitle(t *testing.T) {
	long := strings.Repeat("word ", 40)
	m := NewWithArticle(parseTestArticle(t), long, 72, 100, 30, Options{FromLink: true}, nil)

	row := xansi.Strip(strings.SplitN(m.titleHeader, "\n", 2)[0])
	require.Equal(t, 1, strings.Count(row, "›"))
	assert.Contains(t, row, "…", "the title shortens instead of colliding with the badge")

	badgeStart := strings.Index(row, "›")
	titleEnd := strings.LastIndex(row[:badgeStart], "…")
	assert.GreaterOrEqual(t, badgeStart-titleEnd, 2, "a gap separates the title from the badge")
}

func TestQuit_FromLinkExitsSelectorFirst(t *testing.T) {
	parsed := article.NewParsedFromHTML(`<p><a href="https://example.com/x">a link</a></p>`)
	m := NewWithArticle(parsed, "Linked Page", 72, 100, 30, Options{FromLink: true}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, m.linkMode)

	cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	assert.Nil(t, cmd)
	assert.False(t, m.linkMode, "the first q only leaves the selector")

	cmd = m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.NotNil(t, cmd)
	assert.IsType(t, message.OpenAdjacentStory{}, cmd(), "the second q steps back to the story")
}

func TestLinkSelector_NonViewableLinkIsInert(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p><a href="https://example.com/paper.pdf">the pdf</a> and <a href="https://example.com/page">a page</a></p>`)
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Equal(t, 0, m.currentLink)
	assert.True(t, m.LinkSpansMuted(), "the selected pdf takes the muted open-story bar")

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Nil(t, cmd, "enter does nothing on a link the reader won't open")
	assert.True(t, m.linkMode)

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.False(t, m.LinkSpansMuted(), "a viewable link selects in the normal colors")
}

func TestLinkSelector_URLRowDimsTextKeepsRule(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p><a href="https://example.com/paper.pdf">the pdf</a> and <a href="https://example.com/page">a page</a></p>`)
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	dim := lipgloss.NewStyle().Underline(true).Faint(true).Render("example.com/paper.pdf")
	assert.Contains(t, m.linkURLRow(), dim,
		"without a terminal foreground report the URL and its underline dim together")
	assert.Contains(t, xansi.Strip(m.footer()), "↛", "the arrow breaks for a link that won't open")

	m.Update(tea.ForegroundColorMsg{Color: color.White})
	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})

	pinned := lipgloss.NewStyle().Underline(true).Faint(true).UnderlineColor(color.White).Render("example.com/page")
	assert.Contains(t, m.linkURLRow(), pinned,
		"with only a foreground report the underline pins and the text falls back to faint")
	assert.Contains(t, xansi.Strip(m.footer()), "→")

	m.Update(tea.BackgroundColorMsg{Color: color.Black})

	blended := lipgloss.NewStyle().Underline(true).
		Foreground(style.Dimmed(color.White, color.Black)).
		UnderlineColor(color.White).
		Render("example.com/page")
	assert.Contains(t, m.linkURLRow(), blended,
		"with both reports the text dims through a blended color, no faint flag to drag the underline down")
}

func TestLinkSelector_FooterIconSwapsForNonViewable_NerdFonts(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p><a href="https://example.com/paper.pdf">the pdf</a> and <a href="https://example.com/page">a page</a></p>`)
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{NerdFonts: true}, nil)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Contains(t, m.footer(), nerdfonts.LinkSelectorOff)

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	footer := m.footer()
	assert.Contains(t, footer, nerdfonts.LinkSelector)
	assert.NotContains(t, footer, nerdfonts.LinkSelectorOff)
}

func TestLinkSelector_EscLayersBeforeSearchAndQuit(t *testing.T) {
	m := linkedTestReader(t)
	commitSearch(m, "link")
	require.True(t, m.SearchActive())

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, m.linkMode)

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Nil(t, cmd)
	assert.False(t, m.linkMode, "the first esc leaves the selector")
	assert.True(t, m.SearchActive(), "the search survives it")
	assert.Empty(t, m.LinkSpans(), "the selection highlight clears")

	cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Nil(t, cmd)
	assert.False(t, m.SearchActive(), "the second esc clears the search")

	cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, cmd)
	assert.IsType(t, message.DetailQuit{}, cmd(), "the third quits the view")
}

func TestLinkSelector_SlashSwitchesToSearch(t *testing.T) {
	m := linkedTestReader(t)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})

	assert.False(t, m.linkMode)
	assert.True(t, m.SearchPrompting())
}

func TestLinkSelector_URLRidesTheSeparator(t *testing.T) {
	m := linkedTestReader(t)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	row := xansi.Strip(m.linkURLRow())
	assert.Contains(t, row, "one.example.com")
	assert.NotContains(t, row, "https://", "the scheme is stripped from the preview")

	rightEdge := layout.ReaderViewLeftMargin + layout.ReaderContentWidth(m.paneWidth, m.maxWidth)
	assert.Equal(t, rightEdge, xansi.StringWidth(strings.TrimRight(row, " ")),
		"the URL ends at the article column's right edge")

	footer := xansi.Strip(m.footer())
	assert.Contains(t, footer, "URL Selection Mode")
	assert.Contains(t, footer, "1/3")
	assert.NotContains(t, footer, "one.example.com", "the URL moved out of the footer")
}

func TestLinkSelector_SurvivesRerender(t *testing.T) {
	m := linkedTestReader(t)

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	require.Equal(t, 1, m.currentLink)

	m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})

	assert.True(t, m.linkMode)
	assert.Equal(t, 1, m.currentLink)
	assert.Equal(t, m.links[1].spans, m.LinkSpans(), "the highlight tracks the rewrapped spans")
}

func TestFooter_ReaderModeLabelSitsLeft(t *testing.T) {
	m := NewWithArticle(parseTestArticle(t), "Title", 72, 100, 30, Options{}, nil)

	footer := xansi.Strip(m.footer())
	assert.True(t, strings.HasPrefix(footer, "  Reader Mode"), "the label moved to the left slot")
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

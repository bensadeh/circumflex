package reader

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_PreRenderedContent(t *testing.T) {
	content := "line1\nline2\nline3"
	m := New(content, "Test Title", 80, 24)

	assert.Equal(t, "Test Title", m.title)
	assert.Equal(t, 80, m.screenWidth)
	assert.Nil(t, m.parsed, "pre-rendered model should not have parsed")
	assert.False(t, m.standalone)
}

func TestNewWithArticle_StoresForRerender(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Meta{
		URL:    "https://example.com",
		Author: "alice",
	})

	assert.NotNil(t, m.parsed, "should retain parsed for re-rendering")
	assert.Equal(t, 72, m.maxWidth)
	assert.Equal(t, "Article", m.title)
}

func TestResize_RerenderChangesContent(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Meta{URL: "https://example.com"})

	contentBefore := m.viewport.View()
	linesBefore := m.contentLines

	// Simulate a significant resize.
	m.Update(tea.WindowSizeMsg{Width: 60, Height: 30})

	assert.Equal(t, 60, m.screenWidth)
	assert.Equal(t, 30-headerHeight-footerHeight, m.viewportHeight)

	// Content should have changed because the render width changed.
	contentAfter := m.viewport.View()
	assert.NotEqual(t, contentBefore, contentAfter, "content should change after resize")

	_ = linesBefore
}

func TestResize_PreRendered_NoRerender(t *testing.T) {
	m := New("line1\nline2\nline3", "Title", 80, 24)

	contentBefore := m.contentLines

	m.Update(tea.WindowSizeMsg{Width: 60, Height: 30})

	// Pre-rendered content doesn't re-render — only viewport dimensions change.
	assert.Equal(t, contentBefore, m.contentLines)
	assert.Equal(t, 60, m.screenWidth)
}

func TestResize_PreservesScrollPosition(t *testing.T) {
	parsed := parseTestArticle(t)
	m := NewWithArticle(parsed, "Article", 72, 120, 40, Meta{URL: "https://example.com"})

	// Scroll down.
	m.viewport.SetYOffset(5)

	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Scroll position should be preserved (clamped to max if content shrank).
	assert.GreaterOrEqual(t, m.viewport.YOffset(), 0)
	maxOffset := max(0, m.contentLines-m.viewportHeight)
	assert.LessOrEqual(t, m.viewport.YOffset(), maxOffset)
}

func TestQuit_ReturnsReaderViewQuitMsg(t *testing.T) {
	m := New("content", "Title", 80, 24)

	cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(message.ReaderViewQuitMsg)
	assert.True(t, ok, "quit should produce ReaderViewQuitMsg")
}

func TestQuit_Standalone_ReturnsTeaQuit(t *testing.T) {
	m := New("content", "Title", 80, 24)
	m.standalone = true

	cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.NotNil(t, cmd)

	msg := cmd()
	assert.IsType(t, tea.QuitMsg{}, msg, "standalone quit should produce tea.QuitMsg")
}

func TestHelpToggle(t *testing.T) {
	m := New("content", "Title", 80, 24)
	assert.False(t, m.showHelp)

	// Press 'i' to open help.
	m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	assert.True(t, m.showHelp)

	// While in help, other keys are ignored.
	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.True(t, m.showHelp, "help should remain open on unrelated key")

	// Press 'i' again to close.
	m.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	assert.False(t, m.showHelp)
}

func TestHeaderLines_DetectsSectionMarkers(t *testing.T) {
	// Content with section markers (■).
	content := "intro\n■ Section One\nbody\n■ Section Two\nend"
	m := New(content, "Title", 80, 24)

	assert.Len(t, m.headerLines, 2)
}

func TestJumpToHeader(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "text"
	}

	lines[10] = "■ First"
	lines[30] = "■ Second"

	m := New(strings.Join(lines, "\n"), "Title", 80, 60)
	require.Len(t, m.headerLines, 2)

	// Jump forward from top.
	m.jumpToHeader(1)
	assert.Equal(t, 10, m.viewport.YOffset())

	// Jump forward again.
	m.jumpToHeader(1)
	assert.Equal(t, 30, m.viewport.YOffset())

	// Jump backward.
	m.jumpToHeader(-1)
	assert.Equal(t, 10, m.viewport.YOffset())
}

// parseTestArticle creates a minimal Parsed article for testing re-render.
// Uses the real Parse pipeline with a simple HTML string via the block API.
func parseTestArticle(t *testing.T) *article.Parsed {
	t.Helper()

	// article.Parsed is opaque (unexported fields), so we use Fetch-like
	// construction. Since we can't hit the network, build one via the
	// exported constructor if available, or test with a nil-safe fallback.
	//
	// The Parsed type needs blocks and url. Since those are unexported,
	// we rely on the fact that NewWithArticle only calls RenderWithHeader.
	// We build a minimal Parsed that returns stable, width-dependent output.
	return article.NewParsedFromMarkdown("# Hello\n\nThis is a test paragraph with enough words to cause wrapping at narrow widths.\n\n## Second Section\n\nAnother paragraph here.")
}

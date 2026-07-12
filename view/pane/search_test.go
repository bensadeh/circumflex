package pane

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/nerdfonts"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Default-theme search highlights: yellow text for matches, black on a
// yellow background for the current match, reverse/intensity cleared on both.
const (
	matchHighlight   = "\x1b[27m\x1b[22m\x1b[33m"
	currentHighlight = "\x1b[27m\x1b[22m\x1b[43;30m"
)

func typeKeys(s *Scroller, text string) {
	for _, r := range text {
		s.HandleSearchPromptKey(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func TestFindMatches_SmartCase(t *testing.T) {
	t.Parallel()

	lines := []string{"Foo bar foo", "no hits here", "FOO"}

	caseless := FindMatches(lines, "foo")
	require.Len(t, caseless, 3, "an all-lowercase query matches any case")
	assert.Equal(t, Match{Line: 0, StartCell: 0, EndCell: 3}, caseless[0])
	assert.Equal(t, Match{Line: 0, StartCell: 8, EndCell: 11}, caseless[1])
	assert.Equal(t, Match{Line: 2, StartCell: 0, EndCell: 3}, caseless[2])

	exact := FindMatches(lines, "Foo")
	require.Len(t, exact, 1, "an uppercase character makes the query exact")
	assert.Equal(t, 0, exact[0].Line)
}

func TestFindMatches_CellOffsetsWithWideRunes(t *testing.T) {
	t.Parallel()

	matches := FindMatches([]string{"aＸbＹ foo"}, "foo")

	require.Len(t, matches, 1)
	assert.Equal(t, Match{Line: 0, StartCell: 7, EndCell: 10}, matches[0],
		"cells account for double-width runes before the match")
}

func TestFindMatches_NonOverlapping(t *testing.T) {
	t.Parallel()

	matches := FindMatches([]string{"aaaa"}, "aa")

	require.Len(t, matches, 2)
	assert.Equal(t, 0, matches[0].StartCell)
	assert.Equal(t, 2, matches[1].StartCell)
}

func TestSearchPrompt_Lifecycle(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}

	s.StartSearchPrompt()
	assert.True(t, s.SearchPrompting())

	typeKeys(s, "foxo")
	s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyBackspace})
	s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyBackspace})
	typeKeys(s, "o")

	result := s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, PromptCommitted, result)
	assert.False(t, s.SearchPrompting())
	assert.Equal(t, "foo", s.SearchQuery())
}

func TestSearchPrompt_EscCancelsKeepingPreviousQuery(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}

	s.StartSearchPrompt()
	typeKeys(s, "first")
	s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	s.StartSearchPrompt()
	typeKeys(s, "second")

	result := s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Equal(t, PromptCanceled, result)
	assert.Equal(t, "first", s.SearchQuery(), "a canceled prompt leaves the previous search alone")
}

func TestSearchPrompt_BackspacePastEmptyCancels(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}

	s.StartSearchPrompt()
	typeKeys(s, "a")
	s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyBackspace})

	result := s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyBackspace})
	assert.Equal(t, PromptCanceled, result)
	assert.False(t, s.SearchPrompting())
}

func TestSearchPrompt_EmptyEnterCancels(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}

	s.StartSearchPrompt()

	result := s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, PromptCanceled, result)
	assert.False(t, s.SearchActive())
}

func searchScroller(t *testing.T, contentLines, height int, matchLines ...int) *Scroller {
	t.Helper()

	s := &Scroller{Viewport: NewViewport(80, height)}

	lines := make([]string, contentLines)
	for i := range lines {
		lines[i] = "text"
	}

	s.SetLines(lines)

	matches := make([]Match, len(matchLines))
	for i, line := range matchLines {
		matches[i] = Match{Line: line, StartCell: 0, EndCell: 4}
	}

	s.search.query = "text"
	s.SetSearchMatches(matches)

	return s
}

func TestMatchNavigation_WrapsBothWays(t *testing.T) {
	t.Parallel()

	s := searchScroller(t, 100, 10, 10, 50, 90)

	s.JumpToFirstMatchFrom(20)
	assert.Equal(t, 48, s.Viewport.YOffset(), "first match at or below line 20 sits two lines under the top")
	assert.Equal(t, "2/3", ansi.Strip(s.SearchCountLabel()))

	s.NextMatch()
	assert.Equal(t, 88, s.Viewport.YOffset())

	s.NextMatch()
	assert.Equal(t, 8, s.Viewport.YOffset(), "n wraps past the last match")

	s.PrevMatch()
	assert.Equal(t, 88, s.Viewport.YOffset(), "N wraps back")
}

func TestJumpToFirstMatchFrom_WrapsToTop(t *testing.T) {
	t.Parallel()

	s := searchScroller(t, 100, 10, 10, 20)

	s.JumpToFirstMatchFrom(50)
	assert.Equal(t, 8, s.Viewport.YOffset(), "no match below the offset wraps to the first")
	assert.Equal(t, "1/2", ansi.Strip(s.SearchCountLabel()))
}

func TestSearchHighlights_AppliedAndCleared(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}
	s.SetLines([]string{"hello world"})

	s.search.query = "world"
	s.SetSearchMatches(FindMatches(s.PlainLines(), "world"))
	assert.Contains(t, s.DecorateView(s.Viewport.View()), currentHighlight, "the only match is the current one")

	s.ClearSearch()
	assert.NotContains(t, s.DecorateView(s.Viewport.View()), currentHighlight)
	assert.NotContains(t, s.DecorateView(s.Viewport.View()), matchHighlight)
	assert.False(t, s.SearchActive())
}

func TestSetCurrentMatch_DistinguishesTiers(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}
	s.SetLines([]string{"first needle", "second needle"})

	s.search.query = "needle"
	s.SetSearchMatches(FindMatches(s.PlainLines(), "needle"))
	s.SetCurrentMatch(1)

	rows := strings.Split(s.DecorateView(s.Viewport.View()), "\n")
	assert.Contains(t, rows[0], matchHighlight)
	assert.Contains(t, rows[1], currentHighlight)

	s.SetCurrentMatch(-1)

	view := s.DecorateView(s.Viewport.View())
	assert.NotContains(t, view, currentHighlight, "-1 renders every match as non-current")
	assert.Contains(t, view, matchHighlight)
}

func TestSearchPrompt_LiveMatchesHaveNoCurrent(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}
	s.SetLines([]string{"hello world"})

	s.StartSearchPrompt()
	typeKeys(s, "world")
	s.SetSearchMatches(FindMatches(s.PlainLines(), s.ActiveQuery()))

	view := s.DecorateView(s.Viewport.View())
	assert.Contains(t, view, matchHighlight, "hits highlight while typing")
	assert.NotContains(t, view, currentHighlight, "no current match before commit")
	assert.Equal(t, "1 match", ansi.Strip(s.SearchCountLabel()), "the counter shows the live total")

	s.HandleSearchPromptKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Contains(t, s.DecorateView(s.Viewport.View()), currentHighlight, "commit promotes a hit to current")
	assert.Equal(t, "1/1", ansi.Strip(s.SearchCountLabel()))
}

func TestSearchCountLabel_NoMatches(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}
	s.SetLines([]string{"hello"})

	s.search.query = "absent"
	s.SetSearchMatches(nil)

	assert.Equal(t, "no matches", ansi.Strip(s.SearchCountLabel()))
	assert.Equal(t, "/absent", ansi.Strip(s.SearchFooterLabel(false)))
}

func TestSearchFooterLabel_PromptShowsCursor(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}

	s.StartSearchPrompt()
	typeKeys(s, "que")

	label := s.SearchFooterLabel(false)
	assert.True(t, strings.HasPrefix(ansi.Strip(label), "/que"))
	assert.Contains(t, label, ansi.Reverse, "the prompt renders a block cursor")

	nerd := s.SearchFooterLabel(true)
	assert.True(t, strings.HasPrefix(nerd, nerdfonts.Search+"  "), "nerd fonts swap the / for a magnifier")
}

func TestDecorateView_OverridesAndWindow(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 3)}
	s.SetLines([]string{"aaa", "bbb", "ccc", "ddd", "eee"})

	s.SetRowOverrides([]RowOverride{{Line: 1, Content: "BBB"}})
	s.search.query = "x"
	s.SetSearchMatches([]Match{{Line: 1, StartCell: 0, EndCell: 3}, {Line: 4, StartCell: 0, EndCell: 3}})

	rows := strings.Split(s.DecorateView(s.Viewport.View()), "\n")
	assert.Contains(t, rows[1], "BBB", "the override replaces the row")
	assert.Contains(t, rows[1], currentHighlight, "the current match decorates on top of the override")
	assert.NotContains(t, rows[0], matchHighlight)
	assert.NotContains(t, rows[2], matchHighlight)

	s.Viewport.SetYOffset(2)

	view := s.DecorateView(s.Viewport.View())
	assert.NotContains(t, view, "BBB", "an override outside the window is skipped")
	assert.Contains(t, view, matchHighlight, "the non-current match at line 4 scrolled into view")
	assert.NotContains(t, view, currentHighlight, "the current match stayed above the window")
}

package pane

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Contains(t, s.Viewport.View(), ansi.Reverse, "matches render reversed")

	s.ClearSearch()
	assert.NotContains(t, s.Viewport.View(), ansi.Reverse)
	assert.False(t, s.SearchActive())
}

func TestSearchCountLabel_NoMatches(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}
	s.SetLines([]string{"hello"})

	s.search.query = "absent"
	s.SetSearchMatches(nil)

	assert.Equal(t, "no matches", ansi.Strip(s.SearchCountLabel()))
	assert.Equal(t, "/absent", ansi.Strip(s.SearchFooterLabel()))
}

func TestSearchFooterLabel_PromptShowsCursor(t *testing.T) {
	t.Parallel()

	s := &Scroller{Viewport: NewViewport(80, 10)}

	s.StartSearchPrompt()
	typeKeys(s, "que")

	label := s.SearchFooterLabel()
	assert.True(t, strings.HasPrefix(ansi.Strip(label), "/que"))
	assert.Contains(t, label, ansi.Reverse, "the prompt renders a block cursor")
}

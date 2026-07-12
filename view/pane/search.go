package pane

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"

	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// Match locates one search hit: a cell span on a content line.
type Match struct {
	Line      int
	StartCell int
	EndCell   int
}

type searchState struct {
	prompting bool
	input     string
	query     string
	matches   []Match
	current   int
}

// PromptResult reports how the search prompt reacted to a key press.
type PromptResult int

const (
	PromptPending PromptResult = iota
	PromptCommitted
	PromptCanceled
)

// matchScrollPadding keeps a couple of context lines visible above a match
// jumped to with n/N.
const matchScrollPadding = 2

func (s *Scroller) StartSearchPrompt() {
	s.search.prompting = true
	s.search.input = ""
}

func (s *Scroller) SearchPrompting() bool { return s.search.prompting }

// SearchActive reports whether a committed query is in effect.
func (s *Scroller) SearchActive() bool { return s.search.query != "" }

func (s *Scroller) SearchQuery() string { return s.search.query }

// HandleSearchPromptKey feeds one key press to the open search prompt.
// Printable characters append, enter commits the typed query, esc and
// backspacing past empty cancel. Committing an empty query also cancels,
// leaving any previous search untouched.
func (s *Scroller) HandleSearchPromptKey(msg tea.KeyPressMsg) PromptResult {
	switch msg.Code {
	case tea.KeyEscape:
		s.search.prompting = false
		s.search.input = ""

		return PromptCanceled

	case tea.KeyEnter:
		s.search.prompting = false

		if s.search.input == "" {
			return PromptCanceled
		}

		s.search.query = s.search.input
		s.search.input = ""

		return PromptCommitted

	case tea.KeyBackspace:
		if s.search.input == "" {
			s.search.prompting = false

			return PromptCanceled
		}

		_, size := utf8.DecodeLastRuneInString(s.search.input)
		s.search.input = s.search.input[:len(s.search.input)-size]

		return PromptPending
	}

	if msg.Text != "" && msg.Mod&^tea.ModShift == 0 {
		s.search.input += msg.Text
	}

	return PromptPending
}

// ClearSearch drops the query, matches, and highlights.
func (s *Scroller) ClearSearch() {
	s.search = searchState{}
	s.pushViewport()
}

// SetSearchMatches installs the match list for the committed query and
// repaints. The caller computes the matches — how content maps to matchable
// text differs per view.
func (s *Scroller) SetSearchMatches(matches []Match) {
	s.search.matches = matches
	s.search.current = min(max(0, s.search.current), max(0, len(matches)-1))
	s.pushViewport()
}

func (s *Scroller) SearchMatches() []Match { return s.search.matches }

// JumpToMatch scrolls match i (wrapped into range) to sit a couple of lines
// below the viewport top and makes it the current match.
func (s *Scroller) JumpToMatch(i int) {
	n := len(s.search.matches)
	if n == 0 {
		return
	}

	i = ((i % n) + n) % n
	s.search.current = i
	s.Viewport.SetYOffset(max(0, s.search.matches[i].Line-matchScrollPadding))
}

func (s *Scroller) NextMatch() { s.JumpToMatch(s.search.current + 1) }

func (s *Scroller) PrevMatch() { s.JumpToMatch(s.search.current - 1) }

// JumpToFirstMatchFrom jumps to the first match at or below the given line,
// wrapping to the first match overall when none follows.
func (s *Scroller) JumpToFirstMatchFrom(line int) {
	for i, m := range s.search.matches {
		if m.Line >= line {
			s.JumpToMatch(i)

			return
		}
	}

	s.JumpToMatch(0)
}

// SearchFooterLabel is the footer text for the search state: the live
// prompt with a block cursor while typing, the committed query otherwise,
// empty when no search is in play.
func (s *Scroller) SearchFooterLabel() string {
	if s.search.prompting {
		return style.Faint("/") + s.search.input + ansi.Reverse + " " + ansi.ReverseOff
	}

	if s.search.query != "" {
		return style.Faint("/") + s.search.query
	}

	return ""
}

// SearchCountLabel is the "3/17" match counter, or "no matches".
func (s *Scroller) SearchCountLabel() string {
	if !s.SearchActive() || s.search.prompting {
		return ""
	}

	if len(s.search.matches) == 0 {
		return style.Faint("no matches")
	}

	return style.Faint(fmt.Sprintf("%d/%d", s.search.current+1, len(s.search.matches)))
}

// FindMatches locates query in the given plain lines as cell spans.
// Smart case: an all-lowercase query matches case-insensitively, any
// uppercase character makes the match exact.
func FindMatches(plainLines []string, query string) []Match {
	if query == "" {
		return nil
	}

	caseless := query == strings.ToLower(query)

	var matches []Match

	for lineIdx, line := range plainLines {
		from := 0

		for {
			start, end := indexMatch(line, query, from, caseless)
			if start < 0 {
				break
			}

			matches = append(matches, Match{
				Line:      lineIdx,
				StartCell: xansi.StringWidth(line[:start]),
				EndCell:   xansi.StringWidth(line[:end]),
			})

			from = end
		}
	}

	return matches
}

// indexMatch finds the next occurrence of query in line at or after from,
// returning its start and end byte offsets, or -1, -1. Caseless comparison
// folds rune-wise with unicode.ToLower rather than lowering the whole line,
// so the offsets stay valid in line even when folding changes byte lengths.
func indexMatch(line, query string, from int, caseless bool) (int, int) {
	if !caseless {
		idx := strings.Index(line[from:], query)
		if idx < 0 {
			return -1, -1
		}

		return from + idx, from + idx + len(query)
	}

	for i := from; i < len(line); {
		if end, ok := foldPrefixEnd(line[i:], query); ok {
			return i, i + end
		}

		_, size := utf8.DecodeRuneInString(line[i:])
		i += size
	}

	return -1, -1
}

// foldPrefixEnd reports whether s begins with query under rune-wise ToLower
// folding, returning the byte length of the matched prefix.
func foldPrefixEnd(s, query string) (int, bool) {
	i := 0

	for _, qr := range query {
		if i >= len(s) {
			return 0, false
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		if unicode.ToLower(r) != qr {
			return 0, false
		}

		i += size
	}

	return i, true
}

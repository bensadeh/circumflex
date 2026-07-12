package comments

import (
	"slices"

	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// commentMatch locates one search hit in comment-relative coordinates:
// lineInComment counts from the comment's header line, so the position
// survives collapse-state rebuilds. flatIdx -1 addresses the thread header
// above the first comment, where lineInComment is already absolute.
type commentMatch struct {
	flatIdx       int
	lineInComment int
	startCell     int
	endCell       int
}

// handleSearchKeys routes keys to the search prompt and, while a search is
// active, to match navigation. The bool reports whether the key was consumed.
func (m *Model) handleSearchKeys(msg tea.KeyPressMsg) bool {
	if m.SearchPrompting() {
		switch m.HandleSearchPromptKey(msg) {
		case pane.PromptCommitted:
			m.commitSearch()

		case pane.PromptPending, pane.PromptCanceled:
			// Hits track the prompt live; a cancel recomputes the prior
			// query's matches (or none), where the untouched searchCurrent
			// is valid again — the prompt consumed every key, so the list
			// is the same one it indexed before.
			m.searchMatches = m.findAllMatches(m.ActiveQuery())
			m.syncDecorations()
		}

		return true
	}

	switch {
	case key.Matches(msg, m.keymap.Search):
		// Search covers every comment, so the prompt opens over the fully
		// expanded tree — what's on screen is what's being searched.
		m.expandAll()
		m.StartSearchPrompt()

	case m.SearchActive() && key.Matches(msg, m.keymap.ClearSearch):
		m.clearSearch()

	case m.SearchActive() && key.Matches(msg, m.keymap.NextMatch):
		m.jumpToSearchMatch(m.searchCurrent + 1)

	case m.SearchActive() && key.Matches(msg, m.keymap.PrevMatch):
		m.jumpToSearchMatch(m.searchCurrent - 1)

	default:
		return false
	}

	return true
}

// commitSearch resolves the committed query against the fully expanded tree:
// opening the prompt expanded everything, and the prompt consumes every key,
// so no branch can have re-collapsed since.
func (m *Model) commitSearch() {
	m.searchMatches = m.findAllMatches(m.SearchQuery())
	m.searchCurrent = 0
	m.syncDecorations()
	m.jumpToFirstSearchMatchFrom(m.Viewport.YOffset())
}

func (m *Model) clearSearch() {
	m.searchMatches = nil
	m.searchCurrent = 0
	m.ClearSearch()
}

// findAllMatches searches the thread header and every comment — those inside
// collapsed branches included; opening the prompt expands the tree to show
// them.
func (m *Model) findAllMatches(query string) []commentMatch {
	var out []commentMatch

	for _, hit := range pane.FindMatches(plainLines(splitLines(m.rc.header)), query) {
		out = append(out, commentMatch{flatIdx: -1, lineInComment: hit.Line, startCell: hit.StartCell, endCell: hit.EndCell})
	}

	for i := range m.prerendered {
		pre := &m.prerendered[i]

		// Stripping is memoized per prerender: live search re-runs this on
		// every prompt key, and only the substring scan should be paid then.
		if pre.plain == nil {
			lines := make([]string, 0, len(pre.header)+len(pre.content))
			lines = append(lines, pre.header...)
			lines = append(lines, pre.content...)
			pre.plain = plainLines(lines)
		}

		for _, hit := range pane.FindMatches(pre.plain, query) {
			out = append(out, commentMatch{flatIdx: i, lineInComment: hit.Line, startCell: hit.StartCell, endCell: hit.EndCell})
		}
	}

	return out
}

func plainLines(lines []string) []string {
	plain := make([]string, len(lines))
	for i, line := range lines {
		plain[i] = xansi.Strip(line)
	}

	return plain
}

// absoluteMatches resolves the matches currently on screen to viewport lines
// for highlighting, along with the current match's index in that list — -1
// while the current match sits inside a collapsed branch. Hidden matches
// resolve to nothing — they highlight once revealed.
func (m *Model) absoluteMatches() ([]pane.Match, int) {
	var out []pane.Match

	current := -1

	for i, cm := range m.searchMatches {
		if line, visible := m.matchLine(cm); visible {
			if i == m.searchCurrent {
				current = len(out)
			}

			out = append(out, pane.Match{Line: line, StartCell: cm.startCell, EndCell: cm.endCell})
		}
	}

	return out, current
}

// matchLine resolves a match to its current viewport line; false means the
// match sits inside a collapsed branch.
func (m *Model) matchLine(cm commentMatch) (int, bool) {
	if cm.flatIdx < 0 {
		return cm.lineInComment, true
	}

	lm := m.lineMetrics[cm.flatIdx]
	if lm.LineCount == 0 {
		return 0, false
	}

	return lm.StartLine + cm.lineInComment, true
}

// jumpToSearchMatch makes match idx (wrapped into range) current and scrolls
// it a couple of lines below the viewport top. Opening the prompt expanded
// the whole tree, so the reveal only fires when the user re-collapsed a
// branch while the search was active.
func (m *Model) jumpToSearchMatch(idx int) {
	n := len(m.searchMatches)
	if n == 0 {
		return
	}

	idx = ((idx % n) + n) % n
	m.searchCurrent = idx
	cm := m.searchMatches[idx]

	if cm.flatIdx >= 0 && m.lineMetrics[cm.flatIdx].LineCount == 0 {
		m.revealComment(cm.flatIdx)
	}

	m.focusSearchMatch(cm)
	m.syncDecorations()

	line, _ := m.matchLine(cm)
	m.Viewport.SetYOffset(max(0, line-scrollPadding))
}

// jumpToFirstSearchMatchFrom lands on the first visible match at or below
// the given line, falling back to the first match overall — which may be
// hidden and get revealed.
func (m *Model) jumpToFirstSearchMatchFrom(yOffset int) {
	for i, cm := range m.searchMatches {
		if line, visible := m.matchLine(cm); visible && line >= yOffset {
			m.jumpToSearchMatch(i)

			return
		}
	}

	m.jumpToSearchMatch(0)
}

// focusSearchMatch moves navigate-mode focus to the matched comment so match
// navigation and comment navigation agree on where the user is.
func (m *Model) focusSearchMatch(cm commentMatch) {
	if m.mode != modeNavigate || cm.flatIdx < 0 {
		return
	}

	if vi := slices.Index(m.visible, cm.flatIdx); vi >= 0 && vi != m.focusedIdx {
		m.focusedIdx = vi
		m.syncDecorations()
	}
}

// revealComment uncollapses every ancestor of flatIdx so the comment becomes
// visible, leaving the comment's own collapse state alone. The backward walk
// relies on the pre-order flatten invariant: scanning up, the first entry at
// a shallower depth is an ancestor.
func (m *Model) revealComment(flatIdx int) {
	depth := m.flat[flatIdx].Depth

	for i := flatIdx - 1; i >= 0 && depth > 0; i-- {
		if m.flat[i].Depth < depth {
			m.flat[i].Collapsed = false
			depth = m.flat[i].Depth
		}
	}

	m.rebuildContent()
	m.syncExpandedDepth()
}

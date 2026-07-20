// Package pane holds what the full-screen detail views — the comment section
// and reader mode — share: the content-aware viewport and search engine,
// frame chrome and footers, browser commands, and the standalone runner.
package pane

import (
	"strings"

	"github.com/bensadeh/circumflex/style"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// Scroller wraps a viewport with content-aware scrolling: paging and
// clamping are computed against the real content length, ignoring the
// bottom padding SetLines appends so any line can scroll to the top.
type Scroller struct {
	Viewport     viewport.Model
	ContentLines int // excludes bottom padding

	// SearchCommittedIcon replaces the shared done-searching glyph in the
	// footer's search label; empty keeps the default.
	SearchCommittedIcon string

	lines        []string
	plainLines   []string
	search       searchState
	rowOverrides []RowOverride
	linkSpans    []Match
	linkInert    bool
	linkFetching bool
}

// SetLinkSpans installs the spans of the link the reader's URL selector sits
// on; nil clears them. They paint over any search highlights — the selection
// is the thing being acted on. inert swaps the selection colors for the red
// bar, marking a link the view will not open.
func (s *Scroller) SetLinkSpans(matches []Match, inert bool) {
	s.linkSpans = matches
	s.linkInert = inert
}

// SetLinkFetching repaints the selection in the muted in-flight colors while
// the followed link's fetch runs. The shell that owns the fetch toggles it:
// on at fetch start, off when the fetch ends without a page (failure,
// cancel) — success replaces the view, taking the state with it.
func (s *Scroller) SetLinkFetching(active bool) {
	s.linkFetching = active
}

func (s *Scroller) LinkSpans() []Match { return s.linkSpans }

func (s *Scroller) LinkSpansInert() bool { return s.linkInert }

func (s *Scroller) LinkFetching() bool { return s.linkFetching }

// RowOverride substitutes the rendered row at content line Line for display
// only, leaving the stored content untouched. The comment section uses it to
// swap in the focused header variant without rebuilding the document.
type RowOverride struct {
	Line    int
	Content string
}

// SetRowOverrides installs the display-time row substitutions. Overrides
// must not change a row's visible width.
func (s *Scroller) SetRowOverrides(overrides []RowOverride) {
	s.rowOverrides = overrides
}

// SetLines replaces the viewport content. A viewport-height of padding is
// appended so jump targets near the end can scroll to the top of the view;
// ClampScroll keeps ordinary scrolling within the real content.
func (s *Scroller) SetLines(lines []string) {
	s.lines = lines
	s.plainLines = nil
	s.ContentLines = len(lines)
	s.pushViewport()
}

// pushViewport hands the content to the viewport with the jump padding
// appended. Only slice headers are copied — the line strings themselves are
// shared. Decorations (focus, search highlights) are NOT baked in here;
// DecorateView applies them per frame to the visible rows only, keeping
// focus moves and match updates off this O(document) path — the viewport
// rescans every line's width on each content push.
func (s *Scroller) pushViewport() {
	padded := make([]string, len(s.lines)+s.Viewport.Height())
	copy(padded, s.lines)
	s.Viewport.SetContentLines(padded)
}

// DecorateView applies the display-time decorations to the viewport's
// rendered window: row overrides first, then search-match overlays on top,
// so a match inside a focused header highlights the focused variant. Row k
// of the view shows content line YOffset+k — the invariant that makes this
// mapping valid is that pane content never scrolls horizontally.
func (s *Scroller) DecorateView(view string) string {
	if view == "" || (len(s.rowOverrides) == 0 && len(s.search.matches) == 0 && len(s.linkSpans) == 0) {
		return view
	}

	rows := strings.Split(view, "\n")
	top := s.Viewport.YOffset()

	visibleRow := func(line int) (int, bool) {
		row := line - top

		return row, row >= 0 && row < len(rows) && line < s.ContentLines
	}

	for _, o := range s.rowOverrides {
		if row, ok := visibleRow(o.Line); ok {
			rows[row] = o.Content
		}
	}

	// While the prompt is open the hits are live-updating and none of them
	// is the current one yet — n/N navigation starts on commit.
	current := s.search.current
	if s.search.prompt.Active() {
		current = -1
	}

	// Spans are grouped per row and painted in one pass: match lists come in
	// line order, so each row's spans arrive sorted, and a broad query can
	// put dozens of hits on a single row.
	var spansByRow map[int][]style.SearchSpan

	for i, m := range s.search.matches {
		if row, ok := visibleRow(m.Line); ok {
			if spansByRow == nil {
				spansByRow = make(map[int][]style.SearchSpan)
			}

			spansByRow[row] = append(spansByRow[row],
				style.SearchSpan{StartCell: m.StartCell, EndCell: m.EndCell, Current: i == current})
		}
	}

	for row, spans := range spansByRow {
		rows[row] = style.OverlaySearchSpans(rows[row], spans)
	}

	// Painted after the search spans so a selection overlapping a match ends
	// up in the selection's colors — its SGRs re-assert through the earlier
	// paint's replayed escapes.
	overlayLink := style.OverlayLinkSpans

	switch {
	case s.linkFetching:
		overlayLink = style.OverlayFetchingLinkSpans
	case s.linkInert:
		overlayLink = style.OverlayInertLinkSpans
	}

	for _, m := range s.linkSpans {
		if row, ok := visibleRow(m.Line); ok {
			rows[row] = overlayLink(rows[row],
				[]style.SearchSpan{{StartCell: m.StartCell, EndCell: m.EndCell}})
		}
	}

	return strings.Join(rows, "\n")
}

// Lines is the stored content as styled lines, for views that scan the
// rendered output itself (the reader's link extraction).
func (s *Scroller) Lines() []string {
	return s.lines
}

// RefreshPadding re-pushes the stored content so the bottom padding matches
// the current viewport height, for height-only resizes where the content
// itself is unchanged.
func (s *Scroller) RefreshPadding() {
	s.pushViewport()
}

// PlainLines is the content with ANSI styling stripped — the text as the
// user sees it, for matching against. Stripped on first use after a
// content change, so views that never need it don't pay for it.
// xansi.Strip (a parser walk) is used over ansi.Strip (a backtracking
// regexp): measured ~6× faster on escape-heavy content like image art.
func (s *Scroller) PlainLines() []string {
	if s.plainLines == nil && s.lines != nil {
		s.plainLines = make([]string, len(s.lines))

		for i, line := range s.lines {
			s.plainLines[i] = xansi.Strip(line)
		}
	}

	return s.plainLines
}

// NewViewport returns a viewport with the bindings the detail views handle
// themselves (paging) disabled, along with mouse wheel handling. Horizontal
// scrolling is disabled entirely: pane content is pre-wrapped to fit, and
// DecorateView's row mapping relies on a zero x-offset.
func NewViewport(width, height int) viewport.Model {
	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(max(0, height)),
	)

	vp.KeyMap = viewport.DefaultKeyMap()
	vp.KeyMap.HalfPageDown.SetEnabled(false)
	vp.KeyMap.HalfPageUp.SetEnabled(false)
	vp.KeyMap.PageDown.SetEnabled(false)
	vp.KeyMap.PageUp.SetEnabled(false)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)
	vp.SetHorizontalStep(0)
	vp.MouseWheelEnabled = false

	return vp
}

func (s *Scroller) maxOffset() int {
	return max(0, s.ContentLines-s.Viewport.Height())
}

// ClampScroll prevents scrolling down past the last content line while still
// allowing upward scrolling from a position beyond the clamp point (e.g.
// after an n/N jump).
func (s *Scroller) ClampScroll(before int) {
	after := s.Viewport.YOffset()

	if after > before && after > s.maxOffset() {
		s.Viewport.SetYOffset(max(before, s.maxOffset()))
	}
}

// Forward passes msg to the viewport and clamps the resulting scroll.
func (s *Scroller) Forward(msg tea.Msg) tea.Cmd {
	before := s.Viewport.YOffset()

	var cmd tea.Cmd

	s.Viewport, cmd = s.Viewport.Update(msg)
	s.ClampScroll(before)

	return cmd
}

func (s *Scroller) HandleMouseWheel(msg tea.MouseWheelMsg) {
	delta := s.Viewport.MouseWheelDelta

	switch msg.Button {
	case tea.MouseWheelDown:
		s.scrollDownTo(s.Viewport.YOffset() + delta)
	case tea.MouseWheelUp:
		s.Viewport.SetYOffset(max(0, s.Viewport.YOffset()-delta))
	}
}

func (s *Scroller) HalfPageDown() {
	s.scrollDownTo(s.Viewport.YOffset() + s.Viewport.Height()/2)
}

func (s *Scroller) HalfPageUp() {
	s.Viewport.SetYOffset(max(0, s.Viewport.YOffset()-s.Viewport.Height()/2))
}

func (s *Scroller) PageDown() {
	s.scrollDownTo(s.Viewport.YOffset() + s.Viewport.Height())
}

func (s *Scroller) PageUp() {
	s.Viewport.SetYOffset(max(0, s.Viewport.YOffset()-s.Viewport.Height()))
}

// scrollDownTo clamps a downward move by ClampScroll's monotonic rule: it
// stays within the real content, except when the view already sits past
// maxOffset — an n/N jump parks it there so a match near the end can reach
// the top — where a downward key holds position instead of bouncing back up.
func (s *Scroller) scrollDownTo(target int) {
	s.Viewport.SetYOffset(min(target, max(s.Viewport.YOffset(), s.maxOffset())))
}

// GotoBottom scrolls the last line of real content to the bottom of the
// viewport, ignoring the bottom padding.
func (s *Scroller) GotoBottom() {
	s.Viewport.SetYOffset(s.maxOffset())
}

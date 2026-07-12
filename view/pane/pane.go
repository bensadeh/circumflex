// Package pane holds the viewport plumbing shared by the full-screen detail
// views: the comment section and reader mode.
package pane

import (
	"github.com/bensadeh/circumflex/ansi"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

// Scroller wraps a viewport with content-aware scrolling: paging and
// clamping are computed against the real content length, ignoring the
// bottom padding SetLines appends so any line can scroll to the top.
type Scroller struct {
	Viewport     viewport.Model
	ContentLines int // excludes bottom padding

	lines      []string
	plainLines []string
}

// SetLines replaces the viewport content. A viewport-height of padding is
// appended so jump targets near the end can scroll to the top of the view;
// ClampScroll keeps ordinary scrolling within the real content.
func (s *Scroller) SetLines(lines []string) {
	s.lines = lines
	s.plainLines = nil
	s.ContentLines = len(lines)
	s.Viewport.SetContentLines(append(lines, make([]string, s.Viewport.Height())...))
}

// PlainLines is the content with ANSI styling stripped — the text as the
// user sees it, for matching against. Stripped on first use after a
// content change, so views that never need it don't pay for it.
func (s *Scroller) PlainLines() []string {
	if s.plainLines == nil && s.lines != nil {
		s.plainLines = make([]string, len(s.lines))

		for i, line := range s.lines {
			s.plainLines[i] = ansi.Strip(line)
		}
	}

	return s.plainLines
}

// NewViewport returns a viewport with the bindings the detail views handle
// themselves (paging) disabled, along with mouse wheel handling.
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
		s.Viewport.SetYOffset(min(s.Viewport.YOffset()+delta, s.maxOffset()))
	case tea.MouseWheelUp:
		s.Viewport.SetYOffset(max(0, s.Viewport.YOffset()-delta))
	}
}

func (s *Scroller) HalfPageDown() {
	s.Viewport.SetYOffset(min(s.Viewport.YOffset()+s.Viewport.Height()/2, s.maxOffset()))
}

func (s *Scroller) HalfPageUp() {
	s.Viewport.SetYOffset(max(0, s.Viewport.YOffset()-s.Viewport.Height()/2))
}

func (s *Scroller) PageDown() {
	s.Viewport.SetYOffset(min(s.Viewport.YOffset()+s.Viewport.Height(), s.maxOffset()))
}

func (s *Scroller) PageUp() {
	s.Viewport.SetYOffset(max(0, s.Viewport.YOffset()-s.Viewport.Height()))
}

// GotoBottom scrolls the last line of real content to the bottom of the
// viewport, ignoring the bottom padding.
func (s *Scroller) GotoBottom() {
	s.Viewport.SetYOffset(s.maxOffset())
}

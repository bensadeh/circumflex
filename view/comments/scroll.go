package comments

import "slices"

// findCommentAtScroll returns the visible index of the comment whose header
// line is topmost within the current viewport.
func (m *Model) findCommentAtScroll() int {
	yOffset := m.Viewport.YOffset()
	bottom := yOffset + m.Viewport.VisibleLineCount()

	for vi, flatIdx := range m.visible {
		if m.lineMetrics[flatIdx].StartLine >= yOffset && m.lineMetrics[flatIdx].StartLine < bottom {
			return vi
		}
	}

	return 0
}

// snapFocusToVisible adjusts focusedIdx after a half-page scroll if the
// focused comment's header is no longer on screen. direction > 0 picks the
// topmost visible comment; direction < 0 picks the bottommost.
func (m *Model) snapFocusToVisible(direction int) {
	if len(m.visible) == 0 || m.focusedIdx < 0 {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	top := m.Viewport.YOffset()
	bottom := top + m.Viewport.VisibleLineCount()

	if m.lineMetrics[flatIdx].StartLine >= top && m.lineMetrics[flatIdx].StartLine < bottom {
		return
	}

	if direction > 0 {
		for vi, fi := range m.visible {
			if m.lineMetrics[fi].StartLine >= top && m.lineMetrics[fi].StartLine < bottom {
				m.focusedIdx = vi
				m.syncDecorations()

				return
			}
		}
	} else {
		for vi, fi := range slices.Backward(m.visible) {
			if m.lineMetrics[fi].StartLine >= top && m.lineMetrics[fi].StartLine < bottom {
				m.focusedIdx = vi
				m.syncDecorations()

				return
			}
		}
	}
}

func (m *Model) navigateComment(direction int) {
	if len(m.visible) == 0 {
		return
	}

	newIdx := m.focusedIdx + direction
	if newIdx < 0 || newIdx >= len(m.visible) {
		return
	}

	m.focusedIdx = newIdx
	m.syncDecorations()
	m.scrollToFocused()
}

func (m *Model) jumpToTopLevel(direction int) {
	yOffset := m.Viewport.YOffset()

	if direction > 0 {
		for vi, flatIdx := range m.visible {
			if m.flat[flatIdx].Depth == 0 && m.lineMetrics[flatIdx].StartLine > yOffset {
				m.Viewport.SetYOffset(m.lineMetrics[flatIdx].StartLine)
				m.setFocusIfNavigating(vi)

				return
			}
		}
	} else {
		for i, flatIdx := range slices.Backward(m.visible) {
			if m.flat[flatIdx].Depth == 0 && m.lineMetrics[flatIdx].StartLine < yOffset {
				m.Viewport.SetYOffset(m.lineMetrics[flatIdx].StartLine)
				m.setFocusIfNavigating(i)

				return
			}
		}

		if yOffset > 0 {
			m.Viewport.SetYOffset(0)
		}
	}
}

func (m *Model) setFocusIfNavigating(visibleIdx int) {
	if m.mode != modeNavigate {
		return
	}

	m.focusedIdx = visibleIdx
	m.syncDecorations()
}

// anchorComment returns the flat index of the comment nearest to the top of
// the viewport, used to keep the view stable across content rebuilds.
// Uses SepStart so a comment whose separator is visible at the viewport
// top is chosen as the anchor rather than the previous comment.
func (m *Model) anchorComment() int {
	yOffset := m.Viewport.YOffset()

	best := -1

	for _, flatIdx := range m.visible {
		if m.lineMetrics[flatIdx].SepStart > yOffset {
			break
		}

		best = flatIdx
	}

	return best
}

func (m *Model) screenPosition(flatIdx int) int {
	if flatIdx < 0 {
		return 0
	}

	return m.lineMetrics[flatIdx].StartLine - m.Viewport.YOffset()
}

func (m *Model) restoreScreenPosition(flatIdx, screenPos int) {
	if flatIdx < 0 {
		return
	}

	m.Viewport.SetYOffset(max(0, m.lineMetrics[flatIdx].StartLine-screenPos))
}

func (m *Model) gotoTop() {
	if m.mode == modeNavigate && len(m.visible) > 0 {
		m.focusedIdx = 0
		m.syncDecorations()
	}

	m.Viewport.GotoTop()
}

func (m *Model) gotoBottom() {
	if m.mode == modeNavigate && len(m.visible) > 0 {
		m.focusedIdx = len(m.visible) - 1
		m.syncDecorations()
	}

	m.GotoBottom()
}

func (m *Model) scrollToFocused() {
	if len(m.visible) == 0 || m.focusedIdx < 0 || m.focusedIdx >= len(m.visible) {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	lm := m.lineMetrics[flatIdx]

	top := m.Viewport.YOffset()
	bottom := top + m.Viewport.VisibleLineCount()

	if lm.StartLine < top {
		// Scrolling up — put comment a few lines below the top.
		m.Viewport.SetYOffset(max(0, lm.StartLine-scrollPadding))
	} else if lm.StartLine+lm.LineCount > bottom {
		if lm.LineCount >= m.Viewport.VisibleLineCount() {
			// Comment is taller than viewport — show its start.
			m.Viewport.SetYOffset(max(0, lm.StartLine-scrollPadding))
		} else {
			// Comment fits — scroll just enough to show it fully.
			m.Viewport.SetYOffset(lm.StartLine - m.Viewport.VisibleLineCount() + lm.LineCount + scrollPadding)
		}
	}
}

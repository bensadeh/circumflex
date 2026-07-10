package comments

func (m *Model) setCollapsed(collapsed bool) {
	if len(m.visible) == 0 || m.focusedIdx < 0 {
		return
	}

	flatIdx := m.visible[m.focusedIdx]
	fc := &m.flat[flatIdx]

	if fc.DescendantCount == 0 || fc.Collapsed == collapsed {
		return
	}

	screenPos := m.screenPosition(flatIdx)

	fc.Collapsed = collapsed

	m.rebuildContent()
	m.syncExpandedDepth()

	if m.focusedIdx >= len(m.visible) {
		m.focusedIdx = len(m.visible) - 1
		m.updateViewport()
	}

	m.restoreScreenPosition(flatIdx, screenPos)
}

func (m *Model) focusedComment() *flatComment {
	if m.focusedIdx < 0 || m.focusedIdx >= len(m.visible) {
		return nil
	}

	return &m.flat[m.visible[m.focusedIdx]]
}

func (m *Model) toggleCollapse() {
	if fc := m.focusedComment(); fc != nil && fc.DescendantCount > 0 {
		m.setCollapsed(!fc.Collapsed)
	}
}

func (m *Model) toggleCollapseAll() {
	allCollapsed := true

	for i := range m.flat {
		if m.flat[i].DescendantCount > 0 && !m.flat[i].Collapsed {
			allCollapsed = false

			break
		}
	}

	if allCollapsed {
		m.expandAll()
	} else {
		m.collapseAll()
	}
}

func (m *Model) collapseAll() {
	m.expandedDepth = 0
	m.setCollapseToDepth()
}

func (m *Model) expandAll() {
	m.expandedDepth = m.maxDepth
	m.setCollapseToDepth()
}

// setCollapseToDepth sets collapse state based on expandedDepth:
// comments at depth < expandedDepth are uncollapsed; the rest are collapsed.
func (m *Model) setCollapseToDepth() {
	anchorIdx := m.anchorComment()
	screenPos := m.screenPosition(anchorIdx)

	for i := range m.flat {
		if m.flat[i].DescendantCount == 0 {
			continue
		}

		m.flat[i].Collapsed = m.flat[i].Depth >= m.expandedDepth
	}

	m.rebuildContent()

	// If the anchor is still visible after the collapse, restore its exact
	// screen position so the viewport stays stable.
	if anchorIdx >= 0 && m.lineMetrics[anchorIdx].LineCount > 0 {
		m.restoreScreenPosition(anchorIdx, screenPos)

		return
	}

	// The anchor was collapsed away. Find the nearest visible ancestor,
	// then position the viewport at the next visible comment at the same
	// depth or shallower (the next sibling or uncle). This works at any
	// nesting level so collapsing never jumps out further than necessary.
	ancestorIdx := -1

	for i := anchorIdx - 1; i >= 0; i-- {
		if m.lineMetrics[i].LineCount > 0 {
			ancestorIdx = i

			break
		}
	}

	if ancestorIdx < 0 {
		return
	}

	ancestorDepth := m.flat[ancestorIdx].Depth

	for _, flatIdx := range m.visible {
		if flatIdx > ancestorIdx && m.flat[flatIdx].Depth <= ancestorDepth {
			m.Viewport.SetYOffset(m.lineMetrics[flatIdx].SepStart)

			return
		}
	}

	// No next sibling — position at the end of the ancestor.
	lm := m.lineMetrics[ancestorIdx]
	m.Viewport.SetYOffset(lm.StartLine + lm.LineCount)
}

// syncExpandedDepth derives expandedDepth from the actual collapse state,
// so the depth indicator matches what's on screen after navigate mode
// may have individually collapsed/expanded comments.
func (m *Model) syncExpandedDepth() {
	maxUncollapsed := -1

	for i := range m.flat {
		if m.flat[i].DescendantCount > 0 && !m.flat[i].Collapsed && m.flat[i].Depth > maxUncollapsed {
			maxUncollapsed = m.flat[i].Depth
		}
	}

	m.expandedDepth = maxUncollapsed + 1
}

func (m *Model) expandLevel() {
	if m.expandedDepth >= m.maxDepth {
		return
	}

	m.expandedDepth++
	m.setCollapseToDepth()
}

func (m *Model) collapseLevel() {
	if m.expandedDepth <= 0 {
		return
	}

	m.expandedDepth--
	m.setCollapseToDepth()
}

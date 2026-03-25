package list

import (
	"clx/bubble/ranking"
	"clx/header"
	"clx/version"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// View renders the component.
func (m *Model) View() string {
	if m.state == StateHelpScreen {
		return fmt.Sprintf("%s\n%s\n%s",
			header.Header(
				m.cat.ActiveCategories(),
				m.cat.CurrentIndex(),
				m.width),
			m.viewport.View(),
			m.statusAndPaginationView())
	}

	var (
		sections    []string
		availHeight = m.height
	)

	if m.state == StateEditorOpen {
		return ""
	}

	if m.showTitle {
		v := m.titleView()
		sections = append(sections, v)
		availHeight -= lipgloss.Height(v)
	}

	if m.showStatusBar {
		v := m.statusAndPaginationView()
		availHeight -= lipgloss.Height(v)
	}

	content := m.contentStyle.Height(availHeight).Render(m.populatedView())
	allItems := m.VisibleItems()
	totalItems := len(allItems)

	start, end := m.pager.Paginator.GetSliceBounds(totalItems)
	readStatuses := make([]bool, end-start)

	for i, it := range allItems[start:end] {
		readStatuses[i] = m.history.Contains(it.ID)
	}

	rankings := ranking.Rankings(
		false,
		m.pager.Paginator.PerPage,
		totalItems,
		m.pager.cursor,
		m.pager.Paginator.Page,
		m.pager.Paginator.TotalPages,
		readStatuses,
		m.pager.transition != nil)

	rankingsAndContent := lipgloss.JoinHorizontal(lipgloss.Top, rankings, content)
	sections = append(sections, rankingsAndContent)

	if m.showStatusBar {
		v := m.statusAndPaginationView()
		sections = append(sections, v)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) titleView() string {
	return header.Header(
		m.cat.ActiveCategories(),
		m.cat.CurrentIndex(),
		m.width)
}

func (m *Model) statusAndPaginationView() string {
	var (
		centerContent string
		rightContent  string
	)

	underscore := m.underlineStyle.Render(" ")
	underline := strings.Repeat(underscore, m.width)

	switch {
	case m.state == StateHelpScreen:
		centerContent = m.faintStyle.Render(
			"github.com/bensadeh/circumflex • version " + version.Version)
	case m.status.showSpinner:
		centerContent = m.status.spinnerView()
	default:
		centerContent = m.status.message
	}

	switch m.state {
	case StateHelpScreen:
		// no pagination on help screen
	case StateFetching:
		rightContent = strings.Repeat(m.Styles.InactivePaginationDot.String(), m.config.PageMultiplier)
	case StateStartup, StateBrowsing, StateAddFavoritesPrompt, StateRemoveFavoritesPrompt, StateEditorOpen:
		rightContent = m.pager.Paginator.View()
	}

	left := m.statusLeftStyle.Render("")

	center := m.statusMidStyle.
		Width(m.width - statusBarEdgeWidth - statusBarEdgeWidth).
		Render(centerContent)

	right := m.statusEndStyle.Render(rightContent)

	return underline + "\n" +
		m.Styles.StatusBar.Render(left) +
		m.Styles.StatusBar.Render(center) +
		m.Styles.StatusBar.Render(right)
}

func (m *Model) populatedView() string {
	allItems := m.VisibleItems()

	var b strings.Builder

	// Empty states
	if len(allItems) == 0 {
		return m.Styles.NoItems.Render("")
	}

	if len(allItems) > 0 {
		start, end := m.pager.Paginator.GetSliceBounds(len(allItems))
		itemsToShow := allItems[start:end]

		for i, item := range itemsToShow {
			m.delegate.Render(&b, m, i+start, item)

			if i != len(itemsToShow)-1 {
				fmt.Fprint(&b, strings.Repeat("\n", m.delegate.Spacing()+1))
			}
		}
	}

	// If there aren't enough items to fill up this page (always the last page)
	// then we need to add some newlines to fill up the space where items would
	// have been.
	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(allItems))
	if itemsOnPage < m.pager.Paginator.PerPage {
		n := (m.pager.Paginator.PerPage - itemsOnPage) * (m.delegate.Height() + m.delegate.Spacing())
		if len(allItems) == 0 {
			n -= m.delegate.Height() - 1
		}

		_, _ = fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

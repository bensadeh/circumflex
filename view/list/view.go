package list

import (
	"fmt"
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/view/list/ranking"

	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	if m.state == StateHelpScreen {
		underscore := m.underlineStyle.Render(" ")
		underline := strings.Repeat(underscore, m.width)

		return fmt.Sprintf("%s\n%s\n%s\n%s",
			header.HelpHeader("Keyboard Shortcuts", m.width),
			m.viewport.View(),
			underline,
			help.Footer(m.width))
	}

	var (
		sections    []string
		availHeight = m.height
	)

	if m.state == StateReaderView {
		return m.readerView.View()
	}

	if m.state == StateCommentView {
		return m.commentView.View()
	}

	v := m.titleView()
	sections = append(sections, v)
	availHeight -= lipgloss.Height(v)

	statusView := m.statusAndPaginationView()
	availHeight -= lipgloss.Height(statusView)

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
	sections = append(sections, statusView)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) titleView() string {
	var sv string
	if m.status.showSpinner {
		sv = m.status.spinnerView()
	}

	return header.Header(
		m.cat.ActiveCategories(),
		m.cat.CurrentIndex(),
		m.width,
		sv)
}

func (m *Model) statusAndPaginationView() string {
	var (
		centerContent string
		rightContent  string
	)

	underscore := m.underlineStyle.Render(" ")
	underline := strings.Repeat(underscore, m.width)

	centerContent = m.status.message

	switch m.state {
	case StateFetching:
		rightContent = strings.Repeat(m.styles.InactivePaginationDot.String(), m.config.PageMultiplier)
	case StateStartup, StateBrowsing, StateAddFavoritesPrompt, StateRemoveFavoritesPrompt, StateReaderView:
		rightContent = m.pager.Paginator.View()
	case StateCommentView, StateHelpScreen:
		// These views handle their own footer.
	}

	left := m.statusLeftStyle.Render("")

	center := m.statusMidStyle.
		Width(m.width - statusBarEdgeWidth - statusBarEdgeWidth).
		Render(centerContent)

	right := m.statusEndStyle.Render(rightContent)

	return underline + "\n" + left + center + right
}

func (m *Model) populatedView() string {
	allItems := m.VisibleItems()

	var b strings.Builder

	if len(allItems) == 0 {
		return m.styles.NoItems.Render("")
	}

	start, end := m.pager.Paginator.GetSliceBounds(len(allItems))
	itemsToShow := allItems[start:end]

	for i, item := range itemsToShow {
		m.delegate.Render(&b, m, i+start, item)

		if i != len(itemsToShow)-1 {
			fmt.Fprint(&b, strings.Repeat("\n", m.delegate.Spacing()+1))
		}
	}

	// If there aren't enough items to fill up this page (always the last page)
	// then we need to add some newlines to fill up the space where items would
	// have been.
	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(allItems))
	if itemsOnPage < m.pager.Paginator.PerPage {
		n := (m.pager.Paginator.PerPage - itemsOnPage) * (m.delegate.Height() + m.delegate.Spacing())

		_, _ = fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

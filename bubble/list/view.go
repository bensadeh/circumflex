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
			header.GetHeader(
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
	totalItems := len(m.VisibleItems())
	rankings := ranking.GetRankings(
		false,
		m.Paginator.PerPage,
		totalItems,
		m.cursor,
		m.Paginator.Page,
		m.Paginator.TotalPages)

	rankingsAndContent := lipgloss.JoinHorizontal(lipgloss.Top, rankings, content)
	sections = append(sections, rankingsAndContent)

	if m.showStatusBar {
		v := m.statusAndPaginationView()
		sections = append(sections, v)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) titleView() string {
	return header.GetHeader(
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
	case m.showSpinner:
		centerContent = m.spinnerView()
	default:
		centerContent = m.statusMessage
	}

	if m.state != StateHelpScreen {
		rightContent = m.Paginator.View()
	}

	left := m.statusLeftStyle.Render("")

	center := m.statusMidStyle.
		Width(m.width - 5 - 5).
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
		start, end := m.Paginator.GetSliceBounds(len(allItems))
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
	itemsOnPage := m.Paginator.ItemsOnPage(len(allItems))
	if itemsOnPage < m.Paginator.PerPage {
		n := (m.Paginator.PerPage - itemsOnPage) * (m.delegate.Height() + m.delegate.Spacing())
		if len(allItems) == 0 {
			n -= m.delegate.Height() - 1
		}

		_, _ = fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

func (m *Model) spinnerView() string {
	return m.spinner.View()
}

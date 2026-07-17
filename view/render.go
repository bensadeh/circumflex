package view

import (
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/view/pane"

	xansi "github.com/charmbracelet/x/ansi"

	"charm.land/lipgloss/v2"
)

func (m *model) View() string {
	// In the wide layout help renders in the detail pane instead of taking
	// over the screen, so the story list stays visible next to it.
	if m.screen == screenHelp && !m.isWide() {
		return m.helpView()
	}

	if m.isWide() {
		return m.wideView()
	}

	if m.detail != nil {
		return m.overlayDetailStatus(m.detail.View(), m.width)
	}

	return m.browsingView()
}

// overlayDetailStatus writes fetch and status feedback onto the last row of a
// detail view, which reserves that row as footer space. Narrow J/K story
// navigation and link fetches at either width stay on the open story while
// the next page loads, so their spinners and errors must surface here rather
// than on the front page. width is the pane the view fills: the whole
// terminal in the narrow layout, the detail pane in the wide one.
func (m *model) overlayDetailStatus(view string, width int) string {
	var status string

	switch {
	case m.fetch.inFlight():
		status = m.status.spinnerView()
	case m.status.text.Message() != "":
		status = m.status.text.Message()
	default:
		return view
	}

	return pane.OverlayStatus(view, status, width)
}

// browsingView is the front page: category header, story list, status bar.
// It fills the screen in the narrow layout and the left pane in the wide one.
func (m *model) browsingView() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.titleView(),
		m.list.View(m.listFrame()),
		m.statusAndPaginationView())
}

func (m *model) titleView() string {
	var sv string
	// In the wide layout the spinner always shows centered in the detail pane
	// instead, so loading feedback stays in one place for every kind of fetch.
	if m.status.showSpinner && !m.isWide() {
		sv = m.status.spinnerView()
	}

	// Search mode owns the tab row: the filters render where the
	// (then-inert) tabs would be, the date group right-aligned to the
	// front-page help panels' edge.
	if m.cat.Searching() {
		return header.SearchHeader(m.searchFilters.headerGroups(),
			help.MainMenuPanelRightEdge(m.listWidth()), m.listWidth(), sv, m.wideDetailOpen())
	}

	return header.Header(
		m.cat.ActiveCategories(),
		m.cat.CurrentIndex(),
		m.listWidth(),
		sv,
		m.wideDetailOpen())
}

// bottomBar renders the footer rule, shared with the detail views so the
// list's rule and theirs can never drift apart (memorial tint included).
func (m *model) bottomBar(width int) string {
	return pane.FooterSeparator(width)
}

func (m *model) statusAndPaginationView() string {
	var (
		centerContent string
		rightContent  string
	)

	underline := m.bottomBar(m.listWidth())

	centerContent = m.status.text.Message()

	midStyle := m.statusMidStyle
	leftWidth := statusBarEdgeWidth

	switch {
	// The open search prompt takes over the message slot, rendered from the
	// left like the detail views' search footer. The sigil zone hangs into
	// the left slot — the nerd magnifier is wider than the committed line's
	// plain /, and outdenting it keeps the typed text on the column it
	// stays on after committing.
	case m.prompt == promptSearch:
		centerContent = pane.PromptLabel(m.searchPrompt.Text(), m.config.EnableNerdFonts)
		midStyle = midStyle.Align(lipgloss.Left)
		leftWidth -= pane.PromptSigilWidth(m.config.EnableNerdFonts) - 1

	// Search mode keeps the committed query on the status row where
	// transient messages otherwise show; the filters live in the header.
	// The icon stays on the same outdented cell through the commit — the
	// results glyph replaces the magnifier, the query dims, nothing moves.
	// The committed zone is icon+gap like the prompt's, so the prompt
	// sigil width applies to both.
	case centerContent == "" && m.cat.Searching() && m.searchQuery != "":
		centerContent = pane.CommittedSearchLabel(m.searchQuery, nerdfonts.SearchResults, m.config.EnableNerdFonts)
		midStyle = midStyle.Align(lipgloss.Left)
		leftWidth -= pane.PromptSigilWidth(m.config.EnableNerdFonts) - 1
	}

	// The page dots dim along with the list while the detail pane is open.
	paginatorView := m.list.PaginatorView()
	if m.wideDetailOpen() {
		paginatorView = m.list.DimmedPaginatorView()
	}

	switch {
	case m.fetch.inFlight():
		// A story or link fetch in the wide layout keeps the paginator so the
		// left pane doesn't change; the loading state shows in the detail pane.
		// A search fetch shows no placeholder — its result is a single page.
		switch {
		case m.isWide() && (m.detailLoading() || m.fetch.linkLoading()):
			rightContent = paginatorView
		case !m.cat.Searching():
			rightContent = m.list.InactiveDots(m.config.PageMultiplier)
		}
	case m.screen == screenComments:
		// Full screen, the comment view handles its own footer; in the wide
		// layout the list keeps its paginator next to it.
		if m.isWide() {
			rightContent = paginatorView
		}
	default:
		rightContent = paginatorView
	}

	left := m.statusLeftStyle.Width(leftWidth).MaxWidth(leftWidth).Render("")

	center := midStyle.
		Width(m.listWidth() - leftWidth - statusBarEdgeWidth).
		Render(centerContent)

	right := m.statusEndStyle.Render(rightContent)

	// The fixed edge slots overflow panes narrower than their sum; the center
	// can too when its width comes out non-positive and lipgloss renders the
	// message at its natural width.
	row := xansi.Truncate(left+center+right, m.listWidth(), "")

	return underline + "\n" + row
}

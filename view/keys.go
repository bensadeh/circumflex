package view

import (
	"fmt"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) handleBrowsing(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		m.list.ClampCursor()

		return nil
	}

	numItems := len(m.list.VisibleItems())

	switch {
	case m.prompt == promptSearch:
		return m.handleSearchPromptKey(keyMsg)

	case m.prompt == promptAddFavorite && key.Matches(keyMsg, m.keymap.Confirm):
		return m.handleConfirmAddFavorites()

	case m.prompt == promptRemoveFavorite && key.Matches(keyMsg, m.keymap.Confirm):
		return m.handleConfirmRemoveFavorites()

	case m.prompt != promptNone:
		return m.handleCancelPrompt()

	case key.Matches(keyMsg, m.keymap.Help):
		m.screen = screenHelp

		return nil

	case key.Matches(keyMsg, m.keymap.ToggleWide):
		return m.toggleWideLayout()

	case key.Matches(keyMsg, m.keymap.Quit):
		if m.cat.Searching() {
			m.exitSearch()

			return nil
		}

		return tea.Quit

	case key.Matches(keyMsg, m.keymap.Up):
		m.list.CursorUp()

		return nil

	case key.Matches(keyMsg, m.keymap.Down):
		m.list.CursorDown()

		return nil

	case key.Matches(keyMsg, m.keymap.PrevPage):
		m.list.PrevPage()

		return nil

	case key.Matches(keyMsg, m.keymap.NextPage):
		m.list.NextPage()

		return nil

	// Tabbing out of search rejoins the cycle at the tab search was entered
	// from, advancing as usual.
	case key.Matches(keyMsg, m.keymap.NextCategory):
		m.cat.ExitSearch()

		return m.handleTabForward()

	case key.Matches(keyMsg, m.keymap.PrevCategory):
		m.cat.ExitSearch()

		return m.handleTabBackward()

	case key.Matches(keyMsg, m.keymap.GoToTop):
		m.list.GoToTop()

		return nil

	case key.Matches(keyMsg, m.keymap.GoToBottom):
		m.list.GoToBottom()

		return nil

	case key.Matches(keyMsg, m.keymap.Back):
		if m.cat.Searching() {
			m.exitSearch()
		}

		return nil

	case key.Matches(keyMsg, m.keymap.Search):
		if !m.cat.Searching() {
			m.cat.EnterSearch()
			m.list.ResetPager()
		}

		m.prompt = promptSearch
		m.searchPrompt.Start()

		return nil

	case key.Matches(keyMsg, m.keymap.SearchSort) && categories.IsSearch(m.cat.CurrentCategory()):
		m.searchFilters.cycleSort()

		return m.rerunSearch()

	case key.Matches(keyMsg, m.keymap.SearchAge) && categories.IsSearch(m.cat.CurrentCategory()):
		m.searchFilters.cycleAge()

		return m.rerunSearch()

	case key.Matches(keyMsg, m.keymap.Refresh):
		if categories.IsSearch(m.cat.CurrentCategory()) {
			return m.rerunSearch()
		}

		if !categories.IsFavorites(m.cat.CurrentCategory()) {
			return m.handleRefresh()
		}

	default:
		// The remaining keys all act on the highlighted story, so they do
		// nothing when the list is empty (e.g. the favorites tab with no
		// items). Guarding once here means any item action added below is
		// covered automatically.
		if numItems == 0 {
			break
		}

		switch {
		case key.Matches(keyMsg, m.keymap.OpenLink):
			return m.handleOpenLink()

		case key.Matches(keyMsg, m.keymap.OpenComments):
			return m.handleOpenComments()

		case key.Matches(keyMsg, m.keymap.AddFavorite):
			m.status.SetPermanentStatusMessage(addItemConfirmationMessage())
			m.prompt = promptAddFavorite

			return nil

		case key.Matches(keyMsg, m.keymap.RemoveFavorite):
			if m.cat.CurrentCategory() == categories.Favorites {
				m.status.SetPermanentStatusMessage(removeItemConfirmationMessage())
				m.prompt = promptRemoveFavorite

				return nil
			}

		case key.Matches(keyMsg, m.keymap.ToggleRead):
			if m.cat.CurrentCategory() != categories.Favorites {
				return m.handleToggleRead()
			}

		case key.Matches(keyMsg, m.keymap.EnterComments):
			return m.handleEnterComments()

		case key.Matches(keyMsg, m.keymap.ReaderMode):
			return m.handleEnterReaderMode()
		}
	}

	m.list.ClampCursor()

	return nil
}

func (m *model) handleConfirmAddFavorites() tea.Cmd {
	m.prompt = promptNone

	selectedItem := m.list.SelectedItem()
	addToFavorites := func() tea.Msg {
		return message.AddToFavorites{Item: selectedItem}
	}

	return tea.Batch(addToFavorites, m.status.NewStatusMessageWithDuration("Item added", statusMessageShort))
}

func (m *model) handleConfirmRemoveFavorites() tea.Cmd {
	m.prompt = promptNone

	removedItem := m.favorites.Items()[m.list.Index()]

	if err := m.favorites.Remove(m.list.Index()); err != nil {
		return m.status.NewStatusMessageWithDuration("Could not remove favorite", statusMessageLong)
	}

	if err := m.favorites.Write(); err != nil {
		m.favorites.Add(removedItem)
		m.syncFavorites()

		return m.status.NewStatusMessageWithDuration("Could not save favorites to disk", statusMessageLong)
	}

	m.syncFavorites()

	favItems := m.list.Items(categories.Favorites)
	isEmpty := len(favItems) == 0
	isOnLastItem := m.list.Index() == len(favItems)

	// Removing the last favorite leaves the (now empty) favorites tab in place
	// rather than jumping to another category.
	if isEmpty {
		m.list.SetCursor(0)
		m.list.SetPage(0)
		m.updatePagination()

		return m.status.NewStatusMessageWithDuration("Item removed", statusMessageShort)
	}

	if isOnLastItem {
		m.list.SetIndex(m.list.Index() - 1)
	}

	m.updatePagination()

	return m.status.NewStatusMessageWithDuration("Item removed", statusMessageShort)
}

func (m *model) handleCancelPrompt() tea.Cmd {
	m.prompt = promptNone

	return m.status.NewStatusMessageWithDuration(pane.CancelledStatus(), statusMessageShort)
}

func (m *model) handleTabForward() tea.Cmd {
	return m.handleTab(m.cat.NextIndex(), m.cat.NextCategory(), m.cat.Next)
}

func (m *model) handleTabBackward() tea.Cmd {
	return m.handleTab(m.cat.PrevIndex(), m.cat.PrevCategory(), m.cat.Prev)
}

func (m *model) handleTab(targetIndex int, targetCategory categories.Category, advance func()) tea.Cmd {
	// Favorites is served locally and never fetched, so switch to it directly
	// even when empty.
	if categories.IsFavorites(targetCategory) || m.list.HasItems(targetCategory) {
		advance()
		m.list.ResetPager()

		return nil
	}

	// The rollback point is the category being left, captured before advance.
	tok, startSpinnerCmd := m.startFetch(m.listRollback())

	pane.SetProgressIndeterminate()

	m.list.BeginTransition()
	advance()

	return tea.Batch(startSpinnerCmd, m.fetchCategory(tok, targetCategory, targetIndex, m.list.Cursor()))
}

func (m *model) handleOpenLink() tea.Cmd {
	selected := m.list.SelectedItem()

	url := selected.URL
	if url == "" {
		url = hn.ItemURL(selected.ID)
	}

	return message.OpenInBrowser(url)
}

func (m *model) handleOpenComments() tea.Cmd {
	return message.OpenInBrowser(hn.ItemURL(m.list.SelectedItem().ID))
}

// handleSearchPromptKey feeds one key press to the open search prompt; a
// committed query starts the search fetch. Cancelling with nothing committed
// leaves search altogether — the / was a false start — while cancelling over
// existing results keeps them.
func (m *model) handleSearchPromptKey(msg tea.KeyPressMsg) tea.Cmd {
	switch m.searchPrompt.HandleKey(msg) {
	case pane.PromptCommitted:
		m.prompt = promptNone

		return m.runSearch(m.searchPrompt.Text())

	case pane.PromptCanceled:
		m.prompt = promptNone

		if m.searchQuery == "" {
			m.exitSearch()
		}

	case pane.PromptPending:
		// The prompt renders live from the model; nothing to do.
	}

	return nil
}

// exitSearch leaves search mode for the tab it was entered from, clearing
// the query, results and filters — every / starts a fresh search.
func (m *model) exitSearch() {
	m.cat.ExitSearch()
	m.searchQuery = ""
	m.searchFilters = searchFilters{}
	m.list.SetItems(categories.Search, nil)
	m.list.ResetPager()
}

// runSearch fetches the stories matching query, replacing the search tab's
// results. Old results stay on screen, dimmed, until the new ones arrive;
// a failed or cancelled fetch keeps them.
func (m *model) runSearch(query string) tea.Cmd {
	m.searchQuery = query

	// Captured before the cursor reset below, so a failed search keeps the
	// old results with the selection where it was.
	rb := m.listRollback()

	m.list.BeginTransition()
	m.list.SetPage(0)
	m.list.SetCursor(0)
	m.updatePagination()

	tok, startSpinnerCmd := m.startFetch(rb)

	pane.SetProgressIndeterminate()

	return tea.Batch(startSpinnerCmd, m.fetchSearch(tok, query, m.cat.CurrentIndex()))
}

// rerunSearch re-runs the committed query, picking up the current filters —
// refresh and every filter change route through here. With nothing committed
// yet there is nothing to re-run; the filter change still shows in the
// readout.
func (m *model) rerunSearch() tea.Cmd {
	if m.searchQuery == "" {
		return nil
	}

	return m.runSearch(m.searchQuery)
}

func (m *model) handleRefresh() tea.Cmd {
	currentCategory := m.cat.CurrentCategory()
	currentIndex := m.cat.CurrentIndex()

	// Captured before the cursor reset below, so a failed refresh puts the
	// selection back where it was.
	rb := m.listRollback()

	m.list.BeginTransition()
	m.updatePagination()

	// Stay on the same page during the transition; the cursor resets to the top.
	m.list.SetPage(rb.page)
	m.list.SetCursor(0)

	tok, startSpinnerCmd := m.startFetch(rb)

	pane.SetProgressIndeterminate()

	return tea.Batch(startSpinnerCmd, m.fetchCategory(tok, currentCategory, currentIndex, 0), fetchMemorialStatus())
}

func (m *model) handleEnterComments() tea.Cmd {
	return m.openComments(m.list.Index())
}

func (m *model) handleEnterReaderMode() tea.Cmd {
	selected := m.list.SelectedItem()

	if err := article.Validate(selected.Title, selected.Domain); err != nil {
		return m.showDetailError(err, screenReader)
	}

	return m.openReader(m.list.Index())
}

// openComments starts the comment fetch for the selected story. rollbackStory
// is the selection to restore on failure or cancel: the story the screen
// keeps showing, which for J/K navigation is the story being left rather
// than the selected one.
func (m *model) openComments(rollbackStory int) tea.Cmd {
	selected := m.list.SelectedItem()

	tok, startSpinnerCmd := m.startDetailFetch(0, screenComments, m.detailRollback(rollbackStory))
	// The comment fetch reports percentages, so its indicator starts at 0%
	// instead of flashing indeterminate first.
	pane.SetProgressPercent(0)

	return tea.Batch(startSpinnerCmd, m.fetchComments(tok, selected))
}

func (m *model) openReader(rollbackStory int) tea.Cmd {
	selected := m.list.SelectedItem()

	tok, startSpinnerCmd := m.startDetailFetch(pane.ReaderFetchTimeout, screenReader, m.detailRollback(rollbackStory))

	pane.SetProgressIndeterminate()

	return tea.Batch(startSpinnerCmd, m.fetchArticle(tok, selected))
}

// handleOpenAdjacentStory moves the selection one story up or down and opens
// it in the view the request came from, so the comment section and reader
// can page through the front page without going back to it.
func (m *model) handleOpenAdjacentStory(msg message.OpenAdjacentStory) tea.Cmd {
	// The detail view mints this message a cycle after its keypress, so a
	// rapid second press slips past the in-flight key gate and lands here
	// mid-fetch: acting on it would move the selection again and record the
	// half-open story as the rollback point.
	if m.fetch.inFlight() {
		return nil
	}

	fromReader := m.screen == screenReader
	if !fromReader && m.screen != screenComments {
		return nil
	}

	items := m.list.VisibleItems()
	newIndex := m.list.Index() + msg.Direction

	if newIndex < 0 || newIndex >= len(items) {
		return nil
	}

	// Validate before moving so in the narrow layout a story the reader
	// can't open leaves the current story open and the selection in place.
	// The wide layout swaps the reader for the error view, so there the
	// selection lands on the story that failed and J/K page on from it.
	if fromReader {
		if err := article.Validate(items[newIndex].Title, items[newIndex].Domain); err != nil {
			if m.isWide() {
				m.list.SetIndex(newIndex)
			}

			return m.showDetailError(err, screenReader)
		}
	}

	previousIndex := m.list.Index()
	m.list.SetIndex(newIndex)

	if fromReader {
		return m.openReader(previousIndex)
	}

	return m.openComments(previousIndex)
}

func (m *model) handleToggleRead() tea.Cmd {
	item := m.list.SelectedItem()

	if m.history.Contains(item.ID) {
		if err := m.history.MarkUnread(item.ID); err != nil {
			return m.status.NewStatusMessageWithDuration(
				fmt.Sprintf("Could not mark as unread: %s", err), statusMessageShort)
		}

		return m.status.NewStatusMessageWithDuration(
			"Marked as "+lipgloss.NewStyle().Foreground(lipgloss.Yellow).Render("unread"), statusMessageShort)
	}

	if err := m.history.MarkRead(item.ID, item.CommentsCount); err != nil {
		return m.status.NewStatusMessageWithDuration(
			fmt.Sprintf("Could not mark as read: %s", err), statusMessageShort)
	}

	return m.status.NewStatusMessageWithDuration(
		"Marked as "+lipgloss.NewStyle().Foreground(lipgloss.Magenta).Render("read"), statusMessageShort)
}

func addItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	green := normal.Foreground(lipgloss.Green)
	bold := normal.Foreground(lipgloss.Blue).Bold(true)

	return green.Render("Add") + normal.Render(" to Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

func removeItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	red := normal.Foreground(lipgloss.Red)
	bold := normal.Foreground(lipgloss.Blue).Bold(true)

	return red.Render("Remove") + normal.Render(" from Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

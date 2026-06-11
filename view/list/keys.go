package list

import (
	"context"
	"fmt"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/timeago"
	"github.com/bensadeh/circumflex/view/message"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const readerModeTimeout = 15 * time.Second

func (m *Model) handleBrowsing(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	numItems := len(m.VisibleItems())

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case m.state == stateAddFavoritesPrompt && key.Matches(msg, m.keymap.Confirm):
			return m.handleConfirmAddFavorites()

		case m.state == stateRemoveFavoritesPrompt && key.Matches(msg, m.keymap.Confirm):
			return m.handleConfirmRemoveFavorites()

		case m.state == stateAddFavoritesPrompt || m.state == stateRemoveFavoritesPrompt:
			return m.handleCancelPrompt()

		case m.state == stateFetching && key.Matches(msg, m.keymap.Cancel):
			return m.handleCancelFetch()

		case m.state != stateBrowsing:
			return nil

		case key.Matches(msg, m.keymap.Help):
			m.state = stateHelpScreen

			return nil

		case key.Matches(msg, m.keymap.Quit):
			clearProgress()

			return tea.Quit

		case key.Matches(msg, m.keymap.Up):
			m.pager.CursorUp()

			return nil

		case key.Matches(msg, m.keymap.Down):
			m.pager.CursorDown(m.cat.CurrentCategory())

			return nil

		case key.Matches(msg, m.keymap.PrevPage):
			m.pager.Paginator.PrevPage()
			m.pager.updateCursor(m.cat.CurrentCategory())

			return nil

		case key.Matches(msg, m.keymap.NextPage):
			m.pager.Paginator.NextPage()
			m.pager.updateCursor(m.cat.CurrentCategory())

			return nil

		case key.Matches(msg, m.keymap.NextCategory):
			return m.handleTabForward()

		case key.Matches(msg, m.keymap.PrevCategory):
			return m.handleTabBackward()

		case key.Matches(msg, m.keymap.GoToTop):
			m.pager.cursor = 0

			return nil

		case key.Matches(msg, m.keymap.GoToBottom):
			m.pager.cursor = max(0, m.pager.Paginator.ItemsOnPage(numItems)-1)

			return nil

		case key.Matches(msg, m.keymap.Refresh):
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
			case key.Matches(msg, m.keymap.OpenLink):
				return m.handleOpenLink()

			case key.Matches(msg, m.keymap.OpenComments):
				return m.handleOpenComments()

			case key.Matches(msg, m.keymap.AddFavorite):
				m.status.SetPermanentStatusMessage(addItemConfirmationMessage())
				m.state = stateAddFavoritesPrompt

				return nil

			case key.Matches(msg, m.keymap.RemoveFavorite):
				if m.cat.CurrentCategory() == categories.Favorites {
					m.status.SetPermanentStatusMessage(removeItemConfirmationMessage())
					m.state = stateRemoveFavoritesPrompt

					return nil
				}

			case key.Matches(msg, m.keymap.ToggleRead):
				if m.cat.CurrentCategory() != categories.Favorites {
					return m.handleToggleRead()
				}

			case key.Matches(msg, m.keymap.EnterComments):
				return m.handleEnterComments()

			case key.Matches(msg, m.keymap.ReaderMode):
				return m.handleEnterReaderMode()
			}
		}
	}

	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))
	if m.pager.cursor > itemsOnPage-1 {
		m.pager.cursor = max(0, itemsOnPage-1)
	}

	return tea.Batch(cmds...)
}

func (m *Model) handleConfirmAddFavorites() tea.Cmd {
	m.state = stateBrowsing

	selectedItem := m.SelectedItem()
	addToFavorites := func() tea.Msg {
		return message.AddToFavorites{Item: selectedItem}
	}

	return tea.Batch(addToFavorites, m.status.NewStatusMessageWithDuration("Item added", statusMessageShort))
}

func (m *Model) handleConfirmRemoveFavorites() tea.Cmd {
	m.state = stateBrowsing

	removedItem := m.favorites.Items()[m.Index()]

	if err := m.favorites.Remove(m.Index()); err != nil {
		return m.status.NewStatusMessageWithDuration("Could not remove favorite", statusMessageLong)
	}

	if err := m.favorites.Write(); err != nil {
		m.favorites.Add(removedItem)
		m.syncFavorites()

		return m.status.NewStatusMessageWithDuration("Could not save favorites to disk", statusMessageLong)
	}

	m.syncFavorites()

	isEmpty := len(m.pager.items[categories.Favorites]) == 0
	isOnLastItem := m.Index() == len(m.pager.items[categories.Favorites])

	// Removing the last favorite leaves the (now empty) favorites tab in place
	// rather than jumping to another category.
	if isEmpty {
		m.pager.cursor = 0
		m.pager.Paginator.Page = 0
		m.updatePagination()

		return m.status.NewStatusMessageWithDuration("Item removed", statusMessageShort)
	}

	if isOnLastItem {
		m.pager.cursor--
	}

	m.updatePagination()

	return m.status.NewStatusMessageWithDuration("Item removed", statusMessageShort)
}

func (m *Model) handleCancelPrompt() tea.Cmd {
	m.state = stateBrowsing

	return m.status.NewStatusMessageWithDuration(
		lipgloss.NewStyle().Faint(true).Render("Cancelled"), statusMessageShort)
}

func (m *Model) startFetch(timeout time.Duration) tea.Cmd {
	if m.cancelFetch != nil {
		m.cancelFetch()
	}

	m.fetchID++
	m.state = stateFetching

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		m.fetchCtx = ctx
		m.cancelFetch = cancel
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		m.fetchCtx = ctx
		m.cancelFetch = cancel
	}

	return m.status.StartSpinner()
}

// CancelFetch cancels an in-progress fetch and returns the resulting command.
// Returns nil if no fetch is active.
func (m *Model) CancelFetch() tea.Cmd {
	if m.state != stateFetching {
		return nil
	}

	return m.handleCancelFetch()
}

func (m *Model) handleCancelFetch() tea.Cmd {
	if m.cancelFetch != nil {
		m.cancelFetch()
		m.cancelFetch = nil
	}

	m.fetchID++

	if m.pager.transition != nil {
		m.cat.SetIndex(m.pager.transition.prevIndex)
		m.pager.transition = nil
	}

	clearProgress()
	m.status.StopSpinner()
	m.state = stateBrowsing
	m.updatePagination()

	return m.status.NewStatusMessageWithDuration(
		lipgloss.NewStyle().Faint(true).Render("Cancelled"), statusMessageShort)
}

func (m *Model) handleTabForward() tea.Cmd {
	return m.handleTab(m.cat.NextIndex(), m.cat.NextCategory(), m.cat.Next)
}

func (m *Model) handleTabBackward() tea.Cmd {
	return m.handleTab(m.cat.PrevIndex(), m.cat.PrevCategory(), m.cat.Prev)
}

func (m *Model) handleTab(targetIndex int, targetCategory categories.Category, advance func()) tea.Cmd {
	// Favorites is served locally and never fetched, so switch to it directly
	// even when empty.
	if categories.IsFavorites(targetCategory) || m.pager.categoryHasStories(targetCategory) {
		advance()
		m.resetPager()

		return nil
	}

	m.beginTransition()
	advance()

	startSpinnerCmd := m.startFetch(0)

	cursor := m.pager.cursor
	changeCatCmd := func() tea.Msg {
		return message.FetchAndChangeToCategory{Index: targetIndex, Category: targetCategory, Cursor: cursor}
	}

	return tea.Batch(startSpinnerCmd, changeCatCmd)
}

func (m *Model) handleOpenLink() tea.Cmd {
	selected := m.SelectedItem()

	url := selected.URL
	if url == "" {
		url = hn.ItemURL(selected.ID)
	}

	return message.OpenInBrowser(url)
}

func (m *Model) handleOpenComments() tea.Cmd {
	return message.OpenInBrowser(hn.ItemURL(m.SelectedItem().ID))
}

func (m *Model) handleRefresh() tea.Cmd {
	currentCategory := m.cat.CurrentCategory()
	currentIndex := m.cat.CurrentIndex()
	currentPage := m.pager.Paginator.Page

	m.beginTransition()
	m.updatePagination()

	// Stay on the same page during the transition; the cursor resets to the top.
	m.pager.Paginator.Page = currentPage
	m.pager.cursor = 0

	startSpinnerCmd := m.startFetch(0)

	refreshCmd := func() tea.Msg {
		return message.Refresh{CurrentIndex: currentIndex, CurrentCategory: currentCategory}
	}

	return tea.Batch(startSpinnerCmd, refreshCmd)
}

func (m *Model) handleEnterComments() tea.Cmd {
	selected := m.SelectedItem()

	m.beginTransition()
	startSpinnerCmd := m.startFetch(0)

	enterCommentsCmd := func() tea.Msg {
		return message.EnteringCommentSection{
			ID:           selected.ID,
			CommentCount: selected.CommentsCount,
		}
	}

	return tea.Batch(startSpinnerCmd, enterCommentsCmd)
}

func (m *Model) handleEnterReaderMode() tea.Cmd {
	selected := m.SelectedItem()

	if err := article.Validate(selected.Title, selected.Domain); err != nil {
		return m.status.NewStatusMessageWithDuration(friendlyError(err), statusMessageLong)
	}

	m.beginTransition()
	startSpinnerCmd := m.startFetch(readerModeTimeout)

	enterReaderCmd := func() tea.Msg {
		return message.EnteringReaderMode{
			URL:          selected.URL,
			Title:        selected.Title,
			Domain:       selected.Domain,
			ID:           selected.ID,
			CommentCount: selected.CommentsCount,
			Author:       selected.Author,
			TimeAgo:      timeago.RelativeTime(selected.Time),
			Points:       selected.Points,
		}
	}

	return tea.Batch(startSpinnerCmd, enterReaderCmd)
}

func (m *Model) beginTransition() {
	m.pager.transition = &transition{
		prevIndex: m.cat.CurrentIndex(),
		oldItems:  m.pager.items[m.cat.CurrentCategory()],
	}
}

func (m *Model) resetPager() {
	currentCategory := m.cat.CurrentCategory()

	m.pager.Paginator.Page = 0
	m.pager.cursor = max(0, min(m.pager.cursor, len(m.pager.items[currentCategory])-1))
	m.updatePagination()
}

func (m *Model) handleToggleRead() tea.Cmd {
	item := m.SelectedItem()

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

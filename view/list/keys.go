package list

import (
	"context"
	"strconv"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/browser"
	"github.com/bensadeh/circumflex/categories"
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
		case m.state == StateAddFavoritesPrompt && key.Matches(msg, m.keymap.Confirm):
			return m.handleConfirmAddFavorites()

		case m.state == StateRemoveFavoritesPrompt && key.Matches(msg, m.keymap.Confirm):
			return m.handleConfirmRemoveFavorites()

		case m.state == StateAddFavoritesPrompt || m.state == StateRemoveFavoritesPrompt:
			return m.handleCancelPrompt()

		case m.state == StateFetching && key.Matches(msg, m.keymap.Cancel):
			return m.handleCancelFetch()

		case m.state != StateBrowsing:
			return nil

		case key.Matches(msg, m.keymap.Help):
			m.state = StateHelpScreen

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
			m.pager.cursor = m.pager.Paginator.ItemsOnPage(numItems) - 1

			return nil

		case key.Matches(msg, m.keymap.OpenLink):
			return m.handleOpenLink()

		case key.Matches(msg, m.keymap.OpenComments):
			return m.handleOpenComments()

		case key.Matches(msg, m.keymap.Refresh):
			if m.cat.CurrentCategory() != categories.Favorites {
				return m.handleRefresh()
			}

		case key.Matches(msg, m.keymap.AddFavorite):
			m.status.SetPermanentStatusMessage(getAddItemConfirmationMessage(), false)
			m.state = StateAddFavoritesPrompt

			return nil

		case key.Matches(msg, m.keymap.RemoveFavorite):
			if m.cat.CurrentCategory() == categories.Favorites {
				m.status.SetPermanentStatusMessage(getRemoveItemConfirmationMessage(), false)
				m.state = StateRemoveFavoritesPrompt

				return nil
			}

		case key.Matches(msg, m.keymap.ToggleRead):
			if m.cat.CurrentCategory() != categories.Favorites {
				return m.handleToggleRead()
			}

		case key.Matches(msg, m.keymap.EnterComments):
			currentCategory := m.cat.CurrentCategory()
			m.pager.transition = &transition{
				prevIndex: m.cat.CurrentIndex(),
				oldItems:  m.pager.items[currentCategory],
			}
			startSpinnerCmd := m.startFetch(0)

			id := m.SelectedItem().ID
			commentCount := m.SelectedItem().CommentsCount

			enterCommentsCmd := func() tea.Msg {
				return message.EnteringCommentSection{
					ID:           id,
					CommentCount: commentCount,
				}
			}

			return tea.Batch(startSpinnerCmd, enterCommentsCmd)

		case key.Matches(msg, m.keymap.ReaderMode):
			selected := m.SelectedItem()

			if err := article.Validate(selected.Title, selected.Domain); err != nil {
				return m.status.NewStatusMessageWithDuration(friendlyError(err), statusMessageLong)
			}

			currentCategory := m.cat.CurrentCategory()
			m.pager.transition = &transition{
				prevIndex: m.cat.CurrentIndex(),
				oldItems:  m.pager.items[currentCategory],
			}
			startSpinnerCmd := m.startFetch(readerModeTimeout)

			enterReaderCmd := func() tea.Msg {
				return message.EnteringReaderMode{
					URL:          selected.URL,
					Title:        selected.Title,
					Domain:       selected.Domain,
					ID:           selected.ID,
					CommentCount: selected.CommentsCount,
					Author:       selected.User,
					TimeAgo:      selected.TimeAgo,
					Points:       selected.Points,
				}
			}

			return tea.Batch(startSpinnerCmd, enterReaderCmd)
		}
	}

	// Epilogue: delegate + cursor clamp (only reached if no handler matched)
	cmd := m.delegate.Update(msg, m)
	cmds = append(cmds, cmd)

	// Keep the index in bounds when paginating
	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))
	if m.pager.cursor > itemsOnPage-1 {
		m.pager.cursor = max(0, itemsOnPage-1)
	}

	return tea.Batch(cmds...)
}

func (m *Model) handleConfirmAddFavorites() tea.Cmd {
	m.state = StateBrowsing

	selectedItem := m.SelectedItem()
	addToFavorites := func() tea.Msg {
		return message.AddToFavorites{Item: selectedItem}
	}

	return tea.Batch(addToFavorites, m.status.NewStatusMessageWithDuration("Item added", statusMessageShort))
}

func (m *Model) handleConfirmRemoveFavorites() tea.Cmd {
	m.state = StateBrowsing

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

	isOnLastItem := m.Index() == len(m.pager.items[categories.Favorites])
	hasOnlyOneItem := len(m.pager.items[categories.Favorites]) == 0

	itemRemovedMessage := "Item removed"

	if hasOnlyOneItem {
		m.cat.SetIndex(0)
		m.pager.updateCursor(m.cat.CurrentCategory())
		m.updatePagination()

		catIndex := m.cat.CurrentIndex()
		catValue := m.cat.CurrentCategory()
		changeCatCmd := func() tea.Msg {
			return message.FetchAndChangeToCategory{Index: catIndex, Category: catValue, Cursor: 0}
		}

		return tea.Batch(changeCatCmd, m.status.NewStatusMessageWithDuration(itemRemovedMessage, statusMessageShort))
	}

	if isOnLastItem {
		m.pager.cursor--
	}

	m.updatePagination()

	return m.status.NewStatusMessageWithDuration(itemRemovedMessage, statusMessageShort)
}

func (m *Model) handleCancelPrompt() tea.Cmd {
	m.state = StateBrowsing

	return m.status.NewStatusMessageWithDuration(
		lipgloss.NewStyle().Faint(true).Render("Cancelled"), statusMessageShort)
}

func (m *Model) startFetch(timeout time.Duration) tea.Cmd {
	if m.cancelFetch != nil {
		m.cancelFetch()
	}

	m.fetchID++
	m.state = StateFetching

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
	if m.state != StateFetching {
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

	m.status.StopSpinner()
	m.state = StateBrowsing
	m.updatePagination()

	return m.status.NewStatusMessageWithDuration(
		lipgloss.NewStyle().Faint(true).Render("Cancelled"), statusMessageShort)
}

func (m *Model) handleTabForward() tea.Cmd {
	return m.handleTab(
		m.cat.NextIndex(),
		m.cat.NextCategory(),
		m.changeToNextCategory,
		func() { m.cat.Next() },
	)
}

func (m *Model) handleTabBackward() tea.Cmd {
	return m.handleTab(
		m.cat.PrevIndex(),
		m.cat.PrevCategory(),
		m.changeToPrevCategory,
		func() { m.cat.Prev() },
	)
}

func (m *Model) handleTab(targetIndex int, targetCategory categories.Category, changeCategory func(), advance func()) tea.Cmd {
	if m.pager.categoryHasStories(targetCategory) {
		changeCategory()

		return nil
	}

	currentCategory := m.cat.CurrentCategory()
	m.pager.transition = &transition{
		prevIndex: m.cat.CurrentIndex(),
		oldItems:  m.pager.items[currentCategory],
	}

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
		url = "https://news.ycombinator.com/item?id=" + strconv.Itoa(selected.ID)
	}

	return func() tea.Msg {
		if err := browser.Open(context.Background(), url); err != nil {
			return message.BrowserOpenFailed{Err: err}
		}

		return nil
	}
}

func (m *Model) handleOpenComments() tea.Cmd {
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(m.SelectedItem().ID)

	return func() tea.Msg {
		if err := browser.Open(context.Background(), url); err != nil {
			return message.BrowserOpenFailed{Err: err}
		}

		return nil
	}
}

func (m *Model) handleRefresh() tea.Cmd {
	currentCategory := m.cat.CurrentCategory()
	currentIndex := m.cat.CurrentIndex()
	currentPage := m.pager.Paginator.Page

	m.pager.transition = &transition{
		prevIndex: currentIndex,
		oldItems:  m.pager.items[currentCategory],
		refresh:   true,
	}

	m.pager.Paginator.Page = 0
	m.pager.cursor = min(m.pager.cursor, len(m.pager.items[currentCategory])-1)
	m.updatePagination()

	m.pager.cursor = 0
	m.pager.Paginator.Page = currentPage

	startSpinnerCmd := m.startFetch(0)

	changeCatCmd := func() tea.Msg {
		return message.Refresh{CurrentIndex: currentIndex, CurrentCategory: currentCategory}
	}

	return tea.Batch(startSpinnerCmd, changeCatCmd)
}

func (m *Model) changeToNextCategory() {
	m.cat.Next()
	currentCategory := m.cat.CurrentCategory()

	m.pager.Paginator.Page = 0
	m.pager.cursor = min(m.pager.cursor, len(m.pager.items[currentCategory])-1)
	m.updatePagination()
}

func (m *Model) changeToPrevCategory() {
	m.cat.Prev()
	currentCategory := m.cat.CurrentCategory()

	m.pager.Paginator.Page = 0
	m.pager.cursor = min(m.pager.cursor, len(m.pager.items[currentCategory])-1)
	m.updatePagination()
}

func (m *Model) handleToggleRead() tea.Cmd {
	item := m.SelectedItem()

	if m.history.Contains(item.ID) {
		if err := m.history.MarkAsUnreadAndWriteToDisk(item.ID); err != nil {
			return m.status.NewStatusMessageWithDuration("Could not mark as unread", statusMessageShort)
		}

		return m.status.NewStatusMessageWithDuration(
			"Marked as "+lipgloss.NewStyle().Foreground(lipgloss.Yellow).Render("unread"), statusMessageShort)
	}

	if err := m.history.MarkAsReadAndWriteToDisk(item.ID, item.CommentsCount); err != nil {
		return m.status.NewStatusMessageWithDuration("Could not mark as read", statusMessageShort)
	}

	return m.status.NewStatusMessageWithDuration(
		"Marked as "+lipgloss.NewStyle().Foreground(lipgloss.Magenta).Render("read"), statusMessageShort)
}

func getAddItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	green := normal.Foreground(lipgloss.Green)
	bold := normal.Foreground(lipgloss.Blue).Bold(true)

	return green.Render("Add") + normal.Render(" to Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

func getRemoveItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	red := normal.Foreground(lipgloss.Red)
	bold := normal.Foreground(lipgloss.Blue).Bold(true)

	return red.Render("Remove") + normal.Render(" from Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

package list

import (
	"clx/browser"
	"clx/bubble/list/message"
	"clx/categories"
	"context"
	"strconv"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

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

		case m.state != StateBrowsing:
			return nil

		case key.Matches(msg, m.keymap.Help):
			m.state = StateHelpScreen

			return nil

		case key.Matches(msg, m.keymap.Quit):
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
			m.state = StateEditorOpen

			id := m.SelectedItem().ID
			commentCount := m.SelectedItem().CommentsCount

			return func() tea.Msg {
				return message.EnteringCommentSection{
					Id:           id,
					CommentCount: commentCount,
				}
			}

		case key.Matches(msg, m.keymap.ReaderMode):
			m.state = StateEditorOpen

			selected := m.SelectedItem()

			return func() tea.Msg {
				return message.EnteringReaderMode{
					Url:          selected.URL,
					Title:        selected.Title,
					Domain:       selected.Domain,
					Id:           selected.ID,
					CommentCount: selected.CommentsCount,
				}
			}
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

	return tea.Batch(addToFavorites, m.status.NewStatusMessageWithDuration("Item added", time.Second*2))
}

func (m *Model) handleConfirmRemoveFavorites() tea.Cmd {
	m.state = StateBrowsing

	removedItem := m.favorites.GetItems()[m.Index()]

	if err := m.favorites.Remove(m.Index()); err != nil {
		return m.status.NewStatusMessageWithDuration("Could not remove favorite", time.Second*3)
	}

	if err := m.favorites.Write(); err != nil {
		m.favorites.Add(removedItem)
		m.syncFavorites()

		return m.status.NewStatusMessageWithDuration("Could not save favorites to disk", time.Second*3)
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

		return tea.Batch(changeCatCmd, m.status.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2))
	}

	if isOnLastItem {
		m.pager.cursor--
	}

	m.updatePagination()

	return m.status.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2)
}

func (m *Model) handleCancelPrompt() tea.Cmd {
	m.state = StateBrowsing
	m.status.hideStatusMessage()

	return nil
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

func (m *Model) handleTab(targetIndex, targetCategory int, changeCategory func(), advance func()) tea.Cmd {
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

	m.state = StateFetching
	startSpinnerCmd := m.status.StartSpinner()

	cursor := m.pager.cursor
	changeCatCmd := func() tea.Msg {
		return message.FetchAndChangeToCategory{Index: targetIndex, Category: targetCategory, Cursor: cursor}
	}

	return tea.Batch(startSpinnerCmd, changeCatCmd)
}

func (m *Model) handleOpenLink() tea.Cmd {
	url := m.SelectedItem().URL
	if url == "" {
		url = "https://news.ycombinator.com/item?id=" + strconv.Itoa(m.SelectedItem().ID)
	}

	id := m.SelectedItem().ID
	commentCount := m.SelectedItem().CommentsCount

	return func() tea.Msg {
		if err := browser.Open(context.Background(), url); err != nil {
			return message.BrowserOpenFailed{Err: err}
		}

		return message.OpeningLink{
			Id:           id,
			CommentCount: commentCount,
		}
	}
}

func (m *Model) handleOpenComments() tea.Cmd {
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(m.SelectedItem().ID)
	id := m.SelectedItem().ID
	commentCount := m.SelectedItem().CommentsCount

	return func() tea.Msg {
		if err := browser.Open(context.Background(), url); err != nil {
			return message.BrowserOpenFailed{Err: err}
		}

		return message.OpeningCommentsInBrowser{
			Id:           id,
			CommentCount: commentCount,
		}
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

	m.state = StateFetching
	m.pager.cursor = 0
	m.pager.Paginator.Page = currentPage

	changeCatCmd := func() tea.Msg {
		return message.Refresh{CurrentIndex: currentIndex, CurrentCategory: currentCategory}
	}

	return tea.Batch(m.status.StartSpinner(), changeCatCmd)
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
			return m.status.NewStatusMessageWithDuration("Could not mark as unread", time.Second*2)
		}

		return m.status.NewStatusMessageWithDuration(
			"Marked as "+lipgloss.NewStyle().Foreground(lipgloss.Yellow).Render("unread"), time.Second*2)
	}

	if err := m.history.MarkAsReadAndWriteToDisk(item.ID, item.CommentsCount); err != nil {
		return m.status.NewStatusMessageWithDuration("Could not mark as read", time.Second*2)
	}

	return m.status.NewStatusMessageWithDuration(
		"Marked as "+lipgloss.NewStyle().Foreground(lipgloss.Magenta).Render("read"), time.Second*2)
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

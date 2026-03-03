package list

import (
	"strconv"
	"time"

	"clx/browser"
	"clx/bubble/list/message"
	"clx/constants/category"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) handleBrowsing(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	numItems := len(m.VisibleItems())

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case m.state == StateAddFavoritesPrompt && msg.String() == "y":
			return m.handleConfirmAddFavorites()

		case m.state == StateRemoveFavoritesPrompt && msg.String() == "y":
			return m.handleConfirmRemoveFavorites()

		case m.state == StateAddFavoritesPrompt || m.state == StateRemoveFavoritesPrompt:
			return m.handleCancelPrompt()

		case m.state != StateBrowsing:
			return nil

		case msg.String() == "i" || msg.String() == "?":
			m.state = StateHelpScreen

			return nil

		case msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c":
			return tea.Quit

		case msg.String() == "up" || msg.String() == "k":
			m.CursorUp()
			return nil

		case msg.String() == "down" || msg.String() == "j":
			m.CursorDown()
			return nil

		case msg.String() == "left" || msg.String() == "h":
			m.Paginator.PrevPage()
			m.updateCursor()
			return nil

		case msg.String() == "right" || msg.String() == "l":
			m.Paginator.NextPage()
			m.updateCursor()
			return nil

		case msg.String() == "tab":
			return m.handleTabForward()

		case msg.String() == "shift+tab":
			return m.handleTabBackward()

		case msg.String() == "g":
			m.cursor = 0
			return nil

		case msg.String() == "G":
			m.cursor = m.Paginator.ItemsOnPage(numItems) - 1
			return nil

		case msg.String() == "o":
			return m.handleOpenLink()

		case msg.String() == "c":
			return m.handleOpenComments()

		case msg.String() == "r" && m.cat.GetCurrentCategory(m.favorites.HasItems()) != category.Favorites:
			return m.handleRefresh()

		case msg.String() == "f" || msg.String() == "V":
			m.SetPermanentStatusMessage(getAddItemConfirmationMessage(), false)
			m.state = StateAddFavoritesPrompt
			return nil

		case msg.String() == "x" && m.cat.GetCurrentCategory(m.favorites.HasItems()) == category.Favorites:
			m.SetPermanentStatusMessage(getRemoveItemConfirmationMessage(), false)
			m.state = StateRemoveFavoritesPrompt
			return nil

		case msg.String() == "enter":
			m.isVisible = false
			m.state = StateEditorOpen

			cmd := func() tea.Msg {
				return message.EnteringCommentSection{
					Id:           m.SelectedItem().ID,
					CommentCount: m.SelectedItem().CommentsCount,
				}
			}

			return cmd

		case msg.String() == "space":
			m.isVisible = false
			m.state = StateEditorOpen

			return func() tea.Msg {
				return message.EnteringReaderMode{
					Url:          m.SelectedItem().URL,
					Title:        m.SelectedItem().Title,
					Domain:       m.SelectedItem().Domain,
					Id:           m.SelectedItem().ID,
					CommentCount: m.SelectedItem().CommentsCount,
				}
			}
		}
	}

	// Epilogue: delegate + cursor clamp (only reached if no handler matched)
	cmd := m.delegate.Update(msg, m)
	cmds = append(cmds, cmd)

	// Keep the index in bounds when paginating
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
	if m.cursor > itemsOnPage-1 {
		m.cursor = max(0, itemsOnPage-1)
	}

	return tea.Batch(cmds...)
}

func (m *Model) handleConfirmAddFavorites() tea.Cmd {
	m.state = StateBrowsing

	addToFavorites := func() tea.Msg {
		return message.AddToFavorites{Item: m.SelectedItem()}
	}

	return tea.Batch(addToFavorites, m.NewStatusMessageWithDuration("Item added", time.Second*2))
}

func (m *Model) handleConfirmRemoveFavorites() tea.Cmd {
	m.state = StateBrowsing

	m.favorites.Remove(m.Index())
	m.items[category.Favorites] = m.favorites.GetItems()
	m.favorites.Write()

	isOnLastItem := m.Index() == len(m.items[category.Favorites])
	hasOnlyOneItem := len(m.items[category.Favorites]) == 0

	itemRemovedMessage := "Item removed"

	if hasOnlyOneItem {
		m.cat.SetIndex(0)
		m.updateCursor()
		m.updatePagination()

		changeCatCmd := func() tea.Msg {
			return message.FetchAndChangeToCategory{Index: m.cat.GetCurrentIndex(), Category: m.cat.GetCurrentCategory(false), Cursor: 0}
		}

		return tea.Batch(changeCatCmd, m.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2))
	}

	if isOnLastItem {
		m.cursor = m.cursor - 1
	}

	m.updatePagination()

	return m.NewStatusMessageWithDuration(itemRemovedMessage, time.Second*2)
}

func (m *Model) handleCancelPrompt() tea.Cmd {
	m.state = StateBrowsing
	m.hideStatusMessage()
	return nil
}

func (m *Model) handleTabForward() tea.Cmd {
	nextIndex := m.cat.GetNextIndex(m.favorites.HasItems())
	nextCat := m.cat.GetNextCategory(m.favorites.HasItems())

	if m.categoryHasStories(nextCat) {
		m.changeToNextCategory()
		return nil
	}

	m.state = StateLoading
	startSpinnerCmd := m.StartSpinner()

	changeCatCmd := func() tea.Msg {
		return message.FetchAndChangeToCategory{Index: nextIndex, Category: nextCat, Cursor: m.cursor}
	}

	return tea.Batch(startSpinnerCmd, changeCatCmd)
}

func (m *Model) handleTabBackward() tea.Cmd {
	prevIndex := m.cat.GetPrevIndex(m.favorites.HasItems())
	prevCat := m.cat.GetPrevCategory(m.favorites.HasItems())

	if m.categoryHasStories(prevCat) {
		m.changeToPrevCategory()
		return nil
	}

	m.state = StateLoading
	startSpinnerCmd := m.StartSpinner()

	changeCatCmd := func() tea.Msg {
		return message.FetchAndChangeToCategory{Index: prevIndex, Category: prevCat, Cursor: m.cursor}
	}

	return tea.Batch(startSpinnerCmd, changeCatCmd)
}

func (m *Model) handleOpenLink() tea.Cmd {
	if m.SelectedItem().URL == "" {
		url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(m.SelectedItem().ID)
		browser.Open(url)
	} else {
		browser.Open(m.SelectedItem().URL)
	}

	return func() tea.Msg {
		return message.OpeningLink{
			Id:           m.SelectedItem().ID,
			CommentCount: m.SelectedItem().CommentsCount,
		}
	}
}

func (m *Model) handleOpenComments() tea.Cmd {
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(m.SelectedItem().ID)
	browser.Open(url)

	return func() tea.Msg {
		return message.OpeningCommentsInBrowser{
			Id:           m.SelectedItem().ID,
			CommentCount: m.SelectedItem().CommentsCount,
		}
	}
}

func (m *Model) handleRefresh() tea.Cmd {
	currentCategory := m.cat.GetCurrentCategory(m.favorites.HasItems())
	currentPage := m.Paginator.Page

	m.items[category.Buffer] = m.items[currentCategory]

	m.isBufferActive = true
	m.Paginator.Page = 0
	m.cursor = min(m.cursor, len(m.items[currentCategory])-1)
	m.updatePagination()

	m.state = StateLoading
	m.cursor = 0
	m.Paginator.Page = currentPage

	changeCatCmd := func() tea.Msg {
		return message.Refresh{CurrentIndex: m.cat.GetCurrentIndex(), CurrentCategory: currentCategory}
	}

	return tea.Batch(m.StartSpinner(), changeCatCmd)
}

func (m *Model) categoryHasStories(cat int) bool {
	return len(m.items[cat]) != 0
}

func (m *Model) changeToNextCategory() {
	m.cat.Next(m.favorites.HasItems())
	currentCategory := m.cat.GetCurrentCategory(m.favorites.HasItems())

	m.Paginator.Page = 0
	m.cursor = min(m.cursor, len(m.items[currentCategory])-1)
	m.updatePagination()
}

func (m *Model) changeToPrevCategory() {
	m.cat.Prev(m.favorites.HasItems())
	currentCategory := m.cat.GetCurrentCategory(m.favorites.HasItems())

	m.Paginator.Page = 0
	m.cursor = min(m.cursor, len(m.items[currentCategory])-1)
	m.updatePagination()
}

func getAddItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	green := normal.Copy().
		Foreground(lipgloss.Color("2"))
	bold := normal.Copy().
		Foreground(lipgloss.Color("4")).
		Bold(true)

	return green.Render("Add") + normal.Render(" to Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

func getRemoveItemConfirmationMessage() string {
	normal := lipgloss.NewStyle()
	red := normal.Copy().
		Foreground(lipgloss.Color("1"))
	bold := normal.Copy().
		Foreground(lipgloss.Color("4")).
		Bold(true)

	return red.Render("Remove") + normal.Render(" from Favorites? Press ") + bold.Render("y") +
		normal.Render(" to confirm")
}

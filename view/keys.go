package view

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

func (m *model) handleBrowsing(msg tea.Msg) tea.Cmd {
	numItems := len(m.list.VisibleItems())

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case m.prompt == promptAddFavorite && key.Matches(msg, m.keymap.Confirm):
			return m.handleConfirmAddFavorites()

		case m.prompt == promptRemoveFavorite && key.Matches(msg, m.keymap.Confirm):
			return m.handleConfirmRemoveFavorites()

		case m.prompt != promptNone:
			return m.handleCancelPrompt()

		case key.Matches(msg, m.keymap.Help):
			m.screen = screenHelp

			return nil

		case key.Matches(msg, m.keymap.Quit):
			clearProgress()

			return tea.Quit

		case key.Matches(msg, m.keymap.Up):
			m.list.CursorUp()

			return nil

		case key.Matches(msg, m.keymap.Down):
			m.list.CursorDown()

			return nil

		case key.Matches(msg, m.keymap.PrevPage):
			m.list.PrevPage()

			return nil

		case key.Matches(msg, m.keymap.NextPage):
			m.list.NextPage()

			return nil

		case key.Matches(msg, m.keymap.NextCategory):
			return m.handleTabForward()

		case key.Matches(msg, m.keymap.PrevCategory):
			return m.handleTabBackward()

		case key.Matches(msg, m.keymap.GoToTop):
			m.list.GoToTop()

			return nil

		case key.Matches(msg, m.keymap.GoToBottom):
			m.list.GoToBottom()

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
				m.prompt = promptAddFavorite

				return nil

			case key.Matches(msg, m.keymap.RemoveFavorite):
				if m.cat.CurrentCategory() == categories.Favorites {
					m.status.SetPermanentStatusMessage(removeItemConfirmationMessage())
					m.prompt = promptRemoveFavorite

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

	return m.status.NewStatusMessageWithDuration(
		lipgloss.NewStyle().Faint(true).Render("Cancelled"), statusMessageShort)
}

// startFetch begins a fetch's lifecycle: it invalidates any predecessor and
// captures the category index to restore if the fetch fails or is cancelled —
// so callers that switch category must call it before advancing. The initial
// terminal progress write is the caller's: indeterminate for fetches without
// granular progress, percentage for the comment fetch, which reports it.
func (m *model) startFetch(timeout time.Duration) tea.Cmd {
	if m.cancelFetch != nil {
		m.cancelFetch()
	}

	m.fetchID++
	m.fetching = true
	m.detailFetch = false
	m.rollbackIndex = m.cat.CurrentIndex()
	m.rollbackStory = m.list.Index()

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

// startDetailFetch is startFetch for a story's comments or article: the list
// stays in place, dimmed, while the detail loads.
func (m *model) startDetailFetch(timeout time.Duration) tea.Cmd {
	cmd := m.startFetch(timeout)
	m.detailFetch = true

	return cmd
}

// rollbackFetch recovers from a failed or cancelled fetch: it restores the
// category selection captured at startFetch and unfreezes the list. For a
// story fetch it also moves the list selection back to the story that is
// still open, so the reading marker never points at a story the detail
// view doesn't show. Category fetches keep their cursor: the transition
// mechanics restore it.
func (m *model) rollbackFetch() {
	m.cat.SetIndex(m.rollbackIndex)

	if m.detailFetch {
		m.list.SetIndex(m.rollbackStory)
	}

	m.list.EndTransition()
}

func (m *model) handleCancelFetch() tea.Cmd {
	if m.cancelFetch != nil {
		m.cancelFetch()
		m.cancelFetch = nil
	}

	m.fetchID++

	m.rollbackFetch()

	clearProgress()
	m.status.StopSpinner()
	// The screen stays where the fetch started: canceling a J/K story fetch
	// keeps the open story, canceling a category fetch keeps the front page.
	m.fetching = false
	m.updatePagination()

	return m.status.NewStatusMessageWithDuration(
		lipgloss.NewStyle().Faint(true).Render("Cancelled"), statusMessageShort)
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

	// startFetch first: it captures the rollback index, which must be the
	// category we are leaving, not the one we advance to.
	startSpinnerCmd := m.startFetch(0)

	setProgressIndeterminate()

	m.list.BeginTransition()
	advance()

	cursor := m.list.Cursor()
	changeCatCmd := func() tea.Msg {
		return message.FetchAndChangeToCategory{Index: targetIndex, Category: targetCategory, Cursor: cursor}
	}

	return tea.Batch(startSpinnerCmd, changeCatCmd)
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

func (m *model) handleRefresh() tea.Cmd {
	currentCategory := m.cat.CurrentCategory()
	currentIndex := m.cat.CurrentIndex()
	currentPage := m.list.Page()

	m.list.BeginTransition()
	m.updatePagination()

	// Stay on the same page during the transition; the cursor resets to the top.
	m.list.SetPage(currentPage)
	m.list.SetCursor(0)

	startSpinnerCmd := m.startFetch(0)

	setProgressIndeterminate()

	refreshCmd := func() tea.Msg {
		return message.Refresh{CurrentIndex: currentIndex, CurrentCategory: currentCategory}
	}

	return tea.Batch(startSpinnerCmd, refreshCmd)
}

func (m *model) handleEnterComments() tea.Cmd {
	selected := m.list.SelectedItem()

	startSpinnerCmd := m.startDetailFetch(0)
	// The comment fetch reports percentages, so its indicator starts at 0%
	// instead of flashing indeterminate first.
	setProgressPercent(0)

	enterCommentsCmd := func() tea.Msg {
		return message.EnteringCommentSection{
			ID:           selected.ID,
			CommentCount: selected.CommentsCount,
		}
	}

	return tea.Batch(startSpinnerCmd, enterCommentsCmd)
}

func (m *model) handleEnterReaderMode() tea.Cmd {
	selected := m.list.SelectedItem()

	if err := article.Validate(selected.Title, selected.Domain); err != nil {
		return m.showDetailError(err, screenReader)
	}

	startSpinnerCmd := m.startDetailFetch(readerModeTimeout)

	setProgressIndeterminate()

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

// handleOpenAdjacentStory moves the selection one story up or down and opens
// it in the view the request came from, so the comment section and reader
// can page through the front page without going back to it.
func (m *model) handleOpenAdjacentStory(msg message.OpenAdjacentStory) tea.Cmd {
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

	var cmd tea.Cmd
	if fromReader {
		cmd = m.handleEnterReaderMode()
	} else {
		cmd = m.handleEnterComments()
	}

	// startFetch ran after the selection moved, so it captured the incoming
	// story; the one to restore on failure is the story we are leaving.
	m.rollbackStory = previousIndex

	return cmd
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

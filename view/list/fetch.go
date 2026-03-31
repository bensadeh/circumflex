package list

import (
	"clx/article"
	"clx/categories"
	"clx/comment"
	"clx/history"
	"clx/item"
	"clx/layout"
	"clx/view/message"
	"context"
	"errors"
	"net"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var categoryEndpoints = map[categories.Category]string{
	categories.Top:    "topstories",
	categories.Newest: "newstories",
	categories.Ask:    "askstories",
	categories.Show:   "showstories",
	categories.Best:   "beststories",
}

func (m *Model) FetchStoriesForFirstCategory() tea.Cmd {
	categoryToFetch := m.cat.CurrentCategory()
	service := m.service
	numItems := m.getNumberOfItemsToFetch(categoryToFetch)
	endpoint := categoryEndpoints[categoryToFetch]

	return func() tea.Msg {
		stories, err := service.FetchItems(context.Background(), numItems, endpoint)

		return message.FetchingFinished{
			Stories:  stories,
			Category: categoryToFetch,
			Err:      err,
		}
	}
}

func (m *Model) getNumberOfItemsToFetch(cat categories.Category) int {
	switch cat {
	case categories.Top, categories.Newest, categories.Best:
		return m.pager.Paginator.PerPage * m.config.PageMultiplier

	case categories.Ask, categories.Show, categories.Favorites:
		// Ask and Show pools are ~150-200 IDs, so fetching multiple pages
		// would exceed the pool and waste requests. Always fetch 1 page.
		return m.pager.Paginator.PerPage
	}

	return m.pager.Paginator.PerPage
}

func getHistory(debugMode bool, doNotMarkAsRead bool) history.History {
	if debugMode {
		return history.NewMockHistory()
	}

	if doNotMarkAsRead {
		return history.NewNonPersistentHistory()
	}

	h, err := history.NewPersistentHistory()
	if err != nil {
		return history.NewNonPersistentHistory()
	}

	return h
}

func (m *Model) fetchAndChangeToCategory(msg message.FetchAndChangeToCategory) tea.Cmd {
	service := m.service
	numItems := m.getNumberOfItemsToFetch(msg.Category)
	endpoint := categoryEndpoints[msg.Category]
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		stories, err := service.FetchItems(ctx, numItems, endpoint)

		return message.CategoryFetchingFinished{
			Stories:  stories,
			Category: msg.Category,
			Index:    msg.Index,
			Cursor:   msg.Cursor,
			Err:      err,
			FetchID:  fetchID,
		}
	}
}

func (m *Model) refresh(msg message.Refresh) tea.Cmd {
	service := m.service
	numItems := m.getNumberOfItemsToFetch(msg.CurrentCategory)
	endpoint := categoryEndpoints[msg.CurrentCategory]
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		stories, err := service.FetchItems(ctx, numItems, endpoint)

		return message.CategoryFetchingFinished{
			Stories:  stories,
			Category: msg.CurrentCategory,
			Index:    msg.CurrentIndex,
			Cursor:   0,
			Err:      err,
			FetchID:  fetchID,
		}
	}
}

func (m *Model) handleEnteringCommentSection(msg message.EnteringCommentSection) tea.Cmd {
	isOnFavorites := m.cat.CurrentCategory() == categories.Favorites
	hist := m.history
	service := m.service
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		lastVisited := hist.CommentsLastVisited(msg.ID)
		_ = hist.MarkAsReadAndWriteToDisk(msg.ID, msg.CommentCount)

		story, err := service.FetchComments(ctx, msg.ID)

		var updatedStory *item.Story
		if err == nil && isOnFavorites {
			updatedStory = story
		}

		var thread *comment.Thread
		if err == nil {
			thread = comment.StoryToThread(story)
		}

		return message.CommentTreeDataReady{
			Thread:       thread,
			LastVisited:  lastVisited,
			UpdatedStory: updatedStory,
			Err:          err,
			FetchID:      fetchID,
		}
	}
}

func (m *Model) handleEnteringReaderMode(msg message.EnteringReaderMode) tea.Cmd {
	config := m.config
	hist := m.history
	ctx := m.fetchCtx
	fetchID := m.fetchID
	title := msg.Title
	readerWidth := m.readerWidth()

	return func() tea.Msg {
		if err := article.Validate(msg.Title, msg.Domain); err != nil {
			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		article, err := article.Fetch(ctx, msg.URL, readerWidth, config.IndentationSymbol)
		if err != nil {
			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		_ = hist.MarkArticleAsReadAndWriteToDisk(msg.ID)

		return message.ArticleReady{Content: article, Title: title, FetchID: fetchID}
	}
}

func (m *Model) readerWidth() int {
	w := m.width - 2*layout.ReaderViewLeftMargin
	if w <= 0 {
		return m.config.CommentWidth
	}

	return w
}

func isTimeout(err error) bool {
	var netErr net.Error

	return errors.As(err, &netErr) && netErr.Timeout()
}

var redText = lipgloss.NewStyle().Foreground(lipgloss.Red)

func friendlyError(err error) string {
	if isTimeout(err) {
		return "Timed out — check your connection and try again"
	}

	errStr := err.Error()
	if errStr == "" {
		return "Unknown error"
	}

	msg := strings.ToUpper(errStr[:1]) + errStr[1:]

	if before, after, ok := strings.Cut(msg, "status "); ok {
		msg = before + "status " + redText.Render(after)
	}

	return msg
}

func clearAllCategories(items [][]*item.Story) {
	items[categories.Top] = []*item.Story{}
	items[categories.Newest] = []*item.Story{}
	items[categories.Ask] = []*item.Story{}
	items[categories.Show] = []*item.Story{}
	items[categories.Best] = []*item.Story{}
}

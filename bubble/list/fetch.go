package list

import (
	"clx/bubble/list/message"
	"clx/categories"
	"clx/history"
	"clx/hn"
	"clx/item"
	"clx/reader"
	"clx/tree"
	"clx/validator"
	"context"
	"errors"
	"net"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var categoryEndpoints = map[int]string{
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

func (m *Model) getNumberOfItemsToFetch(cat int) int {
	switch cat {
	case categories.Top:
		return m.pager.Paginator.PerPage * 3

	case categories.Newest:
		return m.pager.Paginator.PerPage * 3

	case categories.Best:
		return m.pager.Paginator.PerPage * 3

	case categories.Ask:
		return m.pager.Paginator.PerPage

	case categories.Show:
		return m.pager.Paginator.PerPage

	default:
		return m.pager.Paginator.PerPage
	}
}

func getService(debugMode, debugFallible bool) hn.Service {
	return hn.NewService(debugMode, debugFallible)
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
	width := m.width
	isOnFavorites := m.cat.CurrentCategory() == categories.Favorites
	hist := m.history
	service := m.service
	config := m.config
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		lastVisited := hist.LastVisited(msg.Id)
		_ = hist.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		story, err := service.FetchComments(ctx, msg.Id)
		if err != nil {
			return message.CommentTreeReady{Err: err, FetchID: fetchID}
		}

		var updatedStory *item.Story
		if isOnFavorites {
			updatedStory = story
		}

		commentTree := tree.Print(story, config, width, lastVisited)

		return message.CommentTreeReady{Content: commentTree, UpdatedStory: updatedStory, FetchID: fetchID}
	}
}

func (m *Model) handleEnteringReaderMode(msg message.EnteringReaderMode) tea.Cmd {
	config := m.config
	hist := m.history
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		if err := validator.Validate(msg.Title, msg.Domain); err != nil {
			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		article, err := reader.Article(ctx, msg.Url, msg.Title, config.CommentWidth, config.IndentationSymbol)
		if err != nil {
			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		_ = hist.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		return message.ArticleReady{Content: article, FetchID: fetchID}
	}
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

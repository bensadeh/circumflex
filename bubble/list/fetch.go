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

	tea "charm.land/bubbletea/v2"
)

func (m *Model) FetchStoriesForFirstCategory() tea.Cmd {
	categoryToFetch := m.cat.CurrentCategory()
	service := m.service
	numItems := m.getNumberOfItemsToFetch(categoryToFetch)

	return func() tea.Msg {
		stories, err := service.FetchItems(numItems, categoryToFetch)

		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}

		return message.FetchingFinished{
			Stories:  stories,
			Category: categoryToFetch,
			Message:  errMsg,
		}
	}
}

func (m *Model) getNumberOfItemsToFetch(cat int) int {
	switch cat {
	case categories.Top:
		return m.Paginator.PerPage * 3

	case categories.Newest:
		return m.Paginator.PerPage * 3

	case categories.Best:
		return m.Paginator.PerPage * 3

	case categories.Ask:
		return m.Paginator.PerPage

	case categories.Show:
		return m.Paginator.PerPage

	default:
		return m.Paginator.PerPage
	}
}

func getService(debugMode bool) hn.Service {
	return hn.NewService(debugMode)
}

func getHistory(debugMode bool, doNotMarkAsRead bool) history.History {
	if debugMode {
		return history.NewMockHistory()
	}

	if doNotMarkAsRead {
		return history.NewNonPersistentHistory()
	}

	h, _ := history.NewPersistentHistory()

	return h
}

func (m *Model) fetchAndChangeToCategory(msg message.FetchAndChangeToCategory) tea.Cmd {
	service := m.service
	numItems := m.getNumberOfItemsToFetch(msg.Category)

	return func() tea.Msg {
		stories, err := service.FetchItems(numItems, msg.Category)

		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}

		return message.CategoryFetchingFinished{
			Stories:  stories,
			Category: msg.Category,
			Index:    msg.Index,
			Cursor:   msg.Cursor,
			Message:  errMsg,
		}
	}
}

func (m *Model) refresh(msg message.Refresh) tea.Cmd {
	service := m.service
	numItems := m.getNumberOfItemsToFetch(msg.CurrentCategory)

	return func() tea.Msg {
		stories, err := service.FetchItems(numItems, msg.CurrentCategory)

		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}

		return message.CategoryFetchingFinished{
			Stories:  stories,
			Category: msg.CurrentCategory,
			Index:    msg.CurrentIndex,
			Cursor:   0,
			Message:  errMsg,
		}
	}
}

func (m *Model) handleEnteringCommentSection(msg message.EnteringCommentSection) tea.Cmd {
	width := m.width
	isOnFavorites := m.cat.CurrentCategory() == categories.Favorites
	hist := m.history
	service := m.service
	config := m.config

	return func() tea.Msg {
		lastVisited := hist.GetLastVisited(msg.Id)
		_ = hist.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		story, err := service.FetchComments(msg.Id)
		if err != nil {
			return message.CommentTreeReady{Error: "Could not fetch comments: " + err.Error()}
		}

		var updatedStory *item.Story
		if isOnFavorites {
			updatedStory = story
		}

		commentTree := tree.Print(story, config, width, lastVisited)

		return message.CommentTreeReady{Content: commentTree, UpdatedStory: updatedStory}
	}
}

func (m *Model) handleEnteringReaderMode(msg message.EnteringReaderMode) tea.Cmd {
	config := m.config
	hist := m.history

	return func() tea.Msg {
		errorMessage := validator.GetErrorMessage(msg.Title, msg.Domain)
		if errorMessage != "" {
			return message.ArticleReady{Error: errorMessage}
		}

		article, err := reader.GetArticle(msg.Url, msg.Title, config.CommentWidth, config.IndentationSymbol)
		if err != nil {
			return message.ArticleReady{Error: "Could not read article in Reader Mode"}
		}

		_ = hist.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		return message.ArticleReady{Content: article}
	}
}

func clearAllCategories(items [][]*item.Story) {
	items[categories.Top] = []*item.Story{}
	items[categories.Newest] = []*item.Story{}
	items[categories.Ask] = []*item.Story{}
	items[categories.Show] = []*item.Story{}
	items[categories.Best] = []*item.Story{}
}

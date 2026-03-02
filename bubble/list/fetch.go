package list

import (
	"clx/bubble/list/message"
	"clx/constants/category"
	"clx/history"
	"clx/hn"
	"clx/hn/services/hybrid"
	"clx/hn/services/mock"
	"clx/item"
	"clx/reader"
	"clx/tree"
	"clx/validator"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) FetchStoriesForFirstCategory() tea.Cmd {
	categoryToFetch := m.cat.GetCurrentCategory(m.favorites.HasItems())
	itemsToFetch := m.getNumberOfItemsToFetch(categoryToFetch)

	return func() tea.Msg {
		stories, errMsg := m.service.FetchItems(itemsToFetch, categoryToFetch)

		return message.FetchingFinished{
			Stories:  stories,
			Category: categoryToFetch,
			Message:  errMsg,
		}
	}
}

func (m *Model) getNumberOfItemsToFetch(cat int) int {
	switch cat {
	case category.Top:
		return m.Paginator.PerPage * 3

	case category.New:
		return m.Paginator.PerPage * 3

	case category.Best:
		return m.Paginator.PerPage * 3

	case category.Ask:
		return m.Paginator.PerPage

	case category.Show:
		return m.Paginator.PerPage

	default:
		return m.Paginator.PerPage
	}
}

func getService(debugMode bool) hn.Service {
	if debugMode {
		return mock.Service{}
	}

	return &hybrid.Service{}
}

func getHistory(debugMode bool, doNotMarkAsRead bool) history.History {
	if debugMode {
		return history.NewMockHistory()
	}

	if doNotMarkAsRead {
		return history.NewNonPersistentHistory()
	}

	return history.NewPersistentHistory()
}

func (m *Model) fetchAndChangeToCategory(msg message.FetchAndChangeToCategory) tea.Cmd {
	return func() tea.Msg {
		itemsToFetch := m.getNumberOfItemsToFetch(msg.Category)
		stories, errMsg := m.service.FetchItems(itemsToFetch, msg.Category)

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
	return func() tea.Msg {
		itemsToFetch := m.getNumberOfItemsToFetch(msg.CurrentCategory)
		stories, errMsg := m.service.FetchItems(itemsToFetch, msg.CurrentCategory)

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
	isOnFavorites := m.cat.GetCurrentCategory(m.favorites.HasItems()) == category.Favorites

	return func() tea.Msg {
		lastVisited := m.history.GetLastVisited(msg.Id)
		m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		story := m.service.FetchComments(msg.Id)

		if isOnFavorites {
			m.favorites.UpdateStoryAndWriteToDisk(story)
		}

		commentTree := tree.Print(story, m.config, width, lastVisited)

		return message.CommentTreeReady{Content: commentTree}
	}
}

func (m *Model) handleEnteringReaderMode(msg message.EnteringReaderMode) tea.Cmd {
	return func() tea.Msg {
		errorMessage := validator.GetErrorMessage(msg.Title, msg.Domain)
		if errorMessage != "" {
			return message.ArticleReady{Error: errorMessage}
		}

		article, err := reader.GetArticle(msg.Url, msg.Title, m.config.CommentWidth, m.config.IndentationSymbol)
		if err != nil {
			return message.ArticleReady{Error: "Could not read article in Reader Mode"}
		}

		m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		return message.ArticleReady{Content: article}
	}
}

func clearAllCategories(items [][]*item.Item) {
	items[category.Top] = []*item.Item{}
	items[category.New] = []*item.Item{}
	items[category.Ask] = []*item.Item{}
	items[category.Show] = []*item.Item{}
	items[category.Best] = []*item.Item{}
}

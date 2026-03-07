package list

import (
	"clx/bubble/list/message"
	"clx/categories"
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

func (m *Model) fetchItems(cat int) ([]*item.Story, string) {
	stories, err := m.service.FetchItems(m.getNumberOfItemsToFetch(cat), cat)
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	return stories, errMsg
}

func (m *Model) FetchStoriesForFirstCategory() tea.Cmd {
	categoryToFetch := m.cat.GetCurrentCategory(m.favorites.HasItems())

	return func() tea.Msg {
		stories, errMsg := m.fetchItems(categoryToFetch)

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
	if debugMode {
		return mock.Service{}
	}

	return hybrid.NewService()
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
	return func() tea.Msg {
		stories, errMsg := m.fetchItems(msg.Category)

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
		stories, errMsg := m.fetchItems(msg.CurrentCategory)

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
	isOnFavorites := m.cat.GetCurrentCategory(m.favorites.HasItems()) == categories.Favorites

	return func() tea.Msg {
		lastVisited := m.history.GetLastVisited(msg.Id)
		_ = m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

		story, err := m.service.FetchComments(msg.Id)
		if err != nil {
			return message.CommentTreeReady{Error: "Could not fetch comments: " + err.Error()}
		}

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

		_ = m.history.MarkAsReadAndWriteToDisk(msg.Id, msg.CommentCount)

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

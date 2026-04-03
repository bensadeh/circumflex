package list

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/item"
	"github.com/bensadeh/circumflex/view/message"

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
		setProgressIndeterminate()

		stories, err := service.FetchItems(context.Background(), numItems, endpoint)
		if err != nil {
			setProgressError()
		} else {
			clearProgress()
		}

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
		setProgressIndeterminate()

		stories, err := service.FetchItems(ctx, numItems, endpoint)
		if err != nil {
			setProgressError()
		} else {
			clearProgress()
		}

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
		setProgressIndeterminate()

		stories, err := service.FetchItems(ctx, numItems, endpoint)
		if err != nil {
			setProgressError()
		} else {
			clearProgress()
		}

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

		onProgress := func(fetched, total int) {
			if total <= 0 {
				return
			}

			pct := min(fetched*100/total, 100)
			fmt.Fprintf(os.Stderr, "\033]9;4;1;%d\a", pct)
		}

		story, err := service.FetchComments(ctx, msg.ID, onProgress)
		if err != nil {
			setProgressError()

			return message.CommentTreeDataReady{
				Err:     err,
				FetchID: fetchID,
			}
		}

		clearProgress()

		_ = hist.MarkAsReadAndWriteToDisk(msg.ID, msg.CommentCount)

		var updatedStory *item.Story
		if isOnFavorites {
			updatedStory = story
		}

		return message.CommentTreeDataReady{
			Thread:       comment.StoryToThread(story),
			LastVisited:  lastVisited,
			UpdatedStory: updatedStory,
			FetchID:      fetchID,
		}
	}
}

func (m *Model) handleEnteringReaderMode(msg message.EnteringReaderMode) tea.Cmd {
	hist := m.history
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		if err := article.Validate(msg.Title, msg.Domain); err != nil {
			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		setProgressIndeterminate()

		parsed, err := article.Parse(ctx, msg.URL)
		if err != nil {
			setProgressError()

			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		clearProgress()

		_ = hist.MarkArticleAsReadAndWriteToDisk(msg.ID)

		return message.ArticleReady{
			Parsed:  parsed,
			Title:   msg.Title,
			URL:     msg.URL,
			Author:  msg.Author,
			TimeAgo: msg.TimeAgo,
			ID:      msg.ID,
			Points:  msg.Points,
			FetchID: fetchID,
		}
	}
}

// Terminal progress bar via OSC 9;4 (supported by Ghostty, ConEmu and others;
// silently ignored by terminals that don't recognise the sequence).
// Writes to stderr to avoid interfering with Bubble Tea's stdout.

func setProgressIndeterminate() { fmt.Fprint(os.Stderr, "\033]9;4;3;0\a") }
func setProgressError()         { fmt.Fprint(os.Stderr, "\033]9;4;2;100\a") }
func clearProgress()            { fmt.Fprint(os.Stderr, "\033]9;4;0\a") }

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

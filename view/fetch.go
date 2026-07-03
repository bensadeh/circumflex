package view

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/memorial"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) fetchStoriesForFirstCategory() tea.Cmd {
	categoryToFetch := m.cat.CurrentCategory()

	// Favorites is served locally — it is never fetched over the network. Hand
	// the already-synced items straight to the normal "fetch finished" path.
	if categories.IsFavorites(categoryToFetch) {
		stories := m.list.Items(categoryToFetch)
		fetchID := m.fetchID

		return func() tea.Msg {
			return message.FetchingFinished{
				Stories:  stories,
				Category: categoryToFetch,
				FetchID:  fetchID,
			}
		}
	}

	service := m.service
	numItems := m.numberOfItemsToFetch(categoryToFetch)
	endpoint := categories.Endpoint(categoryToFetch)
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		stories, err := fetchItems(ctx, service, numItems, endpoint)

		return message.FetchingFinished{
			Stories:  stories,
			Category: categoryToFetch,
			Err:      err,
			FetchID:  fetchID,
		}
	}
}

// fetchItems wraps service.FetchItems with the terminal progress lifecycle.
func fetchItems(ctx context.Context, service hn.Service, numItems int, endpoint string) ([]*hn.Story, error) {
	setProgressIndeterminate()

	stories, err := service.FetchItems(ctx, numItems, endpoint)

	switch {
	case errors.Is(err, context.Canceled):
		clearProgress()
	case err != nil:
		setProgressError()
	default:
		clearProgress()
	}

	return stories, err
}

func fetchMemorialStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		active, err := memorial.Detect(ctx)

		return message.MemorialStatusReady{Active: active, Err: err}
	}
}

func (m *model) numberOfItemsToFetch(cat categories.Category) int {
	if categories.Policy(cat) == categories.MultiPage {
		return m.list.PerPage() * m.config.PageMultiplier
	}

	return m.list.PerPage()
}

func newHistory(debugMode bool, doNotMarkAsRead bool) (history.History, error) {
	if debugMode {
		return history.NewMockHistory(), nil
	}

	if doNotMarkAsRead {
		return history.NewNonPersistentHistory(), nil
	}

	return history.NewPersistentHistory()
}

func (m *model) fetchCategory(cat categories.Category, index, cursor int) tea.Cmd {
	service := m.service
	numItems := m.numberOfItemsToFetch(cat)
	endpoint := categories.Endpoint(cat)
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		stories, err := fetchItems(ctx, service, numItems, endpoint)

		return message.CategoryFetchingFinished{
			Stories:  stories,
			Category: cat,
			Index:    index,
			Cursor:   cursor,
			Err:      err,
			FetchID:  fetchID,
		}
	}
}

func (m *model) handleEnteringCommentSection(msg message.EnteringCommentSection) tea.Cmd {
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

		tree, err := service.FetchComments(ctx, msg.ID, onProgress)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				clearProgress()
			} else {
				setProgressError()
			}

			return message.CommentTreeDataReady{
				Err:     err,
				FetchID: fetchID,
			}
		}

		clearProgress()

		histErr := hist.MarkRead(msg.ID, msg.CommentCount)

		var updatedStory *hn.Story
		if isOnFavorites {
			updatedStory = &hn.Story{
				ID:            tree.ID,
				Title:         tree.Title,
				Points:        tree.Points,
				Author:        tree.Author,
				Time:          tree.Time,
				URL:           tree.URL,
				Domain:        tree.Domain,
				CommentsCount: tree.CommentsCount,
			}
		}

		return message.CommentTreeDataReady{
			Thread:         comment.ToThread(tree),
			LastVisited:    lastVisited,
			UpdatedStory:   updatedStory,
			FetchID:        fetchID,
			HistoryWarning: histErr,
		}
	}
}

func (m *model) handleEnteringReaderMode(msg message.EnteringReaderMode) tea.Cmd {
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
			if errors.Is(err, context.Canceled) {
				clearProgress()
			} else {
				setProgressError()
			}

			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		clearProgress()

		histErr := hist.MarkArticleRead(msg.ID)

		return message.ArticleReady{
			Parsed:         parsed,
			Title:          msg.Title,
			URL:            msg.URL,
			Author:         msg.Author,
			TimeAgo:        msg.TimeAgo,
			ID:             msg.ID,
			Points:         msg.Points,
			FetchID:        fetchID,
			HistoryWarning: histErr,
		}
	}
}

// Terminal progress bar via OSC 9;4 (supported by Ghostty, ConEmu and others;
// silently ignored by terminals that don't recognise the sequence).
// Writes to stderr to avoid interfering with Bubble Tea's stdout.

func setProgressIndeterminate() { fmt.Fprint(os.Stderr, "\033]9;4;3;0\a") }
func setProgressError()         { fmt.Fprint(os.Stderr, "\033]9;4;2;100\a") }

func clearProgress() { fmt.Fprint(os.Stderr, "\033]9;4;0\a") }

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

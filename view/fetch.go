package view

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/memorial"
	"github.com/bensadeh/circumflex/timeago"
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
		stories, err := service.FetchItems(ctx, numItems, endpoint)

		return message.FetchingFinished{
			Stories:  stories,
			Category: categoryToFetch,
			Err:      err,
			FetchID:  fetchID,
		}
	}
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
		stories, err := service.FetchItems(ctx, numItems, endpoint)

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

func (m *model) fetchComments(story *hn.Story) tea.Cmd {
	isOnFavorites := m.cat.CurrentCategory() == categories.Favorites
	hist := m.history
	service := m.service
	ctx := m.fetchCtx
	fetchID := m.fetchID

	return func() tea.Msg {
		lastVisited := hist.CommentsLastVisited(story.ID)

		// Percentage updates are the one progress write left outside the
		// Update loop; the ctx guard stops a canceled fetch from writing
		// over its successor's indicator.
		onProgress := func(fetched, total int) {
			if total <= 0 || ctx.Err() != nil {
				return
			}

			setProgressPercent(min(fetched*100/total, 100))
		}

		tree, err := service.FetchComments(ctx, story.ID, onProgress)
		if err != nil {
			return message.CommentTreeDataReady{
				Err:     err,
				FetchID: fetchID,
			}
		}

		histErr := hist.MarkRead(story.ID, story.CommentsCount)

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

func (m *model) fetchArticle(story *hn.Story) tea.Cmd {
	hist := m.history
	ctx := m.fetchCtx
	fetchID := m.fetchID
	timeAgo := timeago.RelativeTime(story.Time)

	return func() tea.Msg {
		if err := article.Validate(story.Title, story.Domain); err != nil {
			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		parsed, err := article.Parse(ctx, story.URL)
		if err != nil {
			return message.ArticleReady{Err: err, FetchID: fetchID}
		}

		histErr := hist.MarkArticleRead(story.ID)

		return message.ArticleReady{
			Parsed:         parsed,
			Title:          story.Title,
			URL:            story.URL,
			Author:         story.Author,
			TimeAgo:        timeAgo,
			ID:             story.ID,
			Points:         story.Points,
			FetchID:        fetchID,
			HistoryWarning: histErr,
		}
	}
}

// Terminal progress bar via OSC 9;4 (supported by Ghostty, ConEmu and others;
// silently ignored by terminals that don't recognise the sequence).
//
// While the program runs, sequences ride progressCh into its message loop and
// leave through its own output, serialized with frame flushes. Writing to the
// terminal directly would race them: Bubble Tea flushes frames from its own
// goroutine, backpressure splits a frame across several writes, and a
// sequence landing between two chunks corrupts the terminal's parse — the
// cell-diff renderer then leaves ghost text it believes was repainted.
// progressOut serves tests and the final clear after the program exits.

var progressOut io.Writer = os.Stderr

// progressCh is wired by Run; nil whenever no program is running.
var progressCh chan<- string

const progressClearSeq = "\033]9;4;0\a"

func setProgressIndeterminate()  { emitProgress("\033]9;4;3;0\a") }
func setProgressPercent(pct int) { emitProgress(fmt.Sprintf("\033]9;4;1;%d\a", pct)) }
func setProgressError()          { emitProgress("\033]9;4;2;100\a") }

func clearProgress() { emitProgress(progressClearSeq) }

func emitProgress(seq string) {
	if progressCh != nil {
		// Progress is cosmetic: if the program stopped consuming, drop the
		// update rather than block.
		select {
		case progressCh <- seq:
		default:
		}

		return
	}

	_, _ = fmt.Fprint(progressOut, seq)
}

// syncProgress settles the indicator for a finished fetch: an error stays
// visible for the status message lifetime (see showDetailError), success
// clears it. Called only from the Update loop after the fetchID guard, so a
// stale fetch can never write over its successor's indicator.
func syncProgress(err error) {
	if err != nil {
		setProgressError()

		return
	}

	clearProgress()
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

	first, size := utf8.DecodeRuneInString(errStr)
	msg := string(unicode.ToUpper(first)) + errStr[size:]

	if before, after, ok := strings.Cut(msg, "status "); ok {
		msg = before + "status " + redText.Render(after)
	}

	return msg
}

// showDetailError surfaces a failed story load. In the wide layout the error
// replaces whatever the pane was showing as a view of its own: J/K page on to
// the neighboring stories in the target view — the one the failed load was
// for — and quit returns to the front page. The narrow layout keeps the
// previous view on screen and surfaces the error on the status bar instead.
// Either way the terminal progress indicator settles after the usual status
// message lifetime: the narrow layout via the message expiring, the wide via
// the returned timeout.
func (m *model) showDetailError(err error, target screen) tea.Cmd {
	if m.isWide() {
		// The placeholder renders from target, not m.detailTarget: validation
		// errors arrive without a fetch, and the view outlives the fetch state.
		metaBlock := func(paneWidth int) string { return m.placeholderMetaBlock(paneWidth, target) }
		m.detail = newErrorView(friendlyError(err), m.list.SelectedItem().Title, m.config.EnableNerdFonts, metaBlock, m.detailWidth(), m.height)
		m.screen = target

		fetchID := m.fetchID

		return tea.Tick(statusMessageLong, func(time.Time) tea.Msg {
			return message.ErrorProgressTimeout{FetchID: fetchID}
		})
	}

	return m.status.NewStatusMessageWithDuration(friendlyError(err), statusMessageLong)
}

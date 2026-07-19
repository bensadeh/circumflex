package view

import (
	"context"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/memorial"
	"github.com/bensadeh/circumflex/timeago"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"

	tea "charm.land/bubbletea/v2"
)

func (m *model) fetchStoriesForFirstCategory(tok fetchToken) tea.Cmd {
	categoryToFetch := m.cat.CurrentCategory()

	// Favorites is served locally — it is never fetched over the network. Hand
	// the already-synced items straight to the normal "fetch finished" path.
	if categories.IsFavorites(categoryToFetch) {
		stories := m.list.Items(categoryToFetch)
		index := m.cat.CurrentIndex()

		return func() tea.Msg {
			return message.StoriesReady{
				Stories:  stories,
				Category: categoryToFetch,
				Index:    index,
				FetchID:  tok.id,
			}
		}
	}

	return m.fetchCategory(tok, categoryToFetch, m.cat.CurrentIndex(), m.list.Cursor())
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

func (m *model) fetchCategory(tok fetchToken, cat categories.Category, index, cursor int) tea.Cmd {
	service := m.service
	numItems := m.numberOfItemsToFetch(cat)
	endpoint := categories.Endpoint(cat)

	return func() tea.Msg {
		stories, err := service.FetchItems(tok.ctx, numItems, endpoint)

		return message.StoriesReady{
			Stories:  stories,
			Category: cat,
			Index:    index,
			Cursor:   cursor,
			Err:      err,
			FetchID:  tok.id,
		}
	}
}

// fetchSearch mirrors fetchCategory for the search tab: the matching stories
// arrive on the same StoriesReady path, with the cursor reset to the top.
func (m *model) fetchSearch(tok fetchToken, query string, index int) tea.Cmd {
	service := m.service
	req := m.searchFilters.request(query, m.numberOfItemsToFetch(categories.Search))

	return func() tea.Msg {
		stories, err := service.SearchItems(tok.ctx, req)

		return message.StoriesReady{
			Stories:  stories,
			Category: categories.Search,
			Index:    index,
			Err:      err,
			FetchID:  tok.id,
		}
	}
}

func (m *model) fetchComments(tok fetchToken, story *hn.Story) tea.Cmd {
	isOnFavorites := m.cat.CurrentCategory() == categories.Favorites
	hist := m.history
	service := m.service

	return func() tea.Msg {
		lastVisited := hist.CommentsLastVisited(story.ID)

		// Percentage updates are the one progress write left outside the
		// Update loop; the ctx guard stops a canceled fetch from writing
		// over its successor's indicator.
		onProgress := func(fetched, total int) {
			if total <= 0 || tok.ctx.Err() != nil {
				return
			}

			pane.SetProgressPercent(min(fetched*100/total, 100))
		}

		tree, err := service.FetchComments(tok.ctx, story.ID, onProgress)
		if err != nil {
			return message.CommentTreeDataReady{
				Err:     err,
				FetchID: tok.id,
			}
		}

		histErr := hist.MarkRead(story.ID, story.CommentsCount)

		var updatedStory *hn.Story

		if isOnFavorites {
			story := tree.Story
			updatedStory = &story
		}

		return message.CommentTreeDataReady{
			Thread:         comment.ToThread(tree),
			LastVisited:    lastVisited,
			UpdatedStory:   updatedStory,
			FetchID:        tok.id,
			HistoryWarning: histErr,
		}
	}
}

// fetchLinkedComments is fetchComments for a Hacker News discussion link
// followed inside an article: the thread arrives carrying the walk-back
// trail instead of story-list context. The service resolves a comment link
// to the story rooting it, so history reads and marks the resolved ID.
func (m *model) fetchLinkedComments(tok fetchToken, id int, trail []message.TrailEntry) tea.Cmd {
	hist := m.history
	service := m.service

	return func() tea.Msg {
		onProgress := func(fetched, total int) {
			if total <= 0 || tok.ctx.Err() != nil {
				return
			}

			pane.SetProgressPercent(min(fetched*100/total, 100))
		}

		tree, err := service.FetchComments(tok.ctx, id, onProgress)
		if err != nil {
			return message.LinkCommentsReady{
				Err:     err,
				FetchID: tok.id,
			}
		}

		lastVisited := hist.CommentsLastVisited(tree.ID)
		histErr := hist.MarkRead(tree.ID, tree.CommentsCount)

		return message.LinkCommentsReady{
			Thread:         comment.ToThread(tree),
			LastVisited:    lastVisited,
			Trail:          trail,
			FetchID:        tok.id,
			HistoryWarning: histErr,
		}
	}
}

func (m *model) fetchArticle(tok fetchToken, story *hn.Story) tea.Cmd {
	hist := m.history
	timeAgo := timeago.RelativeTime(story.Time)

	return func() tea.Msg {
		if err := article.Validate(story.Title, story.Domain); err != nil {
			return message.ArticleReady{Err: err, FetchID: tok.id}
		}

		parsed, err := article.Parse(tok.ctx, story.URL)
		if err != nil {
			return message.ArticleReady{Err: err, FetchID: tok.id}
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
			CommentsCount:  story.CommentsCount,
			FetchID:        tok.id,
			HistoryWarning: histErr,
		}
	}
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
		// The placeholder renders from target, not m.fetch.target: validation
		// errors arrive without a fetch, and the view outlives the fetch state.
		metaBlock := func(paneWidth int) string { return m.placeholderMetaBlock(paneWidth, target) }
		m.detail = newErrorView(pane.FriendlyError(err), m.list.SelectedItem().Title, m.config.EnableNerdFonts, metaBlock, m.detailWidth(), m.height)
		m.screen = target

		fetchID := m.fetch.currentID()

		return tea.Tick(statusMessageLong, func(time.Time) tea.Msg {
			return message.ErrorProgressTimeout{FetchID: fetchID}
		})
	}

	return m.status.NewStatusMessageWithDuration(pane.FriendlyError(err), statusMessageLong)
}

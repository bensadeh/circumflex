package pane

import (
	"context"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
)

// ReaderFetchTimeout bounds fetching and parsing a page for reader mode — a
// story's article or a followed link — in both shells. One constant, so the
// two cannot drift apart.
const ReaderFetchTimeout = 15 * time.Second

// FetchPage loads a page reached by following a link inside an article, for
// whichever shell dispatched it. Unlike a story fetch there is no story: no
// title to validate against, and nothing marked read. trail rides along
// untouched — it becomes the new page's walk-back chain.
func FetchPage(ctx context.Context, fetchID uint64, url string, trail []message.TrailEntry) tea.Cmd {
	return func() tea.Msg {
		parsed, err := article.Parse(ctx, url)
		if err != nil {
			return message.LinkArticleReady{Err: err, FetchID: fetchID}
		}

		return message.LinkArticleReady{
			Parsed:  parsed,
			Title:   parsed.Title,
			URL:     url,
			Trail:   trail,
			FetchID: fetchID,
		}
	}
}

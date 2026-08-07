package pane

import (
	"context"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/graphics"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
)

// ReaderFetchTimeout bounds fetching and parsing a page for reader mode — a
// story's article or a followed link — in both shells. One constant, so the
// two cannot drift apart.
const ReaderFetchTimeout = 15 * time.Second

// ThreadFetcher loads a Hacker News discussion as the comment views consume
// it. The cmd layer supplies one wrapping its service, so this package stays
// ignorant of which backend answers.
type ThreadFetcher func(ctx context.Context, id int, onProgress func(fetched, total int)) (*comment.Thread, error)

// FetchThread loads a Hacker News discussion linked inside an article, for
// the standalone shells. Nothing is marked read — the one-shot commands
// leave no history — so the thread arrives as if fully visited, exactly as
// the comments subcommand serves its own.
func FetchThread(ctx context.Context, fetchID uint64, fetch ThreadFetcher, id int, trail []message.TrailEntry) tea.Cmd {
	return func() tea.Msg {
		onProgress := func(fetched, total int) {
			if total <= 0 || ctx.Err() != nil {
				return
			}

			SetProgressPercent(min(fetched*100/total, 100))
		}

		thread, err := fetch(ctx, id, onProgress)
		if err != nil {
			return message.LinkCommentsReady{Err: err, FetchID: fetchID}
		}

		return message.LinkCommentsReady{
			Thread:      thread,
			LastVisited: time.Now().Unix(),
			Trail:       trail,
			FetchID:     fetchID,
		}
	}
}

// FetchPage loads a page reached by following a link inside an article, for
// whichever shell dispatched it. Unlike a story fetch there is no story: no
// title to validate against, and nothing marked read. trail rides along
// untouched — it becomes the new page's walk-back chain.
func FetchPage(ctx context.Context, fetchID uint64, url string, trail []message.TrailEntry) tea.Cmd {
	return func() tea.Msg {
		parsed, err := article.Parse(ctx, url, graphics.Enabled())
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

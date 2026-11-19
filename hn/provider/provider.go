package provider

import (
	"context"

	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/services/algolia"
	"github.com/bensadeh/circumflex/hn/services/firebase"
	"github.com/bensadeh/circumflex/hn/services/mock"
	"github.com/bensadeh/circumflex/hn/services/website"
)

// live composes the production backends: Firebase serves the feeds, items
// and comments; Algolia serves search; the Hacker News site itself serves
// the active feed, which the API does not expose.
type live struct {
	feeds  *firebase.Service
	search *algolia.Service
	site   *website.Service
}

func (l live) FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*hn.Story, error) {
	return l.feeds.FetchItems(ctx, itemsToFetch, category)
}

// FetchActiveItems reads the story IDs off the Active Threads page and
// hydrates them through Firebase, so the active category is served by the
// same item pipeline as the feeds.
func (l live) FetchActiveItems(ctx context.Context, itemsToFetch int) ([]*hn.Story, error) {
	ids, err := l.site.FetchActiveStoryIDs(ctx)
	if err != nil {
		return nil, err
	}

	return l.feeds.FetchItemsByID(ctx, ids[:min(len(ids), itemsToFetch)])
}

func (l live) FetchItem(ctx context.Context, id int) (*hn.Story, error) {
	return l.feeds.FetchItem(ctx, id)
}

func (l live) FetchComments(ctx context.Context, id int, onProgress func(fetched, total int)) (*hn.CommentTree, error) {
	return l.feeds.FetchComments(ctx, id, onProgress)
}

func (l live) SearchItems(ctx context.Context, req hn.SearchRequest) ([]*hn.Story, error) {
	return l.search.SearchItems(ctx, req)
}

func NewService(debugMode, debugFallible bool) hn.Service {
	if debugFallible {
		return mock.NewFallibleService()
	}

	if debugMode {
		return mock.Service{}
	}

	return live{feeds: firebase.NewService(), search: algolia.NewService(), site: website.NewService()}
}

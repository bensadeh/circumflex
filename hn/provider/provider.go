package provider

import (
	"context"

	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/services/algolia"
	"github.com/bensadeh/circumflex/hn/services/firebase"
	"github.com/bensadeh/circumflex/hn/services/mock"
)

// live composes the production backends: Firebase serves the feeds, items
// and comments; Algolia serves search.
type live struct {
	feeds  *firebase.Service
	search *algolia.Service
}

func (l live) FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*hn.Story, error) {
	return l.feeds.FetchItems(ctx, itemsToFetch, category)
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

	return live{feeds: firebase.NewService(), search: algolia.NewService()}
}

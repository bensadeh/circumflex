package hn

import (
	"context"

	"github.com/bensadeh/circumflex/hn/services/firebase"
	"github.com/bensadeh/circumflex/hn/services/mock"
	"github.com/bensadeh/circumflex/item"
)

type Service interface {
	FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*item.Story, error)
	FetchItem(ctx context.Context, id int) (*item.Story, error)
	FetchComments(ctx context.Context, id int, onProgress func(fetched, total int)) (*item.Story, error)
}

func NewService(debugMode, debugFallible bool) Service {
	if debugFallible {
		return mock.NewFallibleService()
	}

	if debugMode {
		return mock.Service{}
	}

	return firebase.NewService()
}

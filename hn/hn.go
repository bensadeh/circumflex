package hn

import (
	"clx/hn/services/firebase"
	"clx/hn/services/mock"
	"clx/item"
	"context"
)

type Service interface {
	FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*item.Story, error)
	FetchItem(ctx context.Context, id int) (*item.Story, error)
	FetchComments(ctx context.Context, id int) (*item.Story, error)
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

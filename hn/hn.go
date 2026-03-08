package hn

import (
	"clx/hn/services/hybrid"
	"clx/hn/services/mock"
	"clx/item"
)

type Service interface {
	FetchItems(itemsToFetch int, category int) ([]*item.Story, error)
	FetchItem(id int) (*item.Story, error)
	FetchComments(id int) (*item.Story, error)
}

func NewService(debugMode bool) Service {
	if debugMode {
		return mock.Service{}
	}

	return hybrid.NewService()
}

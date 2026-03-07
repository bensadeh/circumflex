package hn

import "clx/item"

type Service interface {
	FetchItems(itemsToFetch int, category int) ([]*item.Story, error)
	FetchItem(id int) (*item.Story, error)
	FetchComments(id int) (*item.Story, error)
}

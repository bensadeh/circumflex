package hn

import "clx/item"

type Service interface {
	FetchItems(itemsToFetch int, category int) ([]*item.Item, error)
	FetchItem(id int) (*item.Item, error)
	FetchComments(id int) (*item.Item, error)
}

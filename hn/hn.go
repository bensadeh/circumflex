package hn

import "clx/item"

type Service interface {
	FetchItems(itemsToFetch int, category int) (items []*item.Item, errMsg string)
	FetchItem(id int) (*item.Item, error)
	FetchComments(int) (*item.Item, error)
}

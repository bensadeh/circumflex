package hn

import "github.com/f01c33/circumflex/item"

type Service interface {
	FetchItems(itemsToFetch int, category int) (items []*item.Item, errMsg string)
	FetchItem(id int) *item.Item
	FetchComments(int) *item.Item
}

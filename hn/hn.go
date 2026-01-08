package hn

import "github.com/bensadeh/circumflex/item"

type Service interface {
	FetchItems(itemsToFetch int, category int) (items []*item.Item, errMsg string)
	FetchItem(id int) *item.Item
	FetchComments(int) *item.Item
}

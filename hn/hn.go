package hn

import "clx/item"

type Service interface {
	Init(itemsToShow int)
	FetchStories(int, int) []*item.Item
	FetchStory(int) *item.Item
}

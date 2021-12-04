package hn

import "clx/item"

type Service interface {
	Init()
	FetchStories(int, int) []*item.Item
	FetchStory(int) *item.Item
}

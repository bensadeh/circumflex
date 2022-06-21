package hn

import "clx/item"

type Service interface {
	FetchStories(int, int) []*item.Item
	FetchStory(int) *item.Item
}

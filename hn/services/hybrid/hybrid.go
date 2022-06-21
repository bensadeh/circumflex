package hybrid

import (
	algolia_bubble "clx/hn/services/algolia-bubble"
	"clx/hn/services/cheeaun"
	"clx/item"
)

type Service struct {
	algo *algolia_bubble.Service
}

func (s *Service) Init(itemsToShow int) {
	s.algo = new(algolia_bubble.Service)

	s.algo.Init(itemsToShow)
}

func (s *Service) FetchStories(page int, category int) []*item.Item {
	return s.algo.FetchStories(page, category)
}

func (s *Service) FetchStory(id int) *item.Item {
	c := cheeaun.Service{}

	return c.FetchStory(id)
}

package hybrid

import (
	"clx/hn/services/algolia"
	"clx/hn/services/cheeaun"
	"clx/item"
)

type Service struct {
	algo *algolia.Service
}

func (s *Service) Init(itemsToShow int) {
	s.algo = new(algolia.Service)

	s.algo.Init(itemsToShow)
}

func (s *Service) FetchStories(page int, category int) []*item.Item {
	return s.algo.FetchStories(page, category)
}

func (s *Service) FetchStory(id int) *item.Item {
	c := cheeaun.Service{}

	return c.FetchStory(id)
}

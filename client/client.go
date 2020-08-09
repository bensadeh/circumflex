package client

import (
	"fmt"

	"circumflex/client/feed"
	"circumflex/client/scrape"
)

//HNClient for interacting with HackerNews
type HNClient struct {
	scraper Scraper
}

//Scraper defines interface for scraping hackernews
type Scraper interface {
	ScrapeHNFeed(numStories int) (*[]feed.Item, error)
}

const baseURL = "https://news.ycombinator.com/"

//NewHNClient instaniates a new HackerNews Client
func NewHNClient() HNClient {
	return HNClient{
		scraper: scrape.NewScraper(baseURL),
	}
}

//GetTopStories retrieves data from a source given a target amount of max stories
// This currently has to be greater than zero set and less than 100
// The client currently uses a scraper interface but could be changed for the HN API in the future
func (hnc *HNClient) GetTopStories(targetAmount int) (*[]feed.Item, error) {
	if targetAmount <= 0 || targetAmount > 100 {
		return nil, fmt.Errorf("num posts amount cannot be equal to zero or greater than 100, got: %d ",
			targetAmount)
	}
	return hnc.scraper.ScrapeHNFeed(targetAmount)

}

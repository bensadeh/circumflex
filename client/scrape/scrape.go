package scrape

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"circumflex/client/feed"
	"github.com/gocolly/colly"
)

//HNScraper is a scraper for hackernews, using colly as the collector
type HNScraper struct {
	c       *colly.Collector
	baseURL string
}

//NewScraper creates a new instance of the scraper
func NewScraper(baseURL string) HNScraper {
	return HNScraper{c: colly.NewCollector(), baseURL: baseURL}
}

//ScrapeHNFeed scrapes the main page until it hit max items or there is no more data
func (s HNScraper) ScrapeHNFeed(maxItems int) (*[]feed.Item, error) {

	var allItems []feed.Item
	pageNum := 1
	currentTotal := 0
	prevTotal := 0
	shouldBreak := false

	for {
		s.c.OnHTML("#hnmain .itemlist .athing", func(e *colly.HTMLElement) {

			if currentTotal == maxItems {
				return
			}
			Item, err := processFeedItem(e)
			if err != nil {
				// log error, or write to file
				fmt.Println(err)
			} else {
				allItems = append(allItems, *Item)
				currentTotal++

			}
		})

		s.c.OnScraped(func(*colly.Response) {
			if len(allItems) != maxItems {
				pageNum++
			}
			prevTotal = currentTotal

		})

		s.c.Visit(fmt.Sprintf("%s/news?p=%d", s.baseURL, pageNum))
		if shouldBreak {
			break

		} else if currentTotal == prevTotal {
			shouldBreak = true
		}
	}
	if len(allItems) == 0 {
		return nil, errors.New("No feed items found")
	}

	return &allItems, nil

}

//processFeedItem extracts the current html element and tries to create a feed Item
func processFeedItem(e *colly.HTMLElement) (*feed.Item, error) {
	rank := e.ChildText(".rank")
	rankSplit := strings.Split(rank, ".")
	if len(rankSplit) == 0 {
		return nil, fmt.Errorf("rank expected to have . got %s", rank)
	}
	rankI, err := strconv.Atoi(rankSplit[0])
	if err != nil {
		return nil, fmt.Errorf("rank expected to be integer, got %s", rank)
	}

	title := e.ChildText(".storylink")
	link := e.ChildAttr(".title a", "href")

	metaDataRow := e.DOM.Next()
	if metaDataRow == nil {
		return nil, fmt.Errorf("expected to have a metadata row, got none")
	}
	score := strings.TrimSpace(strings.Replace(metaDataRow.Find(".score").Text(), "points", "", -1))
	scoreI, err := strconv.Atoi(score)
	if err != nil {
		scoreI = 0
	}
	author := metaDataRow.Find(".hnuser").Text()
	var comments string
	metaDataRow.Find("a").EachWithBreak((func(i int, s *goquery.Selection) bool {
		if strings.Contains(s.Text(), "comments") {
			comments = strings.TrimSpace(strings.Replace(s.Text(), "comments", "", -1))
			return false
		}
		return true
	}))
	commentsI, err := strconv.Atoi(comments)
	if err != nil {
		commentsI = 0
	}
	feedItem, err := feed.NewItem(title, link, author, scoreI, commentsI, rankI)
	return &feedItem, err
}

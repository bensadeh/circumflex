package http

import (
	"clx/constants/categories"
	"clx/constants/clx"
	"clx/core"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL = "http://api.hackerwebapp.com/"
	page    = "?page="
)

func FetchStories(page int, category int) ([]*core.Story, error) {
	url := getURL(category)
	p := strconv.Itoa(page)

	var s []*core.Story

	client := resty.New()
	client.SetTimeout(5 * time.Second)

	_, err := client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(&s).
		Get(url + p)
	if err != nil {
		return nil, fmt.Errorf("could not fetch stories: %w", err)
	}

	return s, nil
}

func getURL(category int) string {
	switch category {
	case categories.FrontPage:
		return baseURL + "news" + page
	case categories.New:
		return baseURL + "newest" + page
	case categories.Ask:
		return baseURL + "ask" + page
	case categories.Show:
		return baseURL + "show" + page
	}

	return ""
}

package http

import (
	"clx/constants/category"
	"clx/constants/clx"
	"clx/endpoints"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL = "http://api.hackerwebapp.com/"
	page    = "?page="
)

func FetchStories(page int, category int) ([]*endpoints.Story, error) {
	url := getURL(category)
	p := strconv.Itoa(page)

	var s []*endpoints.Story

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

func getURL(cat int) string {
	switch cat {
	case category.FrontPage:
		return baseURL + "news" + page

	case category.New:
		return baseURL + "newest" + page

	case category.Ask:
		return baseURL + "ask" + page

	case category.Show:
		return baseURL + "show" + page

	default:
		return ""
	}
}

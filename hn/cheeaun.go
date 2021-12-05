package hn

import (
	"clx/constants/clx"
	"clx/endpoints"
	"clx/item"
	"clx/utils/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type Cheeaun struct{}

func (r Cheeaun) Init() {
}

func (r Cheeaun) FetchStories(page int, category int) []*item.Item {
	stories, _ := http.FetchStories(page, category)

	return mapStories(stories)
}

func mapStories(stories []*endpoints.Story) []*item.Item {
	items := make([]*item.Item, 0, len(stories))

	for _, story := range stories {
		item := item.Item{
			ID:            story.ID,
			Title:         story.Title,
			Points:        story.Points,
			User:          story.Author,
			Time:          story.Time,
			TimeAgo:       "",
			Type:          story.Type,
			URL:           story.URL,
			Level:         0,
			Domain:        story.Domain,
			Comments:      nil,
			Content:       "",
			CommentsCount: story.CommentsCount,
		}

		items = append(items, &item)
	}

	return items
}

func (r Cheeaun) FetchStory(id int) *item.Item {
	comments := new(endpoints.Comments)

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetHostURL("http://api.hackerwebapp.com/item/")

	_, _ = client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(comments).
		Get(strconv.Itoa(id))
	//if err != nil {
	//	return nil, fmt.Errorf("could not fetch comments: %w", err)
	//}

	return mapComments(comments)
}

func mapComments(comments *endpoints.Comments) *item.Item {
	items := make([]*item.Item, 0, len(comments.Comments))

	for i := range comments.Comments {
		items = append(items, mapComments(&comments.Comments[i]))
	}

	return &item.Item{
		ID:            comments.ID,
		Title:         comments.Title,
		Points:        comments.Points,
		User:          comments.User,
		Time:          comments.Time,
		TimeAgo:       comments.TimeAgo,
		Type:          comments.Type,
		URL:           comments.URL,
		Level:         comments.Level,
		Domain:        comments.Domain,
		Comments:      items,
		Content:       comments.Content,
		CommentsCount: comments.CommentsCount,
	}
}

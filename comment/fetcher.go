package comment

import (
	"clx/constants/clx"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type Comments struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	Points        int        `json:"points"`
	User          string     `json:"user"`
	Time          int        `json:"time"`
	TimeAgo       string     `json:"time_ago"`
	Type          string     `json:"type"`
	URL           string     `json:"url"`
	Level         int        `json:"level"`
	Domain        string     `json:"domain"`
	Comments      []Comments `json:"comments"`
	Content       string     `json:"content"`
	CommentsCount int        `json:"comments_count"`
}

func FetchComments(id string) (*Comments, error) {
	comments := new(Comments)

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetHostURL("http://api.hackerwebapp.com/item/")

	_, err := client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(comments).
		Get(id)
	if err != nil {
		return nil, fmt.Errorf("could not fetch comments: %w", err)
	}

	return comments, nil
}

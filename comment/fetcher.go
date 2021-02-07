package comment

import (
	"clx/constants/clx"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type Comments struct {
	Author        string      `json:"user"`
	Title         string      `json:"title"`
	Comment       string      `json:"content"`
	CommentsCount int         `json:"comments_count"`
	Time          string      `json:"time_ago"`
	Points        int         `json:"points"`
	URL           string      `json:"url"`
	Domain        string      `json:"domain"`
	Level         int         `json:"level"`
	ID            int         `json:"id"`
	Replies       []*Comments `json:"comments"`
}

func FetchComments(id string) (*Comments, error) {
	stations := new(Comments)

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetHostURL("http://api.hackerwebapp.com/item/")

	_, err := client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(stations).
		Get(id)
	if err != nil {
		return nil, fmt.Errorf("could not fetch comments: %w", err)
	}

	return stations, nil
}

package comment

import (
	"clx/constants/clx"
	"clx/endpoints"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

func FetchComments(id string) (*endpoints.Comments, error) {
	comments := new(endpoints.Comments)

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

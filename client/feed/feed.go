package feed

import (
	"errors"
	"fmt"
	"net/url"
)

//Item holds reference to a story on the feed
type Item struct {
	Title    string `json:"title"`
	URI      string `json:"uri"`
	Author   string `json:"author"`
	Points   int    `json:"points"`
	Comments int    `json:"comments"`
	Rank     int    `json:"rank"`
}

// NewItem to create a new feed item, but if it does not match the following criteria then it is not valid:
// title and author are non empty strings not longer than 256 characters
// uri is a valid URI
//points, comments and rank are integers >= 0.
func NewItem(title, link, author string, score, comments, rank int) (Item, error) {
	if author != "" {
		if len(author) > 256 {
			author = author[:256]
		}
	} else {
		return Item{}, errors.New("author cannot be empty")
	}

	if title != "" {
		if len(title) > 256 {
			title = title[:256]
		}
	} else {
		return Item{}, errors.New("title cannot be empty")
	}

	if score < 0 {
		return Item{}, fmt.Errorf("score is not greater or equal to zero, got %d", score)
	}
	if rank < 0 {
		return Item{}, fmt.Errorf("rank is not greater or equal to zero, got %d", rank)
	}
	if comments < 0 {
		return Item{}, fmt.Errorf("comments is not greater or equal to zero, got %d", comments)
	}

	_, err := url.ParseRequestURI(link)
	if err != nil {
		return Item{}, fmt.Errorf("feed item does not contain a valid URI, got %s", link)
	}

	return Item{title, link, author, score, comments, rank}, nil

}

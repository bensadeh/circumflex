package algolia

import (
	"clx/constants/clx"
	"clx/item"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type algolia struct {
	Hits []struct {
		CreatedAt       time.Time   `json:"created_at"`
		Title           string      `json:"title"`
		URL             string      `json:"url"`
		Author          string      `json:"author"`
		Points          int         `json:"points"`
		StoryText       interface{} `json:"story_text"`
		CommentText     interface{} `json:"comment_text"`
		NumComments     int         `json:"num_comments"`
		StoryID         interface{} `json:"story_id"`
		StoryTitle      interface{} `json:"story_title"`
		StoryURL        interface{} `json:"story_url"`
		ParentID        interface{} `json:"parent_id"`
		CreatedAtI      int         `json:"created_at_i"`
		Tags            []string    `json:"_tags"`
		ObjectID        string      `json:"objectID"`
		HighlightResult struct {
			Title struct {
				Value        string        `json:"value"`
				MatchLevel   string        `json:"matchLevel"`
				MatchedWords []interface{} `json:"matchedWords"`
			} `json:"title"`
			URL struct {
				Value        string        `json:"value"`
				MatchLevel   string        `json:"matchLevel"`
				MatchedWords []interface{} `json:"matchedWords"`
			} `json:"url"`
			Author struct {
				Value        string        `json:"value"`
				MatchLevel   string        `json:"matchLevel"`
				MatchedWords []interface{} `json:"matchedWords"`
			} `json:"author"`
		} `json:"_highlightResult"`
	} `json:"hits"`
	NbHits           int      `json:"nbHits"`
	Page             int      `json:"page"`
	NbPages          int      `json:"nbPages"`
	HitsPerPage      int      `json:"hitsPerPage"`
	ExhaustiveNbHits bool     `json:"exhaustiveNbHits"`
	ExhaustiveTypo   bool     `json:"exhaustiveTypo"`
	Query            string   `json:"query"`
	Params           string   `json:"params"`
	RenderingContent struct{} `json:"renderingContent"`
	ProcessingTimeMS int      `json:"processingTimeMS"`
}

type Service struct{}

func (s Service) Init() {
}

const (
	baseURL = "https://hn.algolia.com/api/v1/search?tags=front_page"
	page    = "?page="
)

func (s Service) FetchStories(_ int, category int) []*item.Item {
	// url := getURL(category)
	// p := strconv.Itoa(page)

	var a *algolia

	client := resty.New()
	client.SetTimeout(5 * time.Second)

	_, _ = client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(&a).
		Get("https://hn.algolia.com/api/v1/search?tags=front_page&hitsPerPage=30")

	return mapStories(a)
}

func mapStories(stories *algolia) []*item.Item {
	items := make([]*item.Item, 0, len(stories.Hits))

	for _, story := range stories.Hits {
		id, _ := strconv.Atoi(story.ObjectID)

		item := item.Item{
			ID:            id,
			Title:         story.Title,
			Points:        story.Points,
			User:          story.Author,
			Time:          int64(story.CreatedAtI),
			TimeAgo:       "",
			Type:          "",
			URL:           story.URL,
			Domain:        "",
			Comments:      nil,
			Content:       "",
			Level:         0,
			CommentsCount: story.NumComments,
		}

		items = append(items, &item)
	}

	return items
}

func (s Service) FetchStory(id int) *item.Item {
	return nil
}

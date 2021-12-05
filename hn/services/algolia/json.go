package algolia

import "time"

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

type comment struct {
	ID         int           `json:"id"`
	CreatedAt  time.Time     `json:"created_at"`
	CreatedAtI int           `json:"created_at_i"`
	Type       string        `json:"type"`
	Author     string        `json:"author"`
	Title      string        `json:"title"`
	URL        string        `json:"url"`
	Text       string        `json:"text"`
	Points     int           `json:"points"`
	ParentID   interface{}   `json:"parent_id"`
	StoryID    interface{}   `json:"story_id"`
	Children   []*comment    `json:"children"`
	Options    []interface{} `json:"options"`
}

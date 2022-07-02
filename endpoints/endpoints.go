package endpoints

import "time"

type Story struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Points        int    `json:"points"`
	Author        string `json:"user"`
	Time          int64  `json:"time"`
	CommentsCount int    `json:"comments_count"`
	URL           string `json:"url"`
	Domain        string `json:"domain"`
	Type          string `json:"type"`
}

type HN struct {
	By          string `json:"by"`
	Descendants int    `json:"descendants"`
	Id          int    `json:"id"`
	Kids        []int  `json:"kids"`
	Score       int    `json:"score"`
	Time        int    `json:"time"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Url         string `json:"url"`
}

type Comments struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	Points        int        `json:"points"`
	User          string     `json:"user"`
	Time          int64      `json:"time"`
	TimeAgo       string     `json:"time_ago"`
	Type          string     `json:"type"`
	URL           string     `json:"url"`
	Level         int        `json:"level"`
	Domain        string     `json:"domain"`
	Comments      []Comments `json:"comments"`
	Content       string     `json:"content"`
	CommentsCount int        `json:"comments_count"`
}

type Algolia struct {
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

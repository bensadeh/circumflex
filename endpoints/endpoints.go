package endpoints

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

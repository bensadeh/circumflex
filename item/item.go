package item

type Story struct {
	ID            int
	Title         string
	Points        int
	User          string
	Time          int64
	TimeAgo       string
	URL           string
	Domain        string
	Comments      []*Story
	Content       string
	CommentsCount int
}

package item

type Story struct {
	ID            int
	Title         string
	Points        int
	User          string
	Time          int64
	TimeAgo       string
	Type          string
	URL           string
	Level         int
	Domain        string
	Comments      []*Story
	Content       string
	CommentsCount int
}

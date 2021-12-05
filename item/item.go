package item

type Item struct {
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
	Comments      []*Item
	Content       string
	CommentsCount int
}

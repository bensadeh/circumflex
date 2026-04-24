package comment

type Comment struct {
	ID       int
	Author   string
	Content  string
	Time     int64
	TimeAgo  string
	Children []*Comment
}

type Thread struct {
	ID            int
	Title         string
	Author        string // OP
	URL           string
	Domain        string
	Points        int
	TimeAgo       string
	Content       string // self-text
	CommentsCount int
	Comments      []*Comment
}

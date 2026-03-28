package comment

// Comment represents a single comment in a discussion thread.
type Comment struct {
	ID       int
	Author   string
	Content  string
	Time     int64
	TimeAgo  string
	Depth    int
	Children []*Comment
}

// Thread represents a discussion thread with its metadata and comment tree.
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

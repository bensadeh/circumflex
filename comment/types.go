package comment

import "github.com/bensadeh/circumflex/hn"

type Comment struct {
	ID       int
	Author   string
	Content  string
	Time     int64
	Children []*Comment
}

// Thread is the comment section's view of a story: the story metadata
// (Author is the OP) plus self-text and the rendering-layer comment nodes.
type Thread struct {
	hn.Story

	Content  string // self-text
	Comments []*Comment
}

func NewCommentsCount(thread *Thread, lastVisited int64) int {
	count := 0

	for _, c := range thread.Comments {
		countNewComments(c, &count, lastVisited)
	}

	return count
}

func countNewComments(c *Comment, count *int, lastVisited int64) {
	if lastVisited < c.Time {
		*count++
	}

	for _, reply := range c.Children {
		countNewComments(reply, count, lastVisited)
	}
}

func FirstCommentID(comments []*Comment) int {
	if len(comments) == 0 {
		return 0
	}

	return comments[0].ID
}

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

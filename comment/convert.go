package comment

import (
	"github.com/bensadeh/circumflex/hn"
)

// ToThread maps an hn.CommentTree (API layer) to a Thread (rendering layer).
func ToThread(t *hn.CommentTree) *Thread {
	return &Thread{
		Story:    t.Story,
		Content:  t.Content,
		Comments: mapCommentNodes(t.Comments),
	}
}

func mapCommentNodes(nodes []*hn.CommentNode) []*Comment {
	if nodes == nil {
		return nil
	}

	result := make([]*Comment, 0, len(nodes))

	for _, n := range nodes {
		result = append(result, mapCommentNode(n))
	}

	return result
}

func mapCommentNode(n *hn.CommentNode) *Comment {
	return &Comment{
		ID:       n.ID,
		Author:   n.Author,
		Content:  n.Content,
		Time:     n.Time,
		Children: mapCommentNodes(n.Children),
	}
}

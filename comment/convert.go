package comment

import (
	"github.com/bensadeh/circumflex/hn"
)

// ToThread maps an hn.CommentTree (API layer) to a Thread (rendering layer),
// pruning removed comments as it copies. The prune is post-order, so a
// removed comment whose replies were all pruned goes with them; one with a
// surviving reply stays to anchor the thread. Pruning in this one funnel
// keeps first-comment detection, the new-comment count and rendering
// agreeing on a single rule.
func ToThread(t *hn.CommentTree) *Thread {
	return &Thread{
		Story:    t.Story,
		Content:  t.Content,
		Comments: mapCommentNodes(t.Comments),
	}
}

func mapCommentNodes(nodes []*hn.CommentNode) []*Comment {
	if len(nodes) == 0 {
		return nil
	}

	result := make([]*Comment, 0, len(nodes))

	for _, n := range nodes {
		if c := mapCommentNode(n); c != nil {
			result = append(result, c)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func mapCommentNode(n *hn.CommentNode) *Comment {
	children := mapCommentNodes(n.Children)

	if Removed(n.Content) && children == nil {
		return nil
	}

	return &Comment{
		ID:       n.ID,
		Author:   n.Author,
		Content:  n.Content,
		Time:     n.Time,
		Children: children,
	}
}

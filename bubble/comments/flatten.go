package comments

import (
	"clx/comment"
)

// FlatComment represents a single comment in the flattened view of the tree.
// It holds only tree-structural data — rendering artifacts like line positions
// are tracked separately in LineMetrics.
type FlatComment struct {
	Comment           *comment.Comment
	Depth             int
	Collapsed         bool
	ChildCount        int // total descendants
	GrandParentPoster string
}

// LineMetrics tracks the rendered position of a comment in the viewport.
// Indexed by flat index; recomputed on every render.
type LineMetrics struct {
	StartLine int
	LineCount int
}

// flatten performs a pre-order DFS of the comment tree and returns
// a flat slice. Each entry retains its depth and descendant count.
func flatten(thread *comment.Thread) []FlatComment {
	var result []FlatComment

	grandParentPoster := ""

	for _, child := range thread.Comments {
		flattenRecursive(child, 0, grandParentPoster, &result)
	}

	return result
}

func flattenRecursive(c *comment.Comment, depth int, grandParentPoster string, out *[]FlatComment) {
	if c.Content == "[deleted]" && len(c.Children) == 0 {
		return
	}

	childCount := comment.DescendantCount(c)

	fc := FlatComment{
		Comment:           c,
		Depth:             depth,
		Collapsed:         depth == 0 && childCount > 0,
		ChildCount:        childCount,
		GrandParentPoster: grandParentPoster,
	}
	*out = append(*out, fc)

	gp := grandParentPoster
	if depth == 0 {
		gp = c.Author
	}

	for _, reply := range c.Children {
		flattenRecursive(reply, depth+1, gp, out)
	}
}

// computeVisible returns indices into flat[] for comments that should be
// displayed, skipping children of collapsed nodes.
func computeVisible(flat []FlatComment) []int {
	visible := make([]int, 0, len(flat))

	skipUntilDepth := -1

	for i, fc := range flat {
		if skipUntilDepth >= 0 && fc.Depth > skipUntilDepth {
			continue
		}

		skipUntilDepth = -1

		visible = append(visible, i)

		if fc.Collapsed && fc.ChildCount > 0 {
			skipUntilDepth = fc.Depth
		}
	}

	return visible
}

package comments

import (
	"clx/comment"
	"clx/item"
)

// FlatComment represents a single comment in the flattened view of the tree.
type FlatComment struct {
	Story      *item.Story
	Depth      int
	Collapsed  bool
	ChildCount int // total descendants

	// Line tracking (set during rendering).
	StartLine int
	LineCount int

	// Grandparent poster for label resolution.
	GrandParentPoster string
}

// flatten performs a pre-order DFS of the comment tree and returns
// a flat slice. Each entry retains its depth and descendant count.
func flatten(story *item.Story) []FlatComment {
	var result []FlatComment

	grandParentPoster := ""

	for _, child := range story.Comments {
		flattenRecursive(child, 0, grandParentPoster, &result)
	}

	return result
}

func flattenRecursive(c *item.Story, depth int, grandParentPoster string, out *[]FlatComment) {
	if c.Content == "[deleted]" && len(c.Comments) == 0 {
		return
	}

	fc := FlatComment{
		Story:             c,
		Depth:             depth,
		ChildCount:        comment.DescendantCount(c),
		GrandParentPoster: grandParentPoster,
	}
	*out = append(*out, fc)

	gp := grandParentPoster
	if depth == 0 {
		gp = c.User
	}

	for _, reply := range c.Comments {
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

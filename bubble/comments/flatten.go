package comments

import (
	"clx/comment"
)

// FlatComment represents a single comment in the flattened view of the tree.
// It holds only tree-structural data — rendering artifacts like line positions
// are tracked separately in LineMetrics.
type FlatComment struct {
	Comment         *comment.Comment
	Depth           int
	Collapsed       bool
	DescendantCount int // total descendants
	TopLevelAuthor  string
}

// LineMetrics tracks the rendered position of a comment in the viewport.
// Indexed by flat index; recomputed on every render.
type LineMetrics struct {
	StartLine int
	LineCount int
}

// flatten performs a pre-order DFS of the comment tree and returns
// a flat slice. Each entry retains its depth and descendant count.
// The resulting order is load-bearing: computeVisible assumes children
// immediately follow their parent with strictly increasing depth.
func flatten(thread *comment.Thread) []FlatComment {
	var result []FlatComment

	topLevelAuthor := ""

	for _, child := range thread.Comments {
		flattenRecursive(child, 0, topLevelAuthor, &result)
	}

	fillDescendantCounts(result)

	return result
}

func flattenRecursive(c *comment.Comment, depth int, topLevelAuthor string, out *[]FlatComment) {
	if c.Content == "[deleted]" && len(c.Children) == 0 {
		return
	}

	fc := FlatComment{
		Comment:        c,
		Depth:          depth,
		TopLevelAuthor: topLevelAuthor,
	}
	*out = append(*out, fc)

	gp := topLevelAuthor
	if depth == 0 {
		gp = c.Author
	}

	for _, reply := range c.Children {
		flattenRecursive(reply, depth+1, gp, out)
	}
}

// fillDescendantCounts computes descendant counts directly from the flat
// array. In the DFS ordering, a node's descendants are the contiguous entries
// after it with strictly greater depth. This avoids walking the original tree,
// keeping the pruning decision in flattenRecursive as the single source of truth.
func fillDescendantCounts(flat []FlatComment) {
	for i := range flat {
		count := 0

		for j := i + 1; j < len(flat) && flat[j].Depth > flat[i].Depth; j++ {
			count++
		}

		flat[i].DescendantCount = count
		flat[i].Collapsed = flat[i].Depth == 0 && count > 0
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

		if fc.Collapsed && fc.DescendantCount > 0 {
			skipUntilDepth = fc.Depth
		}
	}

	return visible
}

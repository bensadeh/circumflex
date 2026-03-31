package comments

import (
	"clx/comment"
)

// flatComment represents a single comment in the flattened view of the tree.
// Comment is stored by value (not pointer) so mutations cannot affect the
// original tree. Rendering artifacts like line positions are tracked
// separately in lineMetrics.
type flatComment struct {
	Comment         comment.Comment
	Depth           int
	Collapsed       bool
	DescendantCount int // total descendants
	TopLevelAuthor  string
}

// lineMetrics tracks the rendered position of a comment in the viewport.
// Indexed by flat index; recomputed on every render.
type lineMetrics struct {
	SepStart  int // first line of the separator (before header)
	StartLine int // first line of the header
	LineCount int // lines from header through content (excludes separator)
}

// flatten performs a pre-order DFS of the comment tree and returns
// a flat slice. Each entry retains its depth and descendant count.
// The resulting order is load-bearing: computeVisible assumes children
// immediately follow their parent with strictly increasing depth.
func flatten(thread *comment.Thread) []flatComment {
	var result []flatComment

	topLevelAuthor := ""

	for _, child := range thread.Comments {
		flattenRecursive(child, 0, topLevelAuthor, &result)
	}

	fillDescendantCounts(result)
	collapseTopLevel(result)

	return result
}

func flattenRecursive(c *comment.Comment, depth int, topLevelAuthor string, out *[]flatComment) {
	if c.Content == "[deleted]" && len(c.Children) == 0 {
		return
	}

	copied := *c
	copied.Children = nil

	fc := flatComment{
		Comment:        copied,
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

// fillDescendantCounts computes descendant counts by walking backwards.
// Each node's count is the sum of its direct children's counts plus the
// number of direct children. O(n) instead of O(n²).
func fillDescendantCounts(flat []flatComment) {
	for i := len(flat) - 1; i >= 0; i-- {
		count := 0

		for j := i + 1; j < len(flat) && flat[j].Depth > flat[i].Depth; j += flat[j].DescendantCount + 1 {
			count += flat[j].DescendantCount + 1
		}

		flat[i].DescendantCount = count
	}
}

// collapseTopLevel sets the initial collapse state: top-level comments with
// children start collapsed so the user sees the full set of threads first.
func collapseTopLevel(flat []flatComment) {
	for i := range flat {
		flat[i].Collapsed = flat[i].Depth == 0 && flat[i].DescendantCount > 0
	}
}

// computeVisible returns indices into flat[] for comments that should be
// displayed, skipping children of collapsed nodes.
func computeVisible(flat []flatComment) []int {
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

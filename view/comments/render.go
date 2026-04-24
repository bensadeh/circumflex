package comments

import (
	"image/color"
	"strings"

	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/style"
)

// renderContext holds the stable parameters for rendering a comment view.
// These values don't change between rebuilds unless the window is resized
// or the model is re-created.
type renderContext struct {
	header          string // pre-computed meta block (raw, before margin wrapping)
	originalPoster  string
	firstCommentID  int
	commentWidth    int
	indent          int
	enableNerdFonts bool
	screenWidth     int
	viewportHeight  int
	lastVisited     int64
	story           storyFields // scalar fields needed for header rebuild on resize
	newComments     int
}

// storyFields holds the thread metadata needed to rebuild the comment header
// on window resize, without retaining the full comment tree.
type storyFields struct {
	URL           string
	Domain        string
	Author        string
	TimeAgo       string
	ID            int
	CommentsCount int
	Points        int
	Content       string
}

// renderedComment holds the pre-rendered output for a single flat comment.
// Built once by prerenderComments and reused by renderFromFlat on every
// collapse/expand, avoiding the expensive syntax/wrapping pipeline.
// Header and content are stored separately so that focus highlighting can
// swap in a pre-rendered focused header without re-running the expensive
// content pipeline. Rebuilt on window resize.
type renderedComment struct {
	sep              string // rendered separator (before the comment body)
	sepLines         int
	header           string // rendered header (author + label + time) with margins
	headerFocused    string // same header but with author highlighted
	headerLines      int    // line count is identical for both variants
	content          string // rendered comment content with indentation and margins
	contentLines     int
	repliesCollapsed string // replies indicator when collapsed (empty if no descendants)
	repliesLines     int    // line count of collapsed indicator
}

// prerenderComments renders every comment in flat upfront, so that subsequent
// collapse/expand operations only concatenate pre-rendered strings.
func prerenderComments(rc renderContext, flat []flatComment) []renderedComment {
	leftMargin := strings.Repeat(" ", layout.CommentSectionLeftMargin)
	contentWidth := rc.screenWidth - layout.CommentSectionLeftMargin
	commentWidth := min(contentWidth, rc.commentWidth)

	rendered := make([]renderedComment, len(flat))

	for i := range flat {
		fc := &flat[i]
		out := &rendered[i]

		sep := comment.Separator(fc.Depth, commentWidth, fc.Comment.ID, rc.firstCommentID)
		if sep != "" {
			indentedSep := style.PrefixLines(sep, leftMargin)
			out.sep = indentedSep
			out.sepLines = strings.Count(indentedSep, "\n")
		}

		// Reserve 1 col inside commentWidth for the colored indent symbol when
		// the comment is nested (depth >= 1). Top-level comments have no symbol.
		symbolCols := 0
		if fc.Depth > 0 {
			symbolCols = 1
		}

		// Fold symbolCols into the floor so adjustedCommentWidth >= MinCommentWidth
		// even at the plateau.
		indentCols := comment.EffectiveIndentColumns(fc.Depth, rc.indent, commentWidth, layout.MinCommentWidth+symbolCols)
		depthIndent := strings.Repeat(" ", indentCols)

		screenWidth := contentWidth - indentCols
		adjustedCommentWidth := commentWidth - indentCols - symbolCols

		pad := leftMargin + depthIndent

		header := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.enableNerdFonts, false)
		headerWithMargin := style.PrefixLines(header, pad)
		out.header = headerWithMargin
		out.headerLines = strings.Count(headerWithMargin, "\n")

		focusedHeader := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.enableNerdFonts, true)
		out.headerFocused = style.PrefixLines(focusedHeader, pad)

		// Render the comment content (expensive: syntax highlighting + wrapping).
		var fg color.Color
		if comment.IsMod(fc.Comment.Author) {
			fg = style.CommentModFg()
		}

		content := comment.RenderContent(&fc.Comment, fc.Depth, adjustedCommentWidth, screenWidth, rc.enableNerdFonts, fg)
		contentWithMargin := style.PrefixLines(content+"\n", pad)
		out.content = contentWithMargin
		out.contentLines = strings.Count(contentWithMargin, "\n")

		// Pre-render replies indicator (only shown when collapsed). The indicator
		// sits at the first hidden reply's author column so that toggling
		// collapse/expand swaps content at the same left edge. Children are always
		// at depth >= 1, so the floor accounts for the child's 1-col symbol. The
		// trailing +1 matches the header's 1-col offset where the ▎ would sit.
		if fc.DescendantCount > 0 {
			childIndentCols := comment.EffectiveIndentColumns(fc.Depth+1, rc.indent, commentWidth, layout.MinCommentWidth+1)
			indicatorIndent := strings.Repeat(" ", childIndentCols+1)

			collapsed := comment.RepliesIndicator(fc.DescendantCount, indicatorIndent, true)
			indentedCollapsed := style.PrefixLines(collapsed, leftMargin)
			out.repliesCollapsed = indentedCollapsed
			out.repliesLines = strings.Count(indentedCollapsed, "\n")
		}
	}

	return rendered
}

// renderFromFlat builds the full comment view content from the flat comment
// list, respecting fold state. It returns the rendered content, the number of
// content lines, and line metrics indexed by flat index for navigation.
//
// focusedFlatIdx selects which comment gets the focused header variant;
// pass -1 when no comment is focused (scroll mode).
//
// The pre-rendered slice (indexed by flat index, built by prerenderComments)
// avoids re-running the expensive syntax-highlighting and text-wrapping
// pipeline on every collapse/expand or focus change.
func renderFromFlat(rc renderContext, flat []flatComment, visible []int, prerendered []renderedComment, focusedFlatIdx int) (string, int, []lineMetrics) {
	var sb strings.Builder
	sb.WriteString(rc.header)
	sb.WriteString("\n")

	lineCount := strings.Count(rc.header, "\n") + 1

	metrics := make([]lineMetrics, len(flat))

	for _, flatIdx := range visible {
		fc := flat[flatIdx]
		pre := &prerendered[flatIdx]

		sepStart := lineCount

		sb.WriteString(pre.sep)
		lineCount += pre.sepLines

		startLine := lineCount

		if flatIdx == focusedFlatIdx {
			sb.WriteString(pre.headerFocused)
		} else {
			sb.WriteString(pre.header)
		}

		lineCount += pre.headerLines

		sb.WriteString(pre.content)
		lineCount += pre.contentLines

		if fc.DescendantCount > 0 && fc.Collapsed {
			sb.WriteString(pre.repliesCollapsed)
			lineCount += pre.repliesLines
		}

		metrics[flatIdx] = lineMetrics{
			SepStart:  sepStart,
			StartLine: startLine,
			LineCount: lineCount - startLine,
		}
	}

	contentLines := lineCount

	// Add bottom padding so the last comments can be scrolled to the top.
	sb.WriteString(strings.Repeat("\n", rc.viewportHeight))

	return sb.String(), contentLines, metrics
}

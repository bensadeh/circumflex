package comments

import (
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

		// Separator.
		sep := comment.Separator(fc.Depth, commentWidth, fc.Comment.ID, rc.firstCommentID)
		if sep != "" {
			indentedSep := style.PrefixLines(sep, leftMargin)
			out.sep = indentedSep
			out.sepLines = strings.Count(indentedSep, "\n")
		}

		depthIndent := comment.IndentString(fc.Depth)
		depthIndentLen := len(depthIndent)
		availableWidth := contentWidth - depthIndentLen
		adjustedCommentWidth := commentWidth - fc.Depth

		pad := leftMargin + depthIndent

		// Pre-render both header variants (normal and focused).
		header := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.enableNerdFonts, false)
		headerWithMargin := style.PrefixLines(header, pad)
		out.header = headerWithMargin
		out.headerLines = strings.Count(headerWithMargin, "\n")

		focusedHeader := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.enableNerdFonts, true)
		out.headerFocused = style.PrefixLines(focusedHeader, pad)

		// Render the comment content (expensive: syntax highlighting + wrapping).
		content := comment.RenderContent(&fc.Comment, fc.Depth, adjustedCommentWidth, availableWidth, rc.enableNerdFonts)
		contentWithMargin := style.PrefixLines(content+"\n", pad)
		out.content = contentWithMargin
		out.contentLines = strings.Count(contentWithMargin, "\n")

		// Pre-render replies indicator (only shown when collapsed).
		if fc.DescendantCount > 0 {
			collapsed := comment.RepliesIndicator(fc.DescendantCount, fc.Depth, adjustedCommentWidth, true)
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

		// Pick the focused or normal header variant.
		if flatIdx == focusedFlatIdx {
			sb.WriteString(pre.headerFocused)
		} else {
			sb.WriteString(pre.header)
		}

		lineCount += pre.headerLines

		sb.WriteString(pre.content)
		lineCount += pre.contentLines

		// Replies indicator: only shown when replies are collapsed.
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

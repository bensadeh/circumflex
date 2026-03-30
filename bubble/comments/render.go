package comments

import (
	"clx/comment"
	"clx/constants"
	"clx/settings"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

// renderContext holds the stable parameters for rendering a comment view.
// These values don't change between rebuilds unless the window is resized
// or the model is re-created.
type renderContext struct {
	header         string // pre-computed meta block (raw, before margin wrapping)
	originalPoster string
	firstCommentID int
	config         *settings.Config
	screenWidth    int
	viewportHeight int
	lastVisited    int64
}

// renderedComment holds the pre-rendered output for a single flat comment.
// Built once by prerenderComments and reused by renderFromFlat on every
// collapse/expand, avoiding the expensive syntax/wrapping pipeline.
// Header and content are stored separately so that focus highlighting can
// swap in a pre-rendered focused header without re-running the expensive
// content pipeline. Rebuilt on window resize.
type renderedComment struct {
	sep           string // rendered separator (before the comment body)
	sepLines      int
	header        string // rendered header (author + label + time) with margins
	headerFocused string // same header but with author highlighted
	headerLines   int    // line count is identical for both variants
	content       string // rendered comment content with indentation and margins
	contentLines  int
	fold          string // fold indicator (empty if no descendants)
	foldLines     int
}

// prerenderComments renders every comment in flat upfront, so that subsequent
// collapse/expand operations only concatenate pre-rendered strings.
func prerenderComments(rc renderContext, flat []flatComment) []renderedComment {
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)
	contentWidth := rc.screenWidth - constants.CommentSectionLeftMargin

	rendered := make([]renderedComment, len(flat))

	for i := range flat {
		fc := &flat[i]
		out := &rendered[i]

		// Separator.
		sep := comment.Separator(fc.Depth, rc.config.CommentWidth, fc.Comment.ID, rc.firstCommentID)
		if sep != "" {
			indentedSep, _ := text.WrapWithPad(sep, rc.screenWidth, leftMargin)
			out.sep = indentedSep
			out.sepLines = strings.Count(indentedSep, "\n")
		}

		depthIndent := comment.IndentString(fc.Depth)
		depthIndentLen := len(depthIndent)
		availableWidth := contentWidth - depthIndentLen
		adjustedCommentWidth := rc.config.CommentWidth - fc.Depth

		// Pre-render both header variants (normal and focused).
		header := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.config, false)
		headerWithDepth, _ := text.WrapWithPad(header, contentWidth, depthIndent)
		headerWithMargin, _ := text.WrapWithPad(headerWithDepth, rc.screenWidth, leftMargin)
		out.header = headerWithMargin
		out.headerLines = strings.Count(headerWithMargin, "\n")

		focusedHeader := comment.Header(&fc.Comment, fc.Depth, rc.originalPoster, fc.TopLevelAuthor, rc.lastVisited, rc.config, true)
		focusedWithDepth, _ := text.WrapWithPad(focusedHeader, contentWidth, depthIndent)
		focusedWithMargin, _ := text.WrapWithPad(focusedWithDepth, rc.screenWidth, leftMargin)
		out.headerFocused = focusedWithMargin

		// Render the comment content (expensive: syntax highlighting + wrapping).
		content := comment.RenderContent(&fc.Comment, fc.Depth, rc.config, adjustedCommentWidth, availableWidth)
		contentWithDepth, _ := text.WrapWithPad(content+"\n", contentWidth, depthIndent)
		contentWithMargin, _ := text.WrapWithPad(contentWithDepth, rc.screenWidth, leftMargin)
		out.content = contentWithMargin
		out.contentLines = strings.Count(contentWithMargin, "\n")

		// Pre-render fold indicator.
		if fc.DescendantCount > 0 {
			indicator := comment.FoldIndicator(fc.DescendantCount, fc.Depth, adjustedCommentWidth)
			indentedIndicator, _ := text.WrapWithPad(indicator, rc.screenWidth, leftMargin)
			out.fold = indentedIndicator
			out.foldLines = strings.Count(indentedIndicator, "\n")
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
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)

	// Indent the pre-computed header with left margin.
	header, _ := text.WrapWithPad(rc.header, rc.screenWidth, leftMargin)

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")

	lineCount := strings.Count(header, "\n") + 1

	metrics := make([]lineMetrics, len(flat))

	for _, flatIdx := range visible {
		fc := flat[flatIdx]
		pre := &prerendered[flatIdx]

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

		// Fold indicator for collapsed comments with children.
		if fc.Collapsed && fc.DescendantCount > 0 {
			sb.WriteString(pre.fold)
			lineCount += pre.foldLines
		}

		metrics[flatIdx] = lineMetrics{
			StartLine: startLine,
			LineCount: lineCount - startLine,
		}
	}

	contentLines := lineCount

	// Add bottom padding so the last comments can be scrolled to the top.
	sb.WriteString(strings.Repeat("\n", rc.viewportHeight))

	return sb.String(), contentLines, metrics
}

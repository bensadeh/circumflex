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
// Rebuilt on window resize.
type renderedComment struct {
	sep       string // rendered separator (before the comment body)
	sepLines  int
	body      string // rendered comment body with indentation and margins
	bodyLines int
	fold      string // fold indicator (empty if no descendants)
	foldLines int
}

// prerenderComments renders every comment in flat upfront, so that subsequent
// collapse/expand operations only concatenate pre-rendered strings.
func prerenderComments(rc renderContext, flat []FlatComment) []renderedComment {
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

		// Render the comment body.
		depthIndent := comment.IndentString(fc.Depth)
		depthIndentLen := len(depthIndent)
		availableWidth := contentWidth - depthIndentLen
		adjustedCommentWidth := rc.config.CommentWidth - fc.Depth

		body := comment.RenderBody(&fc.Comment, fc.Depth, rc.config, rc.originalPoster, fc.TopLevelAuthor,
			adjustedCommentWidth, availableWidth, rc.lastVisited)

		withDepth, _ := text.WrapWithPad(body, contentWidth, depthIndent)
		withMargin, _ := text.WrapWithPad(withDepth+"\n", rc.screenWidth, leftMargin)
		out.body = withMargin
		out.bodyLines = strings.Count(withMargin, "\n")

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
// This function has no knowledge of focus state — focus highlighting is
// applied separately in View().
//
// The pre-rendered slice (indexed by flat index, built by prerenderComments)
// avoids re-running the expensive syntax-highlighting and text-wrapping
// pipeline on every collapse/expand.
func renderFromFlat(rc renderContext, flat []FlatComment, visible []int, prerendered []renderedComment) (string, int, []LineMetrics) {
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)

	// Indent the pre-computed header with left margin.
	header, _ := text.WrapWithPad(rc.header, rc.screenWidth, leftMargin)

	var sb strings.Builder
	sb.WriteString(header)

	lineCount := strings.Count(header, "\n")

	metrics := make([]LineMetrics, len(flat))

	for _, flatIdx := range visible {
		fc := flat[flatIdx]
		pre := &prerendered[flatIdx]

		sb.WriteString(pre.sep)
		lineCount += pre.sepLines

		startLine := lineCount

		sb.WriteString(pre.body)
		lineCount += pre.bodyLines

		// Fold indicator for collapsed comments with children.
		if fc.Collapsed && fc.DescendantCount > 0 {
			sb.WriteString(pre.fold)
			lineCount += pre.foldLines
		}

		metrics[flatIdx] = LineMetrics{
			StartLine: startLine,
			LineCount: lineCount - startLine,
		}
	}

	contentLines := lineCount

	// Add bottom padding so the last comments can be scrolled to the top.
	sb.WriteString(strings.Repeat("\n", rc.viewportHeight))

	return sb.String(), contentLines, metrics
}

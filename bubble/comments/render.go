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

// renderFromFlat builds the full comment view content from the flat comment
// list, respecting fold state. It returns the rendered content, the number of
// content lines, and line metrics indexed by flat index for navigation.
// This function has no knowledge of focus state — focus highlighting is
// applied separately in View().
func renderFromFlat(rc renderContext, flat []FlatComment, visible []int) (string, int, []LineMetrics) {
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)
	contentWidth := rc.screenWidth - constants.CommentSectionLeftMargin

	// Indent the pre-computed header with left margin.
	header, _ := text.WrapWithPad(rc.header, rc.screenWidth, leftMargin)

	var sb strings.Builder
	sb.WriteString(header)

	lineCount := strings.Count(header, "\n")

	metrics := make([]LineMetrics, len(flat))

	for _, flatIdx := range visible {
		fc := flat[flatIdx]

		// Separator.
		sep := comment.Separator(fc.Depth, rc.config.CommentWidth, fc.Comment.ID, rc.firstCommentID)
		if sep != "" {
			indentedSep, _ := text.WrapWithPad(sep, rc.screenWidth, leftMargin)
			sb.WriteString(indentedSep)
			lineCount += strings.Count(indentedSep, "\n")
		}

		startLine := lineCount

		// Render the comment body.
		depthIndent := comment.IndentString(fc.Depth)
		depthIndentLen := len(depthIndent)
		availableWidth := contentWidth - depthIndentLen
		adjustedCommentWidth := rc.config.CommentWidth - fc.Depth

		rendered := comment.RenderBody(fc.Comment, fc.Depth, rc.config, rc.originalPoster, fc.GrandParentPoster,
			adjustedCommentWidth, availableWidth, rc.lastVisited)

		// Apply depth indentation then left margin.
		withDepth, _ := text.WrapWithPad(rendered, contentWidth, depthIndent)

		withMargin, _ := text.WrapWithPad(withDepth+"\n", rc.screenWidth, leftMargin)
		sb.WriteString(withMargin)
		lineCount += strings.Count(withMargin, "\n")

		// Fold indicator for collapsed comments with children.
		if fc.Collapsed && fc.ChildCount > 0 {
			indicator := comment.FoldIndicator(fc.ChildCount, fc.Depth)
			indentedIndicator, _ := text.WrapWithPad(indicator, rc.screenWidth, leftMargin)
			sb.WriteString(indentedIndicator)
			lineCount += strings.Count(indentedIndicator, "\n")
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

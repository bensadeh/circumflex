package comments

import (
	"clx/comment"
	"clx/constants"
	"clx/meta"
	"clx/settings"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

// renderFromFlat builds the full comment view content from the flat comment
// list, respecting fold state. It returns the rendered content, the number of
// content lines, and line metrics indexed by flat index for navigation.
// This function has no knowledge of focus state — focus highlighting is
// applied separately in View().
func renderFromFlat(thread *comment.Thread, flat []FlatComment, visible []int, config *settings.Config, screenWidth, viewportHeight int, lastVisited int64) (string, int, []LineMetrics) {
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)
	contentWidth := screenWidth - constants.CommentSectionLeftMargin

	newComments := comment.NewCommentsCount(thread, lastVisited)
	headerRaw := meta.CommentSectionMetaBlock(thread, config, newComments) + "\n\n"

	// Indent the header with left margin.
	header, _ := text.WrapWithPad(headerRaw, screenWidth, leftMargin)

	var sb strings.Builder
	sb.WriteString(header)

	lineCount := strings.Count(header, "\n")

	firstCommentID := comment.FirstCommentID(thread.Comments)

	metrics := make([]LineMetrics, len(flat))

	for _, flatIdx := range visible {
		fc := flat[flatIdx]

		// Separator.
		sep := comment.Separator(fc.Depth, config.CommentWidth, fc.Comment.ID, firstCommentID)
		if sep != "" {
			indentedSep, _ := text.WrapWithPad(sep, screenWidth, leftMargin)
			sb.WriteString(indentedSep)
			lineCount += strings.Count(indentedSep, "\n")
		}

		startLine := lineCount

		// Render the comment body.
		depthIndent := comment.IndentString(fc.Depth)
		depthIndentLen := len(depthIndent)
		availableWidth := contentWidth - depthIndentLen
		adjustedCommentWidth := config.CommentWidth - fc.Depth

		rendered := comment.RenderBody(fc.Comment, config, thread.Author, fc.GrandParentPoster,
			adjustedCommentWidth, availableWidth, lastVisited)

		// Apply depth indentation then left margin.
		withDepth, _ := text.WrapWithPad(rendered, contentWidth, depthIndent)

		withMargin, _ := text.WrapWithPad(withDepth+"\n", screenWidth, leftMargin)
		sb.WriteString(withMargin)
		lineCount += strings.Count(withMargin, "\n")

		// Fold indicator for collapsed comments with children.
		if fc.Collapsed && fc.ChildCount > 0 {
			indicator := comment.FoldIndicator(fc.ChildCount, fc.Depth)
			indentedIndicator, _ := text.WrapWithPad(indicator, screenWidth, leftMargin)
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
	sb.WriteString(strings.Repeat("\n", viewportHeight))

	return sb.String(), contentLines, metrics
}

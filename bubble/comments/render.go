package comments

import (
	"clx/comment"
	"clx/constants"
	"clx/item"
	"clx/meta"
	"clx/settings"
	"strings"

	"charm.land/lipgloss/v2"

	text "github.com/MichaelMure/go-term-text"
)

// renderFromFlat builds the full comment view content from the flat comment
// list, respecting fold state. It updates StartLine/LineCount on each visible
// FlatComment for navigation.
func renderFromFlat(story *item.Story, flat []FlatComment, visible []int, focusedIdx int, config *settings.Config, screenWidth int, lastVisited int64) string {
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)
	contentWidth := screenWidth - constants.CommentSectionLeftMargin

	newComments := comment.NewCommentsCount(story, lastVisited)
	headerRaw := meta.CommentSectionMetaBlock(story, config, newComments) + "\n\n"

	// Indent the header with left margin.
	header, _ := text.WrapWithPad(headerRaw, screenWidth, leftMargin)

	var sb strings.Builder
	sb.WriteString(header)

	lineCount := strings.Count(header, "\n")

	firstCommentID := comment.FirstCommentID(story.Comments)

	for vi, flatIdx := range visible {
		fc := &flat[flatIdx]

		// Separator.
		sep := comment.Separator(fc.Depth, config.CommentWidth, fc.Story.ID, firstCommentID)
		if sep != "" {
			indentedSep, _ := text.WrapWithPad(sep, screenWidth, leftMargin)
			sb.WriteString(indentedSep)
			lineCount += strings.Count(indentedSep, "\n")
		}

		fc.StartLine = lineCount

		// Render the comment body.
		depthIndent := comment.IndentString(fc.Depth)
		depthIndentLen := len(depthIndent)
		availableWidth := contentWidth - depthIndentLen
		adjustedCommentWidth := config.CommentWidth - fc.Depth

		rendered := comment.RenderBody(fc.Story, config, story.User, fc.GrandParentPoster,
			adjustedCommentWidth, availableWidth, lastVisited)

		// Apply depth indentation then left margin.
		withDepth, _ := text.WrapWithPad(rendered, contentWidth, depthIndent)

		isFocused := vi == focusedIdx
		if isFocused {
			withDepth = applyFocusIndicator(withDepth)
		}

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

		fc.LineCount = lineCount - fc.StartLine
	}

	return sb.String()
}

var focusStyle = lipgloss.NewStyle().Reverse(true)

func applyFocusIndicator(rendered string) string {
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		return rendered
	}

	lines[0] = focusStyle.Render(lines[0])

	return strings.Join(lines, "\n")
}

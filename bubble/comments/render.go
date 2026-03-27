package comments

import (
	"clx/comment"
	"clx/constants"
	"clx/item"
	"clx/meta"
	"clx/nerdfonts"
	"clx/settings"
	"clx/style"
	"clx/syntax"
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	text "github.com/MichaelMure/go-term-text"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

// renderFromFlat builds the full comment view content from the flat comment
// list, respecting fold state. It updates StartLine/LineCount on each visible
// FlatComment for navigation.
func renderFromFlat(story *item.Story, flat []FlatComment, visible []int, focusedIdx int, config *settings.Config, screenWidth int, lastVisited int64) string {
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)
	contentWidth := screenWidth - constants.CommentSectionLeftMargin

	newComments := countNewComments(story, lastVisited)
	headerRaw := meta.CommentSectionMetaBlock(story, config, newComments) + newParagraph

	// Indent the header with left margin.
	header, _ := text.WrapWithPad(headerRaw, screenWidth, leftMargin)

	var sb strings.Builder
	sb.WriteString(header)

	lineCount := strings.Count(header, "\n")

	firstCommentID := 0
	if len(story.Comments) > 0 {
		firstCommentID = story.Comments[0].ID
	}

	for vi, flatIdx := range visible {
		fc := &flat[flatIdx]
		fc.StartLine = lineCount

		// Separator.
		sep := commentSeparator(fc.Depth, config.CommentWidth, fc.Story.ID, firstCommentID)
		if sep != "" {
			indentedSep, _ := text.WrapWithPad(sep, screenWidth, leftMargin)
			sb.WriteString(indentedSep)
			lineCount += strings.Count(indentedSep, "\n")
		}

		// Render the comment body.
		depthIndent := indentString(fc.Depth)
		depthIndentLen := len(depthIndent)
		availableWidth := contentWidth - depthIndentLen
		adjustedCommentWidth := config.CommentWidth - fc.Depth

		rendered := renderSingleComment(fc.Story, config, story.User, fc.GrandParentPoster,
			adjustedCommentWidth, availableWidth, lastVisited)

		// Apply depth indentation then left margin.
		withDepth, _ := text.WrapWithPad(rendered, contentWidth, depthIndent)

		isFocused := vi == focusedIdx
		if isFocused {
			withDepth = applyFocusIndicator(withDepth)
		}

		withMargin, _ := text.WrapWithPad(withDepth+newLine, screenWidth, leftMargin)
		sb.WriteString(withMargin)
		lineCount += strings.Count(withMargin, "\n")

		// Fold indicator for collapsed comments with children.
		if fc.Collapsed && fc.ChildCount > 0 {
			indicator := foldIndicator(fc.ChildCount, fc.Depth)
			indentedIndicator, _ := text.WrapWithPad(indicator, screenWidth, leftMargin)
			sb.WriteString(indentedIndicator)
			lineCount += strings.Count(indentedIndicator, "\n")
		}

		fc.LineCount = lineCount - fc.StartLine
	}

	return sb.String()
}

var foldIndicatorBase = lipgloss.NewStyle().Faint(true).Italic(true)

func foldIndicator(childCount, depth int) string {
	replies := "replies"
	if childCount == 1 {
		replies = "reply"
	}

	label := fmt.Sprintf("▶ %d %s hidden", childCount, replies)
	indent := indentString(depth)

	return indent + foldIndicatorBase.Render(label) + newLine
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

func renderSingleComment(c *item.Story, config *settings.Config, originalPoster, grandParentPoster string,
	commentWidth, availableScreenWidth int, lastVisited int64,
) string {
	coloredIndentSymbol := syntax.ColorizeIndentSymbol(config.IndentationSymbol, c.Level)

	header := commentHeader(c, originalPoster, grandParentPoster, lastVisited, config)
	formattedComment := comment.Print(c.Content, config, commentWidth, availableScreenWidth)
	paddedComment, _ := text.WrapWithPad(formattedComment, availableScreenWidth, coloredIndentSymbol)

	return header + paddedComment
}

func commentHeader(c *item.Story, originalPoster, grandParentPoster string, lastVisited int64, config *settings.Config) string {
	indentSize := 0
	if c.Level > 0 {
		indentSize = 1
	}

	author := authorString(c.User, lastVisited, c.Time)
	authorLabel := authorLabelString(c.User, originalPoster, grandParentPoster, config.EnableNerdFonts)
	indentation := strings.Repeat(" ", indentSize)

	return indentation + author + authorLabel + style.Faint(c.TimeAgo) + newLine
}

func commentSeparator(level, commentWidth, currentCommentID, firstCommentID int) string {
	if currentCommentID == firstCommentID {
		return ""
	}

	if level != 0 {
		return newLine
	}

	return style.Faint(strings.Repeat("▁", commentWidth)) + newLine + newLine
}

func indentString(level int) string {
	if level == 0 {
		return ""
	}

	return strings.Repeat(" ", level-1)
}

func authorString(author string, lastVisited, timePosted int64) string {
	authorInBold := style.Bold(author) + " "

	if lastVisited < timePosted {
		return authorInBold + style.CommentNewIndicator("●") + " "
	}

	return authorInBold
}

var mods = []string{"dang", "tomhow"}

func authorLabelString(author, originalPoster, grandParentPoster string, enableNerdFonts bool) string {
	label := computeLabel(author, originalPoster, grandParentPoster, enableNerdFonts)
	if label == "" {
		return ""
	}

	return colorizeLabel(author, originalPoster, grandParentPoster, label)
}

func computeLabel(author, originalPoster, grandParentPoster string, nerdFonts bool) string {
	switch {
	case nerdFonts:
		return nerdfonts.Author + " "
	case isMod(author):
		return "mod "
	case author == originalPoster:
		return "OP "
	case author == grandParentPoster:
		return "GP "
	default:
		return ""
	}
}

func colorizeLabel(author, originalPoster, grandParentPoster, label string) string {
	switch {
	case isMod(author):
		return style.CommentMod(label)
	case author == originalPoster:
		return style.CommentOP(label)
	case author == grandParentPoster:
		return style.CommentGP(label)
	default:
		return ""
	}
}

func isMod(author string) bool {
	return slices.Contains(mods, author)
}

func countNewComments(comments *item.Story, lastVisited int64) int {
	count := 0
	countNewCommentsRecursive(comments, &count, lastVisited)

	return count
}

func countNewCommentsRecursive(comments *item.Story, count *int, lastVisited int64) {
	for _, reply := range comments.Comments {
		if lastVisited < reply.Time {
			*count++
		}

		countNewCommentsRecursive(reply, count, lastVisited)
	}
}

package tree

import (
	"clx/comment"
	"clx/constants"
	"clx/meta"
	"clx/settings"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

func Print(thread *comment.Thread, config *settings.Config, screenWidth int, lastVisited int64) string {
	commentSectionScreenWidth := screenWidth - constants.CommentSectionLeftMargin
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)

	newComments := comment.NewCommentsCount(thread, lastVisited)
	header := meta.CommentSectionMetaBlock(thread, config, newComments) + "\n\n"

	firstCommentID := comment.FirstCommentID(thread.Comments)

	var replies strings.Builder

	for _, reply := range thread.Comments {
		printReplies(reply, config, commentSectionScreenWidth, thread.Author, "", firstCommentID,
			lastVisited, &replies)
	}

	result, _ := text.WrapWithPad(header+replies.String()+"\n", screenWidth, leftMargin)

	return result
}

func printReplies(c *comment.Comment, config *settings.Config, screenWidth int, originalPoster string,
	grandParentPoster string, firstCommentID int, lastVisited int64, sb *strings.Builder,
) {
	if c.Content == "[deleted]" && len(c.Children) == 0 {
		return
	}

	indentation := comment.IndentString(c.Depth)
	indentSize := len(indentation)
	availableScreenWidth := screenWidth - indentSize - constants.CommentSectionLeftMargin
	adjustedCommentWidth := config.CommentWidth - c.Depth

	body := comment.RenderBody(c, config, originalPoster, grandParentPoster, adjustedCommentWidth, availableScreenWidth,
		lastVisited)
	indentedComment, _ := text.WrapWithPad(body, screenWidth, indentation)

	sep := comment.Separator(c.Depth, config.CommentWidth, c.ID, firstCommentID)
	sb.WriteString(sep + indentedComment + "\n")

	if c.Depth == 0 {
		grandParentPoster = c.Author
	}

	for _, reply := range c.Children {
		printReplies(reply, config, screenWidth, originalPoster, grandParentPoster, firstCommentID,
			lastVisited, sb)
	}
}

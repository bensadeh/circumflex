package tree

import (
	"clx/comment"
	"clx/constants"
	"clx/item"
	"clx/meta"
	"clx/settings"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

func Print(comments *item.Story, config *settings.Config, screenWidth int, lastVisited int64) string {
	commentSectionScreenWidth := screenWidth - constants.CommentSectionLeftMargin
	leftMargin := strings.Repeat(" ", constants.CommentSectionLeftMargin)

	newComments := comment.NewCommentsCount(comments, lastVisited)
	header := meta.CommentSectionMetaBlock(comments, config, newComments) + "\n\n"

	firstCommentID := comment.FirstCommentID(comments.Comments)

	var replies strings.Builder

	for _, reply := range comments.Comments {
		printReplies(reply, config, commentSectionScreenWidth, comments.User, "", firstCommentID,
			lastVisited, &replies)
	}

	result, _ := text.WrapWithPad(header+replies.String()+"\n", screenWidth, leftMargin)

	return result
}

func printReplies(c *item.Story, config *settings.Config, screenWidth int, originalPoster string,
	grandParentPoster string, firstCommentID int, lastVisited int64, sb *strings.Builder,
) {
	if c.Content == "[deleted]" && len(c.Comments) == 0 {
		return
	}

	indentation := comment.IndentString(c.Level)
	indentSize := len(indentation)
	availableScreenWidth := screenWidth - indentSize - constants.CommentSectionLeftMargin
	adjustedCommentWidth := config.CommentWidth - c.Level

	body := comment.RenderBody(c, config, originalPoster, grandParentPoster, adjustedCommentWidth, availableScreenWidth,
		lastVisited)
	indentedComment, _ := text.WrapWithPad(body, screenWidth, indentation)

	sep := comment.Separator(c.Level, config.CommentWidth, c.ID, firstCommentID)
	sb.WriteString(sep + indentedComment + "\n")

	if c.Level == 0 {
		grandParentPoster = c.User
	}

	for _, reply := range c.Comments {
		printReplies(reply, config, screenWidth, originalPoster, grandParentPoster, firstCommentID,
			lastVisited, sb)
	}
}

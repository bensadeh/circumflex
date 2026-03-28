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
		printReplies(reply, 0, config, commentSectionScreenWidth, thread.Author, "", firstCommentID,
			lastVisited, &replies)
	}

	result, _ := text.WrapWithPad(header+replies.String()+"\n", screenWidth, leftMargin)

	return result
}

func printReplies(c *comment.Comment, depth int, config *settings.Config, screenWidth int, originalPoster string,
	topLevelAuthor string, firstCommentID int, lastVisited int64, sb *strings.Builder,
) {
	if c.Content == "[deleted]" && len(c.Children) == 0 {
		return
	}

	indentation := comment.IndentString(depth)
	indentSize := len(indentation)
	availableScreenWidth := screenWidth - indentSize - constants.CommentSectionLeftMargin
	adjustedCommentWidth := config.CommentWidth - depth

	body := comment.RenderBody(c, depth, config, originalPoster, topLevelAuthor, adjustedCommentWidth, availableScreenWidth,
		lastVisited)
	indentedComment, _ := text.WrapWithPad(body, screenWidth, indentation)

	sep := comment.Separator(depth, config.CommentWidth, c.ID, firstCommentID)
	sb.WriteString(sep + indentedComment + "\n")

	if depth == 0 {
		topLevelAuthor = c.Author
	}

	for _, reply := range c.Children {
		printReplies(reply, depth+1, config, screenWidth, originalPoster, topLevelAuthor, firstCommentID,
			lastVisited, sb)
	}
}

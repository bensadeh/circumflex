package tree

import (
	"clx/comment"
	"clx/constants"
	"clx/item"
	"clx/meta"
	"clx/nerdfonts"
	"clx/settings"
	"clx/style"
	"clx/syntax"
	"clx/tree/formatter"
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

func Print(comments *item.Story, config *settings.Config, screenWidth int, lastVisited int64) string {
	commentSectionScreenWidth := screenWidth - constants.CommentSectionLeftMargin

	header := getHeader(comments, config, lastVisited)
	firstCommentID := getFirstCommentID(comments.Comments)

	var replies strings.Builder

	for _, reply := range comments.Comments {
		replies.WriteString(printReplies(reply, config, commentSectionScreenWidth, comments.User, "", firstCommentID,
			lastVisited))
	}

	commentSection := formatter.Process(header+replies.String()+newLine, screenWidth)

	return commentSection
}

func getFirstCommentID(comments []*item.Story) int {
	if len(comments) == 0 {
		return 0
	}

	return comments[0].ID
}

func getHeader(c *item.Story, config *settings.Config, lastVisited int64) string {
	newComments := getNewCommentsCount(c, lastVisited)

	return meta.GetCommentSectionMetaBlock(c, config, newComments) + newParagraph
}

func printReplies(c *item.Story, config *settings.Config, screenWidth int, originalPoster string,
	grandParentPoster string, firstCommentID int, lastVisited int64,
) string {
	isDeletedAndHasNoReplies := c.Content == "[deleted]" && len(c.Comments) == 0
	if isDeletedAndHasNoReplies {
		return ""
	}

	indentation := getIndentString(c.Level)
	indentSize := len(indentation)
	availableScreenWidth := screenWidth - indentSize - constants.CommentSectionLeftMargin
	adjustedCommentWidth := config.CommentWidth - c.Level

	comment := formatComment(c, config, originalPoster, grandParentPoster, adjustedCommentWidth, availableScreenWidth,
		lastVisited)
	indentedComment, _ := text.WrapWithPad(comment, screenWidth, indentation)
	fullComment := getSeparator(c.Level, config.CommentWidth, c.ID, firstCommentID) + indentedComment + newLine
	fullComment += getButton(c.Level, getReplyCount(c), config.CommentWidth)

	var fullCommentWithFilterTag strings.Builder
	fullCommentWithFilterTag.WriteString(addFilterTag(c.Level, fullComment))

	if c.Level == 0 {
		grandParentPoster = c.User
	}

	for _, reply := range c.Comments {
		fullCommentWithFilterTag.WriteString(printReplies(reply, config, screenWidth, originalPoster, grandParentPoster, firstCommentID,
			lastVisited))
	}

	return fullCommentWithFilterTag.String()
}

var (
	buttonNotPressedBase = lipgloss.NewStyle().Bold(true).AlignHorizontal(lipgloss.Center)
	buttonPressedBase    = lipgloss.NewStyle().Bold(true).Faint(true).AlignHorizontal(lipgloss.Center)
	buttonContainerBase  = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
)

func getButton(level int, replyCount int, commentWidth int) string {
	if replyCount == 0 || level != 0 {
		return ""
	}

	replies := "replies"
	if replyCount == 1 {
		replies = "reply"
	}

	buttonLabel := fmt.Sprintf("%d %s", replyCount, replies)

	buttonNotPressed := buttonNotPressedBase.SetString("▶ " + buttonLabel).String()
	buttonPressed := buttonPressedBase.SetString("▼ " + buttonLabel).String()

	s := buttonContainerBase.Width(commentWidth)

	return newLine + s.Render(buttonNotPressed) + constants.InvisibleCharacterForCollapse +
		newLine + s.Render(buttonPressed) + constants.InvisibleCharacterForExpansion + newLine
}

func addFilterTag(level int, fullComment string) string {
	if level == 0 {
		return fullComment
	}

	var fullCommentWithFilterTags strings.Builder

	lines := strings.Split(fullComment, "\n")

	for i, line := range lines {
		if i == len(lines)-1 {
			continue
		}

		fullCommentWithFilterTags.WriteString(line + constants.InvisibleCharacterForExpansion + "\n")
	}

	return fullCommentWithFilterTags.String()
}

func formatComment(c *item.Story, config *settings.Config, originalPoster string, grandParentPoster string, commentWidth int,
	availableScreenWidth int, lastVisited int64,
) string {
	coloredIndentSymbol := syntax.ColorizeIndentSymbol(config.IndentationSymbol, c.Level)

	header := getCommentHeader(c, originalPoster, grandParentPoster, lastVisited, config)
	formattedComment := comment.Print(c.Content, config, commentWidth, availableScreenWidth)

	paddedComment, _ := text.WrapWithPad(formattedComment, availableScreenWidth, coloredIndentSymbol)

	return header + paddedComment
}

func getSeparator(level int, commentWidth int, currentCommentID int, firstCommentID int) string {
	if currentCommentID == firstCommentID {
		return ""
	}

	if level != 0 || currentCommentID == firstCommentID {
		return newLine
	}

	return style.Faint(strings.Repeat("▁", commentWidth)) + newLine + newLine
}

func getIndentString(level int) string {
	if level == 0 {
		return ""
	}

	return strings.Repeat(" ", level-1)
}

func getCommentHeader(c *item.Story, originalPoster string, grandParentPoster string, lastVisited int64, config *settings.Config) string {
	if c.Level == 0 {
		return formatHeader(c, originalPoster, grandParentPoster, true,
			0, lastVisited, config)
	}

	return formatHeader(c, originalPoster, grandParentPoster, false,
		1, lastVisited, config)
}

func formatHeader(c *item.Story, originalPoster string, grandParentPoster string,
	enableZeroWidthSpace bool, indentSize int, lastVisited int64, config *settings.Config,
) string {
	author := getAuthor(c.User, lastVisited, c.Time)
	authorLabel := getAuthorLabel(c.User, originalPoster, grandParentPoster, config.EnableNerdFonts)
	zeroWidthSpace := getZeroWidthSpace(enableZeroWidthSpace)
	indentation := strings.Repeat(" ", indentSize)

	return zeroWidthSpace + indentation + author + authorLabel +
		style.Faint(c.TimeAgo) + newLine
}

func getAuthor(author string, lastVisited, timePosted int64) string {
	authorInBold := style.Bold(author) + " "

	commentIsNew := lastVisited < timePosted

	if commentIsNew {
		return authorInBold + style.CommentNewIndicator("●") + " "
	}

	return authorInBold
}

func getZeroWidthSpace(enabled bool) string {
	if enabled {
		return constants.InvisibleCharacterForTopLevelComments
	}

	return ""
}

func getReplyCount(comments *item.Story) int {
	numberOfReplies := 0

	return incrementReplyCount(comments, &numberOfReplies)
}

func incrementReplyCount(comments *item.Story, repliesSoFar *int) int {
	for _, reply := range comments.Comments {
		*repliesSoFar++
		incrementReplyCount(reply, repliesSoFar)
	}

	return *repliesSoFar
}

func getNewCommentsCount(comments *item.Story, lastVisited int64) int {
	numberOfReplies := 0

	return incrementNewCommentsCount(comments, &numberOfReplies, lastVisited)
}

func incrementNewCommentsCount(comments *item.Story, newCommentsSoFar *int, lastVisited int64) int {
	for _, reply := range comments.Comments {
		commentIsNew := lastVisited < reply.Time
		if commentIsNew {
			*newCommentsSoFar++
		}

		incrementNewCommentsCount(reply, newCommentsSoFar, lastVisited)
	}

	return *newCommentsSoFar
}

var mods = []string{"dang", "tomhow"}

func getAuthorLabel(author, originalPoster, grandParentPoster string, enableNerdFonts bool) string {
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

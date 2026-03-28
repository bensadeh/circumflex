package comment

import (
	"clx/nerdfonts"
	"clx/settings"
	"clx/style"
	"clx/syntax"
	"fmt"
	"slices"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

var mods = []string{"dang", "tomhow"}

// Author returns the formatted author name with an optional new-comment indicator.
func Author(author string, lastVisited, timePosted int64) string {
	authorInBold := style.Bold(author) + " "

	if lastVisited < timePosted {
		return authorInBold + style.CommentNewIndicator("\u25cf") + " "
	}

	return authorInBold
}

// AuthorLabel returns a styled label for special users (mod, OP, GP).
func AuthorLabel(author, originalPoster, grandParentPoster string, enableNerdFonts bool) string {
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
	case IsMod(author):
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
	case IsMod(author):
		return style.CommentMod(label)
	case author == originalPoster:
		return style.CommentOP(label)
	case author == grandParentPoster:
		return style.CommentGP(label)
	default:
		return ""
	}
}

// IsMod returns true if the author is a known moderator.
func IsMod(author string) bool {
	return slices.Contains(mods, author)
}

// Separator returns the visual separator between comments.
func Separator(depth, commentWidth, currentCommentID, firstCommentID int) string {
	if currentCommentID == firstCommentID {
		return ""
	}

	if depth != 0 {
		return "\n"
	}

	return style.Faint(strings.Repeat("\u2581", commentWidth)) + "\n\n"
}

// IndentString returns the indentation prefix for a given nesting depth.
func IndentString(depth int) string {
	if depth == 0 {
		return ""
	}

	return strings.Repeat(" ", depth-1)
}

// Header returns the formatted comment header line (author + label + time).
func Header(c *Comment, originalPoster, grandParentPoster string, lastVisited int64, config *settings.Config) string {
	indentSize := 0
	if c.Depth > 0 {
		indentSize = 1
	}

	author := Author(c.Author, lastVisited, c.Time)
	authorLabel := AuthorLabel(c.Author, originalPoster, grandParentPoster, config.EnableNerdFonts)
	indentation := strings.Repeat(" ", indentSize)

	return indentation + author + authorLabel + style.Faint(c.TimeAgo) + "\n"
}

// RenderBody returns the formatted comment with header, indent symbol, and content.
func RenderBody(c *Comment, config *settings.Config, originalPoster, grandParentPoster string,
	commentWidth, availableScreenWidth int, lastVisited int64,
) string {
	coloredIndentSymbol := syntax.ColorizeIndentSymbol(config.IndentationSymbol, c.Depth)

	header := Header(c, originalPoster, grandParentPoster, lastVisited, config)
	formattedComment := Print(c.Content, config, commentWidth, availableScreenWidth)
	paddedComment, _ := text.WrapWithPad(formattedComment, availableScreenWidth, coloredIndentSymbol)

	return header + paddedComment
}

// DescendantCount returns the total number of descendants of a comment,
// skipping deleted comments with no replies.
func DescendantCount(c *Comment) int {
	count := 0

	for _, reply := range c.Children {
		if reply.Content == "[deleted]" && len(reply.Children) == 0 {
			continue
		}

		count++
		count += DescendantCount(reply)
	}

	return count
}

// NewCommentsCount returns the number of new comments since lastVisited.
func NewCommentsCount(thread *Thread, lastVisited int64) int {
	count := 0

	for _, c := range thread.Comments {
		countNewComments(c, &count, lastVisited)
	}

	return count
}

func countNewComments(c *Comment, count *int, lastVisited int64) {
	if lastVisited < c.Time {
		*count++
	}

	for _, reply := range c.Children {
		countNewComments(reply, count, lastVisited)
	}
}

// FirstCommentID returns the ID of the first comment, or 0 if there are none.
func FirstCommentID(comments []*Comment) int {
	if len(comments) == 0 {
		return 0
	}

	return comments[0].ID
}

// FoldIndicator returns a styled fold indicator for collapsed comments.
func FoldIndicator(childCount, depth int) string {
	replies := "replies"
	if childCount == 1 {
		replies = "reply"
	}

	label := fmt.Sprintf("\u25b6 %d %s hidden", childCount, replies)
	indent := IndentString(depth)

	return indent + style.Faint(label) + "\n"
}

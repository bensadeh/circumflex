package comment

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"

	text "github.com/MichaelMure/go-term-text"
)

var mods = []string{"dang", "tomhow"}

// Author returns the formatted author name with an optional new-comment indicator.
// When focused is true, the author name is rendered with reverse video.
func Author(author string, lastVisited, timePosted int64, focused bool) string {
	var styledAuthor string
	if focused {
		styledAuthor = style.BoldReverse(author) + " "
	} else {
		styledAuthor = style.Bold(author) + " "
	}

	if lastVisited < timePosted {
		return styledAuthor + style.CommentNewIndicator("●") + " "
	}

	return styledAuthor
}

// AuthorLabel returns a styled label for special users (mod, OP, GP).
func AuthorLabel(author, originalPoster, topLevelAuthor string, enableNerdFonts bool) string {
	label := computeLabel(author, originalPoster, topLevelAuthor, enableNerdFonts)
	if label == "" {
		return ""
	}

	return colorizeLabel(author, originalPoster, topLevelAuthor, label)
}

func computeLabel(author, originalPoster, topLevelAuthor string, nerdFonts bool) string {
	switch {
	case IsMod(author):
		if nerdFonts {
			return nerdfonts.Author + " "
		}

		return "mod "
	case author == originalPoster:
		if nerdFonts {
			return nerdfonts.Author + " "
		}

		return "OP "
	case author == topLevelAuthor:
		if nerdFonts {
			return nerdfonts.Author + " "
		}

		return "GP "
	default:
		return ""
	}
}

func colorizeLabel(author, originalPoster, topLevelAuthor, label string) string {
	switch {
	case IsMod(author):
		return style.CommentMod(label)
	case author == originalPoster:
		return style.CommentOP(label)
	case author == topLevelAuthor:
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

	return style.Faint(strings.Repeat("▁", commentWidth)) + "\n\n"
}

// IndentString returns the indentation prefix for a given nesting depth.
func IndentString(depth int) string {
	if depth == 0 {
		return ""
	}

	return strings.Repeat(" ", depth-1)
}

// Header returns the formatted comment header line (author + label + time).
func Header(c *Comment, depth int, originalPoster, topLevelAuthor string, lastVisited int64, enableNerdFonts bool, focused bool) string {
	indentSize := 0
	if depth > 0 {
		indentSize = 1
	}

	author := Author(c.Author, lastVisited, c.Time, focused)
	authorLabel := AuthorLabel(c.Author, originalPoster, topLevelAuthor, enableNerdFonts)
	indentation := strings.Repeat(" ", indentSize)

	return indentation + author + authorLabel + style.Faint(c.TimeAgo) + "\n"
}

// RenderContent returns the formatted comment body (without header), with indent symbol.
func RenderContent(c *Comment, depth int, commentWidth, availableScreenWidth int, enableNerdFonts bool) string {
	coloredIndentSymbol := syntax.ColorizeIndentSymbol(style.IndentSymbol, depth)

	formattedComment := Render(c.Content, commentWidth, availableScreenWidth, enableNerdFonts)
	paddedComment, _ := text.WrapWithPad(formattedComment, availableScreenWidth, coloredIndentSymbol)

	return paddedComment
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

// RepliesIndicator returns a styled, left-aligned replies indicator line when
// replies are collapsed. When expanded, it returns an empty string.
func RepliesIndicator(descendantCount, depth, commentWidth int, collapsed bool) string {
	if !collapsed {
		return ""
	}

	replies := "replies"
	if descendantCount == 1 {
		replies = "reply"
	}

	label := fmt.Sprintf("↩ %d %s hidden", descendantCount, replies)
	indent := IndentString(depth)

	extra := " "
	if depth > 0 {
		extra = "  "
	}

	return "\n" + indent + extra + style.FaintItalic(label) + "\n"
}

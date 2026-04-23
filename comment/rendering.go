package comment

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"
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

// EffectiveIndentColumns returns the indent column count for a comment, capped
// so that (commentWidth - result) never drops below minCommentWidth. When the
// desired indent would push the remaining width below the floor, the indent
// plateaus and deeper comments share an ancestor's indent level — nesting is
// then conveyed by the depth-colored indent symbol alone. When commentWidth is
// already at or below minCommentWidth (very narrow terminal), no indent is
// applied. First-level replies (depth 1) inherit the top-level indent of 0,
// matching the historical rendering.
func EffectiveIndentColumns(depth, size, commentWidth, minCommentWidth int) int {
	if depth <= 1 {
		return 0
	}

	desired := (depth - 1) * size
	headroom := max(0, commentWidth-minCommentWidth)

	return min(desired, headroom)
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
// When fg is non-nil, paragraph text is tinted with that foreground color.
func RenderContent(c *Comment, depth int, commentWidth, screenWidth int, enableNerdFonts bool, fg color.Color) string {
	coloredIndentSymbol := syntax.ColorizeIndentSymbol(style.IndentSymbol, depth)

	formattedComment := Render(c.Content, commentWidth, screenWidth, enableNerdFonts, fg)

	padWidth := lipgloss.Width(coloredIndentSymbol)
	wrapped := lipgloss.Wrap(formattedComment, screenWidth-padWidth, "")

	return style.PrefixLines(wrapped, coloredIndentSymbol)
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
// replies are collapsed. When expanded, it returns an empty string. The caller
// is responsible for supplying the full leading whitespace (leftMargin aside)
// so that the ↩ marker lands at the first hidden reply's author column.
func RepliesIndicator(descendantCount int, indent string, collapsed bool) string {
	if !collapsed {
		return ""
	}

	replies := "replies"
	if descendantCount == 1 {
		replies = "reply"
	}

	label := fmt.Sprintf("↩ %d %s hidden", descendantCount, replies)

	return "\n" + indent + style.FaintItalic(label) + "\n"
}

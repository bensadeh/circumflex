package list

import (
	"fmt"
	"io"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/headline"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/timeago"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const (
	nerdFontSpacing = 2

	itemHeight  = 2
	itemSpacing = 1
)

// descStyle is the one style every description row shares; per-state title
// styling lives in headline.Render, the single source for it.
var descStyle = lipgloss.NewStyle().Faint(true)

func (m *Model) renderItem(w io.Writer, index int, item *hn.Story, f Frame) {
	enableNerdFonts := m.config.EnableNerdFonts

	domain := headline.HighlightDomain(item.Domain)

	score := scoreLabel(item.Points, enableNerdFonts)
	author := authorLabel(item.Author, enableNerdFonts)
	comments := commentsLabel(item.CommentsCount, enableNerdFonts)
	timeAgo := parseTime(item.Time, enableNerdFonts)

	var desc string

	if enableNerdFonts {
		spacing := strings.Repeat(" ", nerdFontSpacing)
		desc = score + spacing + comments + spacing + timeAgo + spacing + author
	} else {
		desc = score + author + timeAgo + comments
	}

	// The row renders to the right of the rank gutter, and the domain is
	// appended after the title below — both must fit inside the pane.
	textWidth := max(0, m.width-layout.MainViewLeftMargin)

	var (
		isSelected = index == m.Index()
		markAsRead = m.history.Contains(item.ID)
	)

	var state headline.HighlightType

	switch {
	case isSelected && f.Selection == SelectionAddFavorite:
		state = headline.AddToFavorites

	case isSelected && f.Selection == SelectionRemoveFavorite:
		state = headline.RemoveFromFavorites

	case isSelected && !m.dimmed(f):
		state = headline.Selected

	// While the detail pane is open the selected row renders a muted version
	// of the browsing highlight — for an open story, this is where J/K story
	// navigation currently is.
	case isSelected && m.detailOpen(f):
		state = headline.OpenStory

	case (markAsRead && m.cat.CurrentCategory() != categories.Favorites) ||
		m.dimmed(f):
		state = headline.MarkAsRead

	default:
		state = headline.Unselected
	}

	// Render the full title first, truncate the styled output after: the
	// escape-aware cut cannot strand a half-eaten token pattern, and nerd
	// glyph widths are already accounted for.
	title := headline.Render(item.Title, state, enableNerdFonts)

	if m.width > 0 {
		titleWidth := max(0, textWidth-xansi.StringWidth(domain)-1)
		title = xansi.Truncate(title, titleWidth, ellipsis)
		desc = xansi.Truncate(desc, textWidth, ellipsis)
	}

	title = title + " " + domain
	desc = descStyle.Render(desc)

	// In panes too narrow for the title budget above, the appended domain is
	// what overflows — clamp the assembled row.
	if m.width > 0 {
		title = xansi.Truncate(title, textWidth, ellipsis)
	}

	_, _ = fmt.Fprintf(w, "%s\n%s", title, desc)
}

func commentsLabel(numberOfComments int, enableNerdFonts bool) string {
	if numberOfComments == 0 && enableNerdFonts {
		return "      "
	}

	if numberOfComments == 0 {
		return ""
	}

	if enableNerdFonts {
		return fmt.Sprintf("%s%5d", nerdfonts.Comment, numberOfComments)
	}

	return fmt.Sprintf("| %d comments", numberOfComments)
}

func scoreLabel(score int, enableNerdFonts bool) string {
	if score == 0 {
		return ""
	}

	if enableNerdFonts {
		return fmt.Sprintf("%s%4d", nerdfonts.Score, score)
	}

	return fmt.Sprintf("%d points ", score)
}

func authorLabel(author string, enableNerdFonts bool) string {
	if author == "" {
		return ""
	}

	if enableNerdFonts {
		return fmt.Sprintf("%s %s", nerdfonts.Author, author)
	}

	return fmt.Sprintf("by %s ", author)
}

func parseTime(unixTime int64, enableNerdFonts bool) string {
	relative := timeago.RelativeTime(unixTime)

	if enableNerdFonts {
		return fmt.Sprintf("%s %-12s", nerdfonts.Time, relative)
	}

	return fmt.Sprintf("%s ", relative)
}

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

type itemStyles struct {
	normalTitle lipgloss.Style
	normalDesc  lipgloss.Style

	selectedTitle lipgloss.Style
	selectedDesc  lipgloss.Style

	openStoryTitle lipgloss.Style
	openStoryDesc  lipgloss.Style

	markAsReadTitle lipgloss.Style
	markAsReadDesc  lipgloss.Style

	selectedTitleAddToFavorites lipgloss.Style
	selectedDescAddToFavorites  lipgloss.Style

	selectedTitleRemoveFromFavorites lipgloss.Style
	selectedDescRemoveFromFavorites  lipgloss.Style
}

func newItemStyles() (s itemStyles) {
	s.normalTitle = lipgloss.NewStyle()
	s.normalDesc = s.normalTitle.Faint(true)

	s.selectedTitle = lipgloss.NewStyle().Reverse(true)
	s.selectedDesc = s.selectedTitle.Bold(false).Faint(true).Reverse(false)

	// A muted highlight bar: the color scheme's bright black as background,
	// text in the scheme's default foreground. Reversing a bright-black
	// foreground would look the same on most schemes but renders the row
	// invisible where bright black equals the background (Solarized Dark);
	// this way the text always keeps the scheme's normal contrast.
	s.openStoryTitle = lipgloss.NewStyle().Background(lipgloss.BrightBlack)
	s.openStoryDesc = s.normalDesc

	s.markAsReadTitle = s.normalTitle.Italic(true).Faint(true)
	s.markAsReadDesc = s.normalDesc

	s.selectedTitleAddToFavorites = s.normalTitle.Foreground(lipgloss.Green).Reverse(true)
	s.selectedDescAddToFavorites = s.normalDesc

	s.selectedTitleRemoveFromFavorites = s.normalTitle.Foreground(lipgloss.Red).Reverse(true)
	s.selectedDescRemoveFromFavorites = s.normalDesc

	return s
}

func (m *Model) renderItem(w io.Writer, index int, item *hn.Story, f Frame) {
	s := &m.itemStyles
	enableNerdFonts := m.config.EnableNerdFonts

	title := headline.ReplaceSpecialContentTags(item.Title, enableNerdFonts)
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
	textWidth := max(0, m.width-layout.MainViewLeftMargin-s.normalTitle.GetPaddingLeft()-s.normalTitle.GetPaddingRight())

	if m.width > 0 {
		titleWidth := max(0, textWidth-xansi.StringWidth(domain)-1)
		title = xansi.Truncate(title, titleWidth, ellipsis)
		desc = xansi.Truncate(desc, textWidth, ellipsis)
	}

	var (
		isSelected = index == m.Index()
		markAsRead = m.history.Contains(item.ID)
	)

	switch {
	case isSelected && f.Selection == SelectionAddFavorite:
		title, desc = styleTitleAndDesc(title, s.selectedTitleAddToFavorites, s.selectedDescAddToFavorites, domain,
			desc, headline.AddToFavorites, enableNerdFonts)

	case isSelected && f.Selection == SelectionRemoveFavorite:
		title, desc = styleTitleAndDesc(title, s.selectedTitleRemoveFromFavorites, s.selectedDescRemoveFromFavorites, domain,
			desc, headline.RemoveFromFavorites, enableNerdFonts)

	case isSelected && !m.dimmed(f):
		title, desc = styleTitleAndDesc(title, s.selectedTitle, s.selectedDesc, domain,
			desc, headline.Selected, enableNerdFonts)

	// While the detail pane is open the selected row renders a muted version
	// of the browsing highlight — for an open story, this is where J/K story
	// navigation currently is.
	case isSelected && m.detailOpen(f):
		title, desc = styleTitleAndDesc(title, s.openStoryTitle, s.openStoryDesc, domain,
			desc, headline.OpenStory, enableNerdFonts)

	case (markAsRead && m.cat.CurrentCategory() != categories.Favorites) ||
		m.dimmed(f):
		title, desc = styleTitleAndDesc(title, s.markAsReadTitle, s.markAsReadDesc, domain,
			desc, headline.MarkAsRead, enableNerdFonts)

	default:
		title, desc = styleTitleAndDesc(title, s.normalTitle, s.normalDesc, domain,
			desc, headline.Unselected, enableNerdFonts)
	}

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

func styleTitleAndDesc(title string, titleStyle lipgloss.Style, descStyle lipgloss.Style, domain string, desc string,
	highlightType headline.HighlightType, enableNerdFont bool,
) (string, string) {
	title = titleStyle.Render(title)
	title = headline.HighlightYCStartupsInHeadlines(title, highlightType, enableNerdFont)
	title = headline.HighlightYear(title, highlightType)
	title = headline.HighlightHackerNewsHeadlines(title, highlightType)
	title = headline.HighlightSpecialContent(title, highlightType, enableNerdFont)

	title = title + " " + domain
	desc = descStyle.Render(desc)

	return title, desc
}

func parseTime(unixTime int64, enableNerdFonts bool) string {
	relative := timeago.RelativeTime(unixTime)

	if enableNerdFonts {
		return fmt.Sprintf("%s %-12s", nerdfonts.Time, relative)
	}

	return fmt.Sprintf("%s ", relative)
}

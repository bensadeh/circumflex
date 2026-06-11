package list

import (
	"fmt"
	"io"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/syntax"
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

	s.markAsReadTitle = s.normalTitle.Italic(true).Faint(true)
	s.markAsReadDesc = s.normalDesc

	s.selectedTitleAddToFavorites = s.normalTitle.Foreground(lipgloss.Green).Reverse(true)
	s.selectedDescAddToFavorites = s.normalDesc

	s.selectedTitleRemoveFromFavorites = s.normalTitle.Foreground(lipgloss.Red).Reverse(true)
	s.selectedDescRemoveFromFavorites = s.normalDesc

	return s
}

func (m *Model) renderItem(w io.Writer, index int, item *hn.Story) {
	s := &m.itemStyles
	enableNerdFonts := m.config.EnableNerdFonts

	title := syntax.ReplaceSpecialContentTags(item.Title, enableNerdFonts)
	domain := syntax.HighlightDomain(item.Domain)

	score := getScore(item.Points, enableNerdFonts)
	author := getAuthor(item.Author, enableNerdFonts)
	comments := getComments(item.CommentsCount, enableNerdFonts)
	timeAgo := parseTime(item.Time, enableNerdFonts)

	var desc string

	if enableNerdFonts {
		spacing := strings.Repeat(" ", nerdFontSpacing)
		desc = score + spacing + comments + spacing + timeAgo + spacing + author
	} else {
		desc = score + author + timeAgo + comments
	}

	if m.width > 0 {
		textWidth := m.width - s.normalTitle.GetPaddingLeft() - s.normalTitle.GetPaddingRight()
		title = xansi.Truncate(title, textWidth, ellipsis)
		desc = xansi.Truncate(desc, textWidth, ellipsis)
	}

	var (
		isSelected = index == m.Index()
		markAsRead = m.history.Contains(item.ID)
	)

	switch {
	case isSelected && m.state == StateAddFavoritesPrompt:
		title, desc = styleTitleAndDesc(title, s.selectedTitleAddToFavorites, s.selectedDescAddToFavorites, domain,
			desc, syntax.AddToFavorites, enableNerdFonts)

	case isSelected && m.state == StateRemoveFavoritesPrompt:
		title, desc = styleTitleAndDesc(title, s.selectedTitleRemoveFromFavorites, s.selectedDescRemoveFromFavorites, domain,
			desc, syntax.RemoveFromFavorites, enableNerdFonts)

	case isSelected && m.state == StateBrowsing:
		title, desc = styleTitleAndDesc(title, s.selectedTitle, s.selectedDesc, domain,
			desc, syntax.Selected, enableNerdFonts)

	case (markAsRead && m.cat.CurrentCategory() != categories.Favorites) ||
		m.pager.transition != nil || m.state == StateReaderView:
		title, desc = styleTitleAndDesc(title, s.markAsReadTitle, s.markAsReadDesc, domain,
			desc, syntax.MarkAsRead, enableNerdFonts)

	default:
		title, desc = styleTitleAndDesc(title, s.normalTitle, s.normalDesc, domain,
			desc, syntax.Unselected, enableNerdFonts)
	}

	_, _ = fmt.Fprintf(w, "%s\n%s", title, desc)
}

func getComments(numberOfComments int, enableNerdFonts bool) string {
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

func getScore(score int, enableNerdFonts bool) string {
	if score == 0 {
		return ""
	}

	if enableNerdFonts {
		return fmt.Sprintf("%s%4d", nerdfonts.Score, score)
	}

	return fmt.Sprintf("%d points ", score)
}

func getAuthor(author string, enableNerdFonts bool) string {
	if author == "" {
		return ""
	}

	if enableNerdFonts {
		return fmt.Sprintf("%s %s", nerdfonts.Author, author)
	}

	return fmt.Sprintf("by %s ", author)
}

func styleTitleAndDesc(title string, titleStyle lipgloss.Style, descStyle lipgloss.Style, domain string, desc string,
	syntaxStyle syntax.HighlightType, enableNerdFont bool,
) (string, string) {
	title = titleStyle.Render(title)
	title = syntax.HighlightYCStartupsInHeadlines(title, syntaxStyle, enableNerdFont)
	title = syntax.HighlightYear(title, syntaxStyle)
	title = syntax.HighlightHackerNewsHeadlines(title, syntaxStyle)
	title = syntax.HighlightSpecialContent(title, syntaxStyle, enableNerdFont)

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

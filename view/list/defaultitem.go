package list

import (
	"fmt"
	"io"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/item"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/syntax"
	"github.com/bensadeh/circumflex/timeago"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const nerdFontSpacing = 2

type DefaultItemStyles struct {
	NormalTitle lipgloss.Style
	NormalDesc  lipgloss.Style

	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style

	MarkAsReadTitle lipgloss.Style
	MarkAsReadDesc  lipgloss.Style

	SelectedTitleAddToFavorites lipgloss.Style
	SelectedDescAddToFavorites  lipgloss.Style

	SelectedTitleRemoveFromFavorites lipgloss.Style
	SelectedDescRemoveFromFavorites  lipgloss.Style
}

func NewDefaultItemStyles() (s DefaultItemStyles) {
	s.NormalTitle = lipgloss.NewStyle()
	s.NormalDesc = s.NormalTitle.Copy().
		Faint(true)

	s.SelectedTitle = lipgloss.NewStyle().
		Reverse(true)

	s.SelectedDesc = s.SelectedTitle.Copy().
		Bold(false).
		Faint(true).
		Reverse(false)

	s.MarkAsReadTitle = s.NormalTitle.Copy().Italic(true).Faint(true)
	s.MarkAsReadDesc = s.NormalDesc.Copy()

	s.SelectedTitleAddToFavorites = s.NormalTitle.Copy().Foreground(lipgloss.Green).Reverse(true)
	s.SelectedDescAddToFavorites = s.NormalDesc.Copy()

	s.SelectedTitleRemoveFromFavorites = s.NormalTitle.Copy().Foreground(lipgloss.Red).Reverse(true)
	s.SelectedDescRemoveFromFavorites = s.NormalDesc.Copy()

	return s
}

type DefaultDelegate struct {
	Styles  DefaultItemStyles
	spacing int
}

// NewDefaultDelegate creates a new delegate with default styles.
func NewDefaultDelegate() *DefaultDelegate {
	return &DefaultDelegate{
		Styles:  NewDefaultItemStyles(),
		spacing: 1,
	}
}

// Height returns the delegate's preferred height.
func (d *DefaultDelegate) Height() int {
	return 2
}

// Spacing returns the delegate's spacing.
func (d *DefaultDelegate) Spacing() int {
	return d.spacing
}

// Update is a no-op; satisfies the ItemDelegate interface.
func (d *DefaultDelegate) Update(tea.Msg, *Model) tea.Cmd {
	return nil
}

// Render prints an item.
func (d *DefaultDelegate) Render(w io.Writer, m *Model, index int, item *item.Story) {
	var (
		title, desc, domain string
		s                   = &d.Styles
	)

	enableNerdFonts := m.config.EnableNerdFonts

	title = item.Title
	title = syntax.ReplaceSpecialContentTags(title, enableNerdFonts)

	domain = syntax.HighlightDomain(item.Domain)

	score := getScore(item.Points, enableNerdFonts)
	author := getAuthor(item.User, enableNerdFonts)
	comments := getComments(item.CommentsCount, enableNerdFonts)
	timeAgo := parseTime(item.Time, enableNerdFonts)

	if enableNerdFonts {
		spacing := strings.Repeat(" ", nerdFontSpacing)
		desc = score + spacing + comments + spacing + timeAgo + spacing + author
	} else {
		desc = score + author + timeAgo + comments
	}

	// Prevent text from exceeding list width
	if m.width > 0 {
		textWidth := m.width - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()
		title = xansi.Truncate(title, textWidth, ellipsis)
		desc = xansi.Truncate(desc, textWidth, ellipsis)
	}

	var (
		isSelected = index == m.Index()
		markAsRead = m.history.Contains(item.ID)
	)

	switch {
	case isSelected && m.state == StateAddFavoritesPrompt:
		title, desc = styleTitleAndDesc(title, s.SelectedTitleAddToFavorites, s.SelectedDescAddToFavorites, domain,
			desc, syntax.AddToFavorites, enableNerdFonts)

	case isSelected && m.state == StateRemoveFavoritesPrompt:
		title, desc = styleTitleAndDesc(title, s.SelectedTitleRemoveFromFavorites, s.SelectedDescRemoveFromFavorites, domain,
			desc, syntax.RemoveFromFavorites, enableNerdFonts)

	case isSelected && m.state == StateBrowsing:
		title, desc = styleTitleAndDesc(title, s.SelectedTitle, s.SelectedDesc, domain,
			desc, syntax.Selected, enableNerdFonts)

	case markAsRead && m.cat.CurrentCategory() != categories.Favorites:
		title, desc = styleTitleAndDesc(title, s.MarkAsReadTitle.Italic(true), s.MarkAsReadDesc, domain,
			desc, syntax.MarkAsRead, enableNerdFonts)

	case m.pager.transition != nil || m.state == StateReaderView:
		title, desc = styleTitleAndDesc(title, s.MarkAsReadTitle.Italic(true), s.MarkAsReadDesc, domain,
			desc, syntax.MarkAsRead, enableNerdFonts)

	default:
		title, desc = styleTitleAndDesc(title, s.NormalTitle, s.NormalDesc, domain,
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

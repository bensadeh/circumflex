package list

import (
	"fmt"
	"io"
	"strings"

	"clx/constants/nerdfonts"

	"clx/constants/category"
	"clx/item"
	"clx/syntax"

	"github.com/nleeper/goment"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

type DefaultItemStyles struct {
	NormalTitle lipgloss.Style
	NormalDesc  lipgloss.Style

	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style

	MarkAsReadTitle lipgloss.Style
	MarkAsReadDesc  lipgloss.Style

	SelectedTitleAddToFavorites lipgloss.Style
	SelectedDescAddToFavorites  lipgloss.Style

	SelectedTitleRemoveFromFavorites         lipgloss.Style
	SelectedDescRemoveFromFavoritesFavorites lipgloss.Style

	DimmedTitle lipgloss.Style
	DimmedDesc  lipgloss.Style
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

	s.SelectedTitleAddToFavorites = s.NormalTitle.Copy().Foreground(lipgloss.Color("2")).Reverse(true)
	s.SelectedDescAddToFavorites = s.NormalDesc.Copy()

	s.SelectedTitleRemoveFromFavorites = s.NormalTitle.Copy().Foreground(lipgloss.Color("1")).Reverse(true)
	s.SelectedDescRemoveFromFavoritesFavorites = s.NormalDesc.Copy()

	s.DimmedTitle = lipgloss.NewStyle()
	s.DimmedDesc = s.DimmedTitle.Copy()

	return s
}

type DefaultDelegate struct {
	ShowDescription bool
	Styles          DefaultItemStyles
	UpdateFunc      func(tea.Msg, *Model) tea.Cmd
	ShortHelpFunc   func() []key.Binding
	FullHelpFunc    func() [][]key.Binding
	spacing         int
}

// NewDefaultDelegate creates a new delegate with default styles.
func NewDefaultDelegate() DefaultDelegate {
	return DefaultDelegate{
		ShowDescription: true,
		Styles:          NewDefaultItemStyles(),
		spacing:         1,
	}
}

// Height returns the delegate's preferred height.
func (d DefaultDelegate) Height() int {
	if d.ShowDescription {
		return 2 //nolint:gomnd
	}
	return 1
}

// SetSpacing set the delegate's spacing.
func (d *DefaultDelegate) SetSpacing(i int) {
	d.spacing = i
}

// Spacing returns the delegate's spacing.
func (d DefaultDelegate) Spacing() int {
	return d.spacing
}

// Update checks whether the delegate's UpdateFunc is set and calls it.
func (d DefaultDelegate) Update(msg tea.Msg, m *Model) tea.Cmd {
	if d.UpdateFunc == nil {
		return nil
	}
	return d.UpdateFunc(msg, m)
}

// Render prints an item.
func (d DefaultDelegate) Render(w io.Writer, m Model, index int, item *item.Item) {
	var (
		title, desc, domain string
		s                   = &d.Styles
	)

	enableNerdFonts := m.config.EnableNerdFonts

	title = item.Title

	domain = syntax.HighlightDomain(item.Domain)

	score := getScore(item.Points, enableNerdFonts)
	author := getAuthor(item.User, enableNerdFonts)
	comments := getComments(item.CommentsCount, enableNerdFonts)
	time := parseTime(item.Time, enableNerdFonts)

	if enableNerdFonts {
		spacingSize := 2
		spacing := strings.Repeat(" ", spacingSize)
		desc = score + spacing + comments + spacing + time + spacing + author
	} else {
		desc = score + author + time + comments
	}

	// Prevent text from exceeding list width
	if m.width > 0 {
		textWidth := uint(m.width - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
		title = truncate.StringWithTail(title, textWidth, ellipsis)
		desc = truncate.StringWithTail(desc, textWidth, ellipsis)
	}

	var (
		isSelected = index == m.Index()
		markAsRead = m.history.Contains(item.ID)
	)

	switch {
	case isSelected && m.onAddToFavoritesPrompt:
		title, desc = styleTitleAndDesc(title, s.SelectedTitleAddToFavorites, s.SelectedDescAddToFavorites, domain,
			desc, syntax.AddToFavorites, m.config.DisableHeadlineHighlighting, enableNerdFonts)

	case isSelected && m.onRemoveFromFavoritesPrompt:
		title, desc = styleTitleAndDesc(title, s.SelectedTitleRemoveFromFavorites, s.SelectedDescRemoveFromFavoritesFavorites, domain,
			desc, syntax.RemoveFromFavorites, m.config.DisableHeadlineHighlighting, enableNerdFonts)

	case isSelected && !m.disableInput:
		title, desc = styleTitleAndDesc(title, s.SelectedTitle, s.SelectedDesc, domain,
			desc, syntax.Selected, m.config.DisableHeadlineHighlighting, enableNerdFonts)

	case markAsRead && m.category != category.Favorites:
		title, desc = styleTitleAndDesc(title, s.MarkAsReadTitle.Italic(true), s.MarkAsReadDesc, domain,
			desc, syntax.MarkAsRead, m.config.DisableHeadlineHighlighting, enableNerdFonts)

	case m.disableInput && !(m.onAddToFavoritesPrompt || m.onRemoveFromFavoritesPrompt):
		title, desc = styleTitleAndDesc(title, s.MarkAsReadTitle.Italic(false), s.MarkAsReadDesc, domain,
			desc, syntax.MarkAsRead, m.config.DisableHeadlineHighlighting, enableNerdFonts)

	default:
		title, desc = styleTitleAndDesc(title, s.NormalTitle, s.NormalDesc, domain,
			desc, syntax.Unselected, m.config.DisableHeadlineHighlighting, enableNerdFonts)
	}

	if d.ShowDescription {
		_, _ = fmt.Fprintf(w, "%s\n%s", title, desc)
		return
	}
	_, _ = fmt.Fprintf(w, "%s", title)
}

func getComments(numberOfComments int, enableNerdFonts bool) string {
	if numberOfComments == 0 && enableNerdFonts {
		return "     "
	}

	if numberOfComments == 0 {
		return ""
	}

	if enableNerdFonts {
		return fmt.Sprintf("%s%4d", nerdfonts.Comment, numberOfComments)
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
	syntaxStyle int, disableHeadlineHighlighting bool, enableNerdFont bool,
) (string, string) {
	title = titleStyle.Render(title)

	if !disableHeadlineHighlighting {
		title = syntax.HighlightYCStartupsInHeadlines(title, syntaxStyle, enableNerdFont)
		title = syntax.HighlightYear(title, syntaxStyle, enableNerdFont)
		title = syntax.HighlightHackerNewsHeadlines(title, syntaxStyle)
		title = syntax.HighlightSpecialContent(title, syntaxStyle, enableNerdFont)
	}

	title = title + " " + domain
	desc = descStyle.Render(desc)

	return title, desc
}

func parseTime(unixTime int64, enableNerdFonts bool) string {
	moment, _ := goment.Unix(unixTime)
	now, _ := goment.New()

	if enableNerdFonts {
		return fmt.Sprintf("%s %-14s", nerdfonts.Time, moment.From(now))
	}

	return fmt.Sprintf("%s ", moment.From(now))
}

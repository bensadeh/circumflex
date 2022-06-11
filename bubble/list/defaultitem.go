package list

import (
	"clx/item"
	"clx/syntax"
	"fmt"
	"github.com/nleeper/goment"
	"io"

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

	title = item.Title
	domain = syntax.HighlightDomain(item.Domain)
	desc = fmt.Sprintf("%d points by %s %s | %d comments",
		item.Points, item.User, parseTime(item.Time), item.CommentsCount)

	// Prevent text from exceeding list width
	if m.width > 0 {
		textWidth := uint(m.width - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
		title = truncate.StringWithTail(title, textWidth, ellipsis)
		desc = truncate.StringWithTail(desc, textWidth, ellipsis)
	}

	var (
		isSelected = index == m.Index()
		markAsRead = m.history.Contains(item.ID)
		faint      = "\033[2m"
		italic     = "\033[3m"
	)

	switch {
	case isSelected && m.onAddToFavoritesPrompt:
		title = s.SelectedTitleAddToFavorites.Render(title)
		title = syntax.HighlightYCStartupsInHeadlines(title, syntax.Green)
		title = syntax.HighlightYearInHeadlines(title, syntax.Green)
		title = syntax.HighlightHackerNewsHeadlines(title, syntax.Green)
		title = syntax.HighlightSpecialContent(title)

		title = title + " " + domain
		desc = s.SelectedDescAddToFavorites.Render(desc)

	case isSelected && m.onRemoveFromFavoritesPrompt:
		title = s.SelectedTitleRemoveFromFavorites.Render(title)
		title = syntax.HighlightYCStartupsInHeadlines(title, syntax.Red)
		title = syntax.HighlightYearInHeadlines(title, syntax.Red)
		title = syntax.HighlightHackerNewsHeadlines(title, syntax.Red)
		title = syntax.HighlightSpecialContent(title)

		title = title + " " + domain
		desc = s.SelectedDescRemoveFromFavoritesFavorites.Render(desc)

	case isSelected && !m.disableInput:
		title = s.SelectedTitle.Render(title)
		title = syntax.HighlightYCStartupsInHeadlines(title, syntax.Reverse)
		title = syntax.HighlightYearInHeadlines(title, syntax.Reverse)
		title = syntax.HighlightHackerNewsHeadlines(title, syntax.Reverse)
		title = syntax.HighlightSpecialContent(title)

		title = title + " " + domain
		desc = s.SelectedDesc.Render(desc)

	case markAsRead:
		title = syntax.HighlightYCStartupsInHeadlines(title, syntax.FaintAndItalic)
		title = syntax.HighlightYearInHeadlines(title, syntax.FaintAndItalic)
		title = syntax.HighlightHackerNewsHeadlines(title, syntax.FaintAndItalic)
		title = syntax.HighlightSpecialContent(title)

		title = faint + italic + title + " " + domain
		desc = s.NormalDesc.Render(desc)

	case m.disableInput && !(m.onAddToFavoritesPrompt || m.onRemoveFromFavoritesPrompt):
		title = syntax.HighlightYCStartupsInHeadlines(title, syntax.FaintAndItalic)
		title = syntax.HighlightYearInHeadlines(title, syntax.FaintAndItalic)
		title = syntax.HighlightHackerNewsHeadlines(title, syntax.FaintAndItalic)
		title = syntax.HighlightSpecialContent(title)

		title = faint + italic + title + " " + domain
		desc = s.NormalDesc.Render(desc)

	default:
		title = syntax.HighlightYCStartupsInHeadlines(title, syntax.Normal)
		title = syntax.HighlightYearInHeadlines(title, syntax.Normal)
		title = syntax.HighlightHackerNewsHeadlines(title, syntax.Normal)
		title = syntax.HighlightSpecialContent(title)

		title = s.NormalTitle.Render(title) + " " + domain
		desc = s.NormalDesc.Render(desc)
	}

	if d.ShowDescription {
		_, _ = fmt.Fprintf(w, "%s\n%s", title, desc)
		return
	}
	_, _ = fmt.Fprintf(w, "%s", title)
}

func parseTime(unixTime int64) string {
	moment, _ := goment.Unix(unixTime)
	now, _ := goment.New()

	return moment.From(now)
}

// ShortHelp returns the delegate's short help.
func (d DefaultDelegate) ShortHelp() []key.Binding {
	if d.ShortHelpFunc != nil {
		return d.ShortHelpFunc()
	}
	return nil
}

// FullHelp returns the delegate's full help.
func (d DefaultDelegate) FullHelp() [][]key.Binding {
	if d.FullHelpFunc != nil {
		return d.FullHelpFunc()
	}
	return nil
}

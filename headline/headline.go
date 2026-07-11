// Package headline styles story titles for the list rows and pane headers:
// YC-batch labels, years, Ask/Show/Tell prefixes, special-content tags and
// domains, each aware of the row's selection state.
package headline

import (
	"image/color"
	"regexp"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

type HighlightType int

const (
	noBreakSpace = "\u00a0"
	ansiBlack    = 16 // ANSI 256-color black
)

const (
	Unselected HighlightType = iota
	HeadlineInCommentSection
	Selected
	// OpenStory is the muted reading marker: bright-black background instead
	// of Selected's reverse video. Every highlighter restores the row's base
	// style after a token via getHighlight, so the two must stay in sync.
	OpenStory
	MarkAsRead
	AddToFavorites
	RemoveFromFavorites
)

var (
	reYCWithSeason    = regexp.MustCompile(`\((YC ([SWFXP]\d{2}))\)`)
	reYCWithoutSeason = regexp.MustCompile(`\((YC [SWFXP]\d{2})\)`)
	reYear            = regexp.MustCompile(`\((\d{4})\)`)
)

func HighlightYCStartupsInHeadlines(comment string, highlightType HighlightType, enableNerdFonts bool) string {
	if enableNerdFonts {
		highlightedStartup := ansi.Reset + getYCBarNerdFonts(nerdfonts.YCombinator+noBreakSpace+`$2`, highlightType) +
			getHighlight(highlightType)

		return reYCWithSeason.ReplaceAllString(comment, highlightedStartup)
	}

	highlightedStartup := ansi.Reset + highlightWithColor(`$1`, style.HeadlineYCLabelColor(), highlightType) +
		getHighlight(highlightType)

	return reYCWithoutSeason.ReplaceAllString(comment, highlightedStartup)
}

func highlightWithColor(text string, c color.Color, highlightType HighlightType) string {
	s := lipgloss.NewStyle().Foreground(c)

	switch highlightType {
	case Selected:
		s = s.Reverse(true)
	case OpenStory:
		s = s.Background(lipgloss.BrightBlack)
	case MarkAsRead:
		s = s.Faint(true)
	case Unselected, HeadlineInCommentSection, AddToFavorites, RemoveFromFavorites:
	}

	return s.Render(text)
}

func getYCBarNerdFonts(text string, highlightType HighlightType) string {
	c := style.HeadlineYCLabelColor()
	black := lipgloss.ANSIColor(ansiBlack)

	if highlightType == Selected {
		return label(text, c, black, highlightType)
	}

	return label(text, black, c, highlightType)
}

func HighlightYear(comment string, highlightType HighlightType) string {
	content := highlightWithColor(`$1`, style.HeadlineYearColor(), highlightType)

	return reYear.ReplaceAllString(comment, ansi.Reset+content+getHighlight(highlightType))
}

func label(text string, fg color.Color, bg color.Color, highlightType HighlightType) string {
	content := lipgloss.NewStyle().
		Foreground(fg).
		Background(bg)

	if highlightType == MarkAsRead {
		content = content.Italic(true).Faint(true)
	}

	if highlightType == HeadlineInCommentSection {
		content = content.Bold(true)
	}

	return ansi.Reset +
		getLeftBorder(bg, highlightType) +
		content.Render(text) +
		getRightBorder(bg, highlightType)
}

func getLeftBorder(bg color.Color, highlightType HighlightType) string {
	return borderStyle(bg, highlightType).Render(nerdfonts.LeftSeparator)
}

func getRightBorder(bg color.Color, highlightType HighlightType) string {
	return borderStyle(bg, highlightType).Render(nerdfonts.RightSeparator)
}

func borderStyle(bg color.Color, highlightType HighlightType) lipgloss.Style {
	if highlightType == Selected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.NoColor{}).
			Background(bg).
			Reverse(true)
	}

	if highlightType == OpenStory {
		return lipgloss.NewStyle().
			Foreground(bg).
			Background(lipgloss.BrightBlack)
	}

	return lipgloss.NewStyle().
		Foreground(bg)
}

func HighlightHackerNewsHeadlines(title string, highlightType HighlightType) string {
	askHN := "Ask HN:"
	showHN := "Show HN:"
	tellHN := "Tell HN:"
	thankHN := "Thank HN:"
	launchHN := "Launch HN:"

	highlight := getHighlight(highlightType)

	title = strings.ReplaceAll(title, askHN, style.HeadlineAskHN(askHN)+highlight)
	title = strings.ReplaceAll(title, showHN, style.HeadlineShowHN(showHN)+highlight)
	title = strings.ReplaceAll(title, tellHN, style.HeadlineTellHN(tellHN)+highlight)
	title = strings.ReplaceAll(title, thankHN, style.HeadlineThankHN(thankHN)+highlight)
	title = strings.ReplaceAll(title, launchHN, style.HeadlineLaunchHN(launchHN)+highlight)

	return title
}

func getHighlight(highlightType HighlightType) string {
	switch highlightType {
	case HeadlineInCommentSection:
		return ansi.Bold
	case Selected:
		return ansi.Reverse
	case OpenStory:
		return ansi.BgBrightBlack
	case MarkAsRead:
		return ansi.Faint + ansi.Italic
	case AddToFavorites:
		return ansi.Green + ansi.Reverse
	case RemoveFromFavorites:
		return ansi.Red + ansi.Reverse
	case Unselected:
		return ""
	}

	return ""
}

// ReplaceSpecialContentTags substitutes [video], [audio], [pdf], [PDF] with
// their compact nerdfont icons. Call this BEFORE truncation so the shorter
// icons are accounted for in width calculations.
func ReplaceSpecialContentTags(title string, enableNerdFonts bool) string {
	if !enableNerdFonts {
		return title
	}

	title = strings.ReplaceAll(title, "[audio]", nerdfonts.Audio)
	title = strings.ReplaceAll(title, "[video]", nerdfonts.Video)
	title = strings.ReplaceAll(title, "[pdf]", nerdfonts.Document)
	title = strings.ReplaceAll(title, "[PDF]", nerdfonts.Document)

	return title
}

func HighlightSpecialContent(title string, highlightType HighlightType, enableNerdFonts bool) string {
	highlight := getHighlight(highlightType)

	if enableNerdFonts {
		title = strings.ReplaceAll(title, nerdfonts.Audio, style.HeadlineAudio(nerdfonts.Audio)+highlight)
		title = strings.ReplaceAll(title, nerdfonts.Video, style.HeadlineVideo(nerdfonts.Video)+highlight)
		title = strings.ReplaceAll(title, nerdfonts.Document, style.HeadlinePDF(nerdfonts.Document)+highlight)

		return title
	}

	title = strings.ReplaceAll(title, "[audio]", style.HeadlineAudio("audio")+highlight)
	title = strings.ReplaceAll(title, "[video]", style.HeadlineVideo("video")+highlight)
	title = strings.ReplaceAll(title, "[pdf]", style.HeadlinePDF("pdf")+highlight)
	title = strings.ReplaceAll(title, "[PDF]", style.HeadlinePDF("PDF")+highlight)

	return title
}

func HighlightDomain(domain string) string {
	if domain == "" {
		return ansi.Reset
	}

	return ansi.Reset + style.Faint("("+domain+")")
}

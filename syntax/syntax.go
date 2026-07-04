package syntax

import (
	"html"
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
	newParagraph = "\n\n"
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
	reUnwantedNewLine = regexp.MustCompile(`([\w\W[:cntrl:]])(\n)([a-zA-Z0-9" \-<[:cntrl:]…])`)
	reHTMLAnchor      = regexp.MustCompile(`<a href=".*?"[^>]*>`)
	reAnchorWithURL   = regexp.MustCompile(`<a href="([^"]*)"[^>]*>https?://[^,"\) \n]+`)
	reURL             = regexp.MustCompile(`https?://([^,"\) \n]+)`)
	reMention         = regexp.MustCompile(`((?:^| )\B@[\w.]+)`)
	reVariable        = regexp.MustCompile(`(\$+[a-zA-Z_\-]+)`)
	reDoubleDash      = regexp.MustCompile(`([a-zA-Z])--([a-zA-Z])`)
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

var smileys = []struct{ from, to string }{
	{`:)`, "😊"},
	{`(:`, "😊"},
	{`:-)`, "😊"},
	{`:D`, "😄"},
	{`=)`, "😃"},
	{`=D`, "😃"},
	{`;)`, "😉"},
	{`;-)`, "😉"},
	{`:P`, "😜"},
	{`;P`, "😜"},
	{`:o`, "😮"},
	{`:O`, "😮"},
	{`:(`, "😔"},
	{`:-(`, "😔"},
	{`:/`, "😕"},
	{`:-/`, "😕"},
	{`-_-`, "😑"},
	{`:|`, "😐"},
}

func ConvertSmileys(text string) string {
	for _, s := range smileys {
		text = replaceBetweenWhitespace(text, s.from, s.to)
	}

	return text
}

func replaceBetweenWhitespace(text string, target string, replacement string) string {
	if text == target {
		return replacement
	}

	return strings.ReplaceAll(text, " "+target, " "+replacement)
}

func RemoveUnwantedNewLines(text string) string {
	return reUnwantedNewLine.ReplaceAllString(text, `$1`+" "+`$3`)
}

func RemoveUnwantedWhitespace(text string) string {
	singleSpace := " "
	doubleSpace := "  "
	tripleSpace := "   "

	text = strings.ReplaceAll(text, tripleSpace, singleSpace)
	text = strings.ReplaceAll(text, doubleSpace, singleSpace)

	return text
}

func HighlightDomain(domain string) string {
	if domain == "" {
		return ansi.Reset
	}

	return ansi.Reset + style.Faint("("+domain+")")
}

func HighlightReferences(input string) string {
	input = strings.ReplaceAll(input, "[0]", "["+style.White("0")+"]")
	input = strings.ReplaceAll(input, "[1]", "["+style.Red("1")+"]")
	input = strings.ReplaceAll(input, "[2]", "["+style.Yellow("2")+"]")
	input = strings.ReplaceAll(input, "[3]", "["+style.Green("3")+"]")
	input = strings.ReplaceAll(input, "[4]", "["+style.Blue("4")+"]")
	input = strings.ReplaceAll(input, "[5]", "["+style.Cyan("5")+"]")
	input = strings.ReplaceAll(input, "[6]", "["+style.Magenta("6")+"]")
	input = strings.ReplaceAll(input, "[7]", "["+style.BrightWhite("7")+"]")
	input = strings.ReplaceAll(input, "[8]", "["+style.BrightRed("8")+"]")
	input = strings.ReplaceAll(input, "[9]", "["+style.BrightYellow("9")+"]")
	input = strings.ReplaceAll(input, "[10]", "["+style.BrightGreen("10")+"]")

	return input
}

func ColorizeIndentSymbol(indentSymbol string, level int) string {
	if level == 0 {
		return ansi.Reset
	}

	cycle := style.IndentCycle()
	idx := (level - 1) % len(cycle)

	return ansi.Reset + cycle[idx](indentSymbol)
}

const maxURLDisplay = 50

func TrimURLs(comment string, highlightURLs bool) string {
	// Replace anchor-wrapped URLs with the full URL from href so that
	// HN-truncated display text is restored to the complete URL.
	comment = reAnchorWithURL.ReplaceAllString(comment, "$1")

	comment = reHTMLAnchor.ReplaceAllString(comment, "")

	if !highlightURLs {
		return comment
	}

	// Process all URLs in a single pass: scheme-stripped, truncated
	// display text with an OSC 8 hyperlink pointing to the full URL.
	comment = reURL.ReplaceAllStringFunc(comment, func(match string) string {
		display := truncateURL(reURL.FindStringSubmatch(match)[1])

		return style.CommentURL(display, match)
	})

	comment = strings.ReplaceAll(comment, "."+ansi.Reset+" ", ansi.Reset+". ")

	return comment
}

func truncateURL(display string) string {
	runes := []rune(display)
	if len(runes) <= maxURLDisplay {
		return display
	}

	return string(runes[:maxURLDisplay]) + "…"
}

func HighlightBackticks(input string) string {
	numberOfBackticks := strings.Count(input, "`")
	if numberOfBackticks == 0 || numberOfBackticks%2 != 0 {
		return input
	}

	parts := strings.Split(input, "`")

	var result strings.Builder

	for i, part := range parts {
		if i%2 == 1 {
			result.WriteString(style.CommentBacktick(part))
		} else {
			result.WriteString(part)
		}
	}

	return result.String()
}

func HighlightMentions(input string) string {
	input = reMention.ReplaceAllString(input, style.CommentMention(`$1`))

	input = strings.ReplaceAll(input, style.CommentMention("@dang"),
		style.CommentMod("@dang"))
	input = strings.ReplaceAll(input, style.CommentMention(" @dang"),
		style.CommentMod(" @dang"))

	return input
}

func HighlightVariables(input string) string {
	// Highlighting variables inside commands marked with backticks
	// messes with the formatting. If there are both backticks and variables
	// in the comment, we give priority to the backticks.
	numberOfBackticks := strings.Count(input, "`")
	if numberOfBackticks > 0 {
		return input
	}

	return reVariable.ReplaceAllString(input, style.CommentVariable(`$1`))
}

func HighlightAbbreviations(input string) string {
	iAmNotALawyer := "IANAL"
	iAmALawyer := "IAAL"

	input = strings.ReplaceAll(input, iAmNotALawyer, style.Red(iAmNotALawyer))
	input = strings.ReplaceAll(input, iAmALawyer, style.Green(iAmALawyer))

	return input
}

func ReplaceCharacters(input string) string {
	return html.UnescapeString(input)
}

func ReplaceHTML(input string) string {
	input = strings.TrimPrefix(input, "<p>")

	input = strings.ReplaceAll(input, "<p>", newParagraph)
	input = strings.ReplaceAll(input, "<i>", ansi.Italic)
	input = strings.ReplaceAll(input, "</i>", ansi.Reset)
	input = strings.ReplaceAll(input, "</a>", "")
	input = strings.ReplaceAll(input, "<pre><code>", "")
	input = strings.ReplaceAll(input, "</code></pre>", "")

	return input
}

func ReplaceSymbols(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, "...", "…")
	paragraph = strings.ReplaceAll(paragraph, "CO2", "CO₂")

	paragraph = replaceDoubleDashesWithEmDash(paragraph)
	paragraph = convertFractions(paragraph)

	return paragraph
}

func replaceDoubleDashesWithEmDash(paragraph string) string {
	paragraph = strings.ReplaceAll(paragraph, " -- ", " — ")

	return reDoubleDash.ReplaceAllString(paragraph, `$1`+"—"+`$2`)
}

// The narrow ⅒ glyph gets a trailing space to preserve alignment.
var fractionReplacer = strings.NewReplacer(
	" 1/2", " ½", "1/2 ", "½ ",
	" 1/3", " ⅓", "1/3 ", "⅓ ",
	" 2/3", " ⅔", "2/3 ", "⅔ ",
	" 1/4", " ¼", "1/4 ", "¼ ",
	" 3/4", " ¾", "3/4 ", "¾ ",
	" 1/5", " ⅕", "1/5 ", "⅕ ", "1/5th", "⅕th",
	" 2/5", " ⅖", "2/5 ", "⅖ ",
	" 3/5", " ⅗", "3/5 ", "⅗ ",
	" 4/5", " ⅘", "4/5 ", "⅘ ",
	" 1/6", " ⅙", "1/6 ", "⅙ ", "1/6th", "⅙th",
	" 1/10", " ⅒ ", "1/10 ", "⅒  ", "1/10th", "⅒ th",
)

func convertFractions(text string) string {
	return fractionReplacer.Replace(text)
}
